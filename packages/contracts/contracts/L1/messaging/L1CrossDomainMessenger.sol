// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

// solhint-disable max-line-length
/* Library Imports */
import { AddressAliasHelper } from "@eth-optimism/contracts/standards/AddressAliasHelper.sol";
import { Lib_OVMCodec } from "@eth-optimism/contracts/libraries/codec/Lib_OVMCodec.sol";
import {
    Lib_DefaultValues
} from "@eth-optimism/contracts/libraries/constants/Lib_DefaultValues.sol";
import {
    Lib_PredeployAddresses
} from "@eth-optimism/contracts/libraries/constants/Lib_PredeployAddresses.sol";
import {
    Lib_CrossDomainUtils
} from "@eth-optimism/contracts/libraries/bridge/Lib_CrossDomainUtils.sol";

/* Interface Imports */
import { IL1CrossDomainMessenger } from "./IL1CrossDomainMessenger.sol";
import { OptimismPortal } from "../OptimismPortal.sol";
import { WithdrawalVerifier } from "../../libraries/Lib_WithdrawalVerifier.sol";
import { L2OutputOracle } from "../L2OutputOracle.sol";

import { AddressAliasHelper } from "@eth-optimism/contracts/standards/AddressAliasHelper.sol";

/* External Imports */
import {
    OwnableUpgradeable
} from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import {
    PausableUpgradeable
} from "@openzeppelin/contracts-upgradeable/security/PausableUpgradeable.sol";
import {
    ReentrancyGuardUpgradeable
} from "@openzeppelin/contracts-upgradeable/security/ReentrancyGuardUpgradeable.sol";

// solhint-enable max-line-length

/**
 * @title L1CrossDomainMessenger
 * @dev The L1 Cross Domain Messenger contract sends messages from L1 to L2, and relays messages
 * from L2 onto L1.
 */
