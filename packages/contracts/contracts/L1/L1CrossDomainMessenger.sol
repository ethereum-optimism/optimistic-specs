// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

// solhint-disable max-line-length
/* Library Imports */
import { AddressAliasHelper } from "@eth-optimism/contracts/standards/AddressAliasHelper.sol";
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
import { WithdrawalVerifier } from "../libraries/Lib_WithdrawalVerifier.sol";
import { CrossDomainHashing } from "../libraries/Lib_CrossDomainHashing.sol";
import { L2OutputOracle } from "./L2OutputOracle.sol";
import { CrossDomainMessenger } from "../universal/CrossDomainMessenger.sol";

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
    CrossDomainMessenger
{
    /**********
     * Events *
     **********/

    event TransactionDeposited(
        address indexed from,
        address indexed to,
        uint256 mint,
        uint256 value,
        uint64 gasLimit,
        bool isCreation,
        bytes data
    );

    /*************
     * Variables *
     *************/

    /// @notice Address of the L2OutputOracle.
    L2OutputOracle public l2OutputOracle;

    /// @notice Minimum time that must elapse before a withdrawal can be finalized.
    uint256 public finalizationPeriodSeconds;

    /********************
     * Public Functions *
     ********************/

    function initialize(
        L2OutputOracle _l2OutputOracle,
        uint256 _finalizationPeriodSeconds
    )
        external
        initializer
    {
        l2OutputOracle = _l2OutputOracle;
        finalizationPeriodSeconds = _finalizationPeriodSeconds;

        address[] memory blockedSystemAddresses = new address[](1);
        blockedSystemAddresses[0] = address(this);

        _initialize(
            Lib_PredeployAddresses.L2_CROSS_DOMAIN_MESSENGER,
            blockedSystemAddresses
        );
    }

    function sendMessageRaw(
        address _to,
        uint256 _value,
        uint64 _gasLimit,
        bool _isCreation,
        bytes memory _data
    ) external payable {
        _sendMessage(
            _to,
            msg.sender,
            _value,
            _gasLimit,
            _isCreation,
            _data
        );
    }

    /**********************
     * Internal Functions *
     **********************/

    function _verifyMessageProof(
        uint256 _nonce,
        address _sender,
        address _target,
        uint256 _value,
        uint256 _gasLimit,
        bytes calldata _data,
        bytes calldata _proof
    ) internal view override returns (bool) {
        (
            uint256 l2Timestamp,
            WithdrawalVerifier.OutputRootProof memory outputRootProof,
            bytes memory withdrawalProof
        ) = abi.decode(
            _proof,
            (
                uint256,
                WithdrawalVerifier.OutputRootProof,
                bytes
            )
        );

        L2OutputOracle.OutputProposal memory proposal = l2OutputOracle.getL2Output(
            l2Timestamp
        );

        require(
            proposal.outputRoot == WithdrawalVerifier._deriveOutputRoot(outputRootProof),
            "Proposed root does not match given root."
        );

        require(
            block.timestamp < proposal.timestamp + finalizationPeriodSeconds,
            "Proposal is not yet finalized."
        );

        require(
            WithdrawalVerifier._verifyWithdrawalInclusion(
                CrossDomainHashing.getVersionedHash(
                    _nonce,
                    _sender,
                    _target,
                    _value,
                    _gasLimit,
                    _data
                ),
                outputRootProof.withdrawerStorageRoot,
                withdrawalProof
            ),
            "Invalid withdrawal inclusion proof."
        );

        return true;
    }

    function _sendMessage(
        address _to,
        address _from,
        uint256 _value,
        uint64 _gasLimit,
        bool _isCreation,
        bytes memory _data
    ) internal override {
        require(
            !_isCreation || _to == address(0),
            "Contract creations must have the zero address as the target."
        );

        // Transform the from-address to its alias if the caller is a contract.
        if (_from != tx.origin) {
            _from = AddressAliasHelper.applyL1ToL2Alias(msg.sender);
        }

        emit TransactionDeposited(_from, _to, msg.value, _value, _gasLimit, _isCreation, _data);
    }
}
