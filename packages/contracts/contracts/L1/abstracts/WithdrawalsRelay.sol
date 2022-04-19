//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

/* Interactions Imports */
import { L2OutputOracle } from "../L2OutputOracle.sol";

/* Library Imports */
import { WithdrawalVerifier } from "../libraries/Lib_WithdrawalVerifier.sol";

/**
 * @title WithdrawalsRelay
 * @notice The WithdrawalsRelay is inherited by the OptimismPortal on L1, and faciliates finalizing
 * withdrawals between L2 and L1.
 */
abstract contract WithdrawalsRelay {
    /**********
     * Errors *
     **********/

    /// @notice Error emitted when attempting to finalize a withdrawal too early.
    error NotYetFinal();

    /// @notice Error emitted when the output root proof is invalid.
    error InvalidOutputRootProof();

    /// @notice Error emitted when the withdrawal inclusion proof is invalid.
    error InvalidWithdrawalInclusionProof();

    /// @notice Error emitted when a withdrawal has already been finalized.
    error WithdrawalAlreadyFinalized();

    /**********
     * Events *
     **********/

    /// @notice Emitted when a withdrawal is finalized
    event WithdrawalFinalized(bytes32 indexed);

    /// @notice Value used to reset the l2Sender, this is more efficient than setting it to zero.
    address internal constant DEFAULT_L2_SENDER = 0x000000000000000000000000000000000000dEaD;

    /**********************
     * Contract Variables *
     **********************/

    /// @notice Minimum time that must elapse before a withdrawal can be finalized.
    uint256 public immutable FINALIZATION_PERIOD;

    /// @notice Address of the L2OutputOracle.
    L2OutputOracle public immutable L2_ORACLE;

    /**
     * @notice Public variable which can be used to read the address of the L2 account which
     * initated the withdrawal. Can also be used to determine whether or not execution is occuring
     * downstream of a call to finalizeWithdrawalTransaction().
     */
    address public l2Sender = DEFAULT_L2_SENDER;

    /**
     * @notice A list of withdrawal hashes which have been successfully finalized.
     * Used for replay protection.
     */
    mapping(bytes32 => bool) public finalizedWithdrawals;

    /***************
     * Constructor *
     ***************/

    constructor(L2OutputOracle _l2Oracle, uint256 _finalizationPeriod) {
        L2_ORACLE = _l2Oracle;
        FINALIZATION_PERIOD = _finalizationPeriod;
    }

    /**********************
     * External Functions *
     **********************/

    /**
     * @notice Finalizes a withdrawal transaction.
     * @param _nonce Nonce for the provided message.
     * @param _sender Message sender address on L2.
     * @param _target Target address on L1.
     * @param _value ETH to send to the target.
     * @param _gasLimit Gas to be forwarded to the target.
     * @param _data Data to send to the target.
     * @param _timestamp L2 timestamp of the outputRoot.
     * @param _outputRootProof Inclusion proof of the withdrawer contracts storage root.
     * @param _withdrawalProof Inclusion proof for the given withdrawal in the withdrawer contract.
     */
    function finalizeWithdrawalTransaction(
        uint256 _nonce,
        address _sender,
        address _target,
        uint256 _value,
        uint256 _gasLimit,
        bytes calldata _data,
        uint256 _timestamp,
        WithdrawalVerifier.OutputRootProof calldata _outputRootProof,
        bytes calldata _withdrawalProof
    ) external {
        // Check that the timestamp is sufficiently finalized.
        unchecked {
            if (block.timestamp < _timestamp + FINALIZATION_PERIOD) {
                revert NotYetFinal();
            }
        }

        // Get the output root.
        bytes32 outputRoot = L2_ORACLE.getL2Output(_timestamp);

        // Verify that the output root can be generated with the elements in the proof.
        if (outputRoot != WithdrawalVerifier._deriveOutputRoot(_outputRootProof)) {
            revert InvalidOutputRootProof();
        }

        // Verify that the hash of the withdrawal transaction's arguments are included in the
        // storage hash of the withdrawer contract.
        bytes32 withdrawalHash = WithdrawalVerifier._deriveWithdrawalHash(
            _nonce,
            _sender,
            _target,
            _value,
            _gasLimit,
            _data
        );
        if (
            WithdrawalVerifier._verifyWithdrawalInclusion(
                withdrawalHash,
                _outputRootProof.withdrawerStorageRoot,
                _withdrawalProof
            ) == false
        ) {
            revert InvalidWithdrawalInclusionProof();
        }

        // Check that this withdrawal has not already been finalized.
        if (finalizedWithdrawals[withdrawalHash] == true) {
            revert WithdrawalAlreadyFinalized();
        }
        finalizedWithdrawals[withdrawalHash] = true;

        l2Sender = _sender;
        // Make the call.
        (bool s, ) = _target.call{ value: _value, gas: _gasLimit }(_data);
        s; // Silence the compiler's "Return value of low-level calls not used" warning.
        l2Sender = DEFAULT_L2_SENDER;

        // All withdrawals are immediately finalized. If the ability to replay a transaction is
        // required, that support can be provided in external contracts.
        emit WithdrawalFinalized(withdrawalHash);
    }
}