contract L1CrossDomainMessenger is
    IL1CrossDomainMessenger,
    OwnableUpgradeable,
    PausableUpgradeable,
    ReentrancyGuardUpgradeable
{
    uint16 constant HASH_VERSION = 1;

    /// @notice Error emitted when attempting to finalize a withdrawal too early.
    error NotYetFinal();

    /// @notice Error emitted when the output root proof is invalid.
    error InvalidOutputRootProof();

    /// @notice Error emitted when the withdrawal inclusion proof is invalid.
    error InvalidWithdrawalInclusionProof();

    /// @notice Error emitted when a withdrawal has already been finalized.
    error WithdrawalAlreadyFinalized();

    /**********************
     * Contract Variables *
     **********************/

    // TODO: blockedMessages is no longer in use
    mapping(bytes32 => bool) public blockedMessages;
    // TODO: removing this, can update to using it later
    mapping(bytes32 => bool) public relayedMessages;
    mapping(bytes32 => bool) public successfulMessages;

    // This must be set in the initialize function
    address internal xDomainMsgSender;

    // Bedrock upgrade note: the nonce must be initialized to greater than the last value of
    // CanonicalTransactionChain.queueElements.length. Otherwise it will be possible to have
    // messages which cannot be relayed on L2.
    uint256 public messageNonce;

    // TODO: ideally these are immutable but cannot be immutable
    // because this is behind a proxy

    /// @notice Address of the L2OutputOracle.
    L2OutputOracle public L2_ORACLE;
    /// @notice Minimum time that must elapse before a withdrawal can be finalized.
    uint256 public FINALIZATION_PERIOD;

    /********************
     * Public Functions *
     ********************/

    function initialize(L2OutputOracle _l2Oracle, uint256 _finalizationPeriod)
        external
        initializer
    {
        xDomainMsgSender = Lib_DefaultValues.DEFAULT_XDOMAIN_SENDER;
        L2_ORACLE = _l2Oracle;
        FINALIZATION_PERIOD = _finalizationPeriod;

        // TODO: ensure we know what these are doing and why they are here
        // Initialize upgradable OZ contracts
        __Context_init_unchained(); // Context is a dependency for both Ownable and Pausable
        __Ownable_init_unchained();
        __Pausable_init_unchained();
        __ReentrancyGuard_init_unchained();
    }

    /**
     * Pause relaying.
     */
    function pause() external onlyOwner {
        _pause();
    }

    /**
     * Get the xDomainMessageSender
     */
    function xDomainMessageSender() external view returns (address) {
        require(
            xDomainMsgSender != Lib_DefaultValues.DEFAULT_XDOMAIN_SENDER,
            "xDomainMessageSender is not set"
        );
        return xDomainMsgSender;
    }

    /**
     * Sends a cross domain message to the target messenger.
     * This function faciliates L1 to L2 communication.
     * @param _target Target contract address.
     * @param _message Message to send to the target.
     * @param _gasLimit Gas limit for the provided message.
     */
    function sendMessage(
        address _target,
        bytes memory _message,
        uint32 _gasLimit
    ) external payable {
        _sendMessageRaw(
            Lib_PredeployAddresses.L2_CROSS_DOMAIN_MESSENGER,
            address(this),
            msg.value,
            uint64(_gasLimit),
            false,
            Lib_CrossDomainUtils.encodeXDomainCalldata(
                _target,
                msg.sender,
                _message,
                WithdrawalVerifier.addVersionToNonce(
                    messageNonce,
                    HASH_VERSION
                )
            )
        );

        emit SentMessage(_target, msg.sender, _message, messageNonce, _gasLimit);

        unchecked {
            ++messageNonce;
        }
    }

    function sendMessageRaw(
        address _to,
        uint256 _value,
        uint64 _gasLimit,
        bool _isCreation,
        bytes memory _data
    ) external payable {
        _sendMessageRaw(
            _to,
            msg.sender,
            _value,
            _gasLimit,
            _isCreation,
            _data
        );
    }

    /**
     * Relays a cross domain message to a contract.
     * This function faciliates L2 to L1 communication.
     * Calls WithdrawalsRelay.finalizeWithdrawalTransaction
     * @inheritdoc IL1CrossDomainMessenger
     */
    function relayMessage(
        uint256 _nonce,
        address _sender,
        address _target,
        uint256 _value,
        uint256 _gasLimit,
        bytes calldata _data,
        uint256 _l2Timestamp,
        WithdrawalVerifier.OutputRootProof calldata _outputRootProof,
        bytes calldata _withdrawalProof
    ) external nonReentrant whenNotPaused payable {
        // Check that the timestamp is sufficiently finalized.
        // The timestamp corresponds to a particular L2 output,
        // so it is safe to be passed in by a user.

        // Get the output root.
        bytes32 outputRoot = WithdrawalVerifier._deriveOutputRoot(_outputRootProof);
        L2OutputOracle.OutputProposal memory outputProposal = L2_ORACLE.getL2Output(_l2Timestamp);
        require(outputProposal.outputRoot == outputRoot);

        // Sequencer can steal funds if they wait 1 week before submitting
        // the L2 output commitment
        unchecked {
            if (block.timestamp < outputProposal.timestamp + FINALIZATION_PERIOD) {
                revert NotYetFinal();
            }
        }

        bytes32 versionedHash = WithdrawalVerifier.getVersionedHash(
            _nonce,
            _sender,
            _target,
            _value,
            _gasLimit,
            _data
        );

        if (
            WithdrawalVerifier._verifyWithdrawalInclusion(
                versionedHash,
                _outputRootProof.withdrawerStorageRoot,
                _withdrawalProof
            ) == false
        ) {
            revert InvalidWithdrawalInclusionProof();
        }

        require(successfulMessages[versionedHash] == false);

        require(
            _target != address(this),
            "Cannot send L2->L1 messages to L1 system contracts."
        );

        xDomainMsgSender = _sender;
        // Make the call.
        (bool success, ) = _target.call{ value: _value, gas: _gasLimit }(_data);
        xDomainMsgSender = Lib_DefaultValues.DEFAULT_XDOMAIN_SENDER;

        // Mark the message as received if the call was successful. Ensures that a message can be
        // relayed multiple times in the case that the call reverted.
        if (success == true) {
            // slither-disable-next-line reentrancy-no-eth
            successfulMessages[versionedHash] = true;
            // slither-disable-next-line reentrancy-events
            emit RelayedMessage(versionedHash);
            // TODO:
            // relayedMessages is no longer used because it was not originally
            // secure in the first place
        } else {
            // slither-disable-next-line reentrancy-events
            emit FailedRelayedMessage(versionedHash);
        }
    }

    // TODO: internal functions
    function _sendMessageRaw(
        address _to,
        address _from,
        uint256 _value,
        uint64 _gasLimit,
        bool _isCreation,
        bytes memory _data
    ) internal {
        require(!_isCreation || _to == address(0), "");

        // Transform the from-address to its alias if the caller is a contract.
        if (_from != tx.origin) {
            _from = AddressAliasHelper.applyL1ToL2Alias(msg.sender);
        }
        // emit TransactionDeposited which causes the message to actually
        // end up in L2
        emit TransactionDeposited(_from, _to, msg.value, _value, _gasLimit, _isCreation, _data);
    }

}
