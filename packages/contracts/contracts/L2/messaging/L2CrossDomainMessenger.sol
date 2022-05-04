// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

// solhint-disable max-line-length
/* Library Imports */
import { AddressAliasHelper } from "@eth-optimism/contracts/standards/AddressAliasHelper.sol";
import {
    Lib_CrossDomainUtils
} from "@eth-optimism/contracts/libraries/bridge/Lib_CrossDomainUtils.sol";
import {
    Lib_DefaultValues
} from "@eth-optimism/contracts/libraries/constants/Lib_DefaultValues.sol";
import { Lib_BedrockPredeployAddresses } from "../../libraries/Lib_BedrockPredeployAddresses.sol";

/* Interface Imports */
import { IL2CrossDomainMessenger } from "./IL2CrossDomainMessenger.sol";

import { Withdrawer } from "../Withdrawer.sol";
import { WithdrawalVerifier } from "../../libraries/Lib_WithdrawalVerifier.sol";

// solhint-enable max-line-length

/**
 * @title L2CrossDomainMessenger
 * @dev The L2 Cross Domain Messenger contract sends messages from L2 to L1, and is the entry point
 * for L2 messages sent via the L1 Cross Domain Messenger.
 *
 */
contract L2CrossDomainMessenger is IL2CrossDomainMessenger {
    /*************
     * Variables *
     *************/

    mapping(bytes32 => bool) public relayedMessages;
    mapping(bytes32 => bool) public successfulMessages;
    mapping(bytes32 => bool) public sentMessages;
    uint256 public messageNonce;
    address internal xDomainMsgSender = Lib_DefaultValues.DEFAULT_XDOMAIN_SENDER;
    address public l1CrossDomainMessenger;

    uint16 constant HASH_VERSION = 1;

    /***************
     * Constructor *
     ***************/

    constructor(address _l1CrossDomainMessenger) {
        l1CrossDomainMessenger = _l1CrossDomainMessenger;
    }

    /********************
     * Public Functions *
     ********************/

    function xDomainMessageSender() external view returns (address) {
        require(
            xDomainMsgSender != Lib_DefaultValues.DEFAULT_XDOMAIN_SENDER,
            "xDomainMessageSender is not set"
        );
        return xDomainMsgSender;
    }

    /**
     * Sends a cross domain message to the target messenger.
     * @param _target Target contract address.
     * @param _message Message to send to the target.
     * @param _gasLimit Gas limit for the provided message.
     */
    function sendMessage(
        address _target,
        bytes memory _message,
        uint32 _gasLimit
    ) external payable {
        uint256 nonce = WithdrawalVerifier.addVersionToNonce(
            messageNonce,
            HASH_VERSION
        );

        bytes32 versionedHash = WithdrawalVerifier.getVersionedHash(
            nonce,
            msg.sender,
            _target,
            msg.value,
            _gasLimit,
            _message
        );

        sentMessages[versionedHash] = true;

        // Emit an event before we bump the nonce or the nonce will be off by one.
        emit SentMessage(_target, msg.sender, _message, nonce, _gasLimit);
        unchecked {
            ++messageNonce;
        }

        // Actually send the message.
        Withdrawer(Lib_BedrockPredeployAddresses.WITHDRAWER).initiateWithdrawal(
            l1CrossDomainMessenger,
            _gasLimit,
            _message
        );
    }

    /**
     * Relays a cross domain message to a contract.
     * @inheritdoc IL2CrossDomainMessenger
     */
    function relayMessage(
        address _target,
        address _sender,
        bytes memory _message,
        uint256 _messageNonce
    ) external payable {
        // Since it is impossible to deploy a contract to an address on L2 which matches
        // the alias of the L1CrossDomainMessenger, this check can only pass when it is called in
        // the first call frame of a deposit transaction. Thus reentrancy is prevented here.
        require(
            AddressAliasHelper.undoL1ToL2Alias(msg.sender) == l1CrossDomainMessenger,
            "Provided message could not be verified."
        );

        // the gasLimit is set to 0
        bytes32 versionedHash = WithdrawalVerifier.getVersionedHash(
            _messageNonce,
            _sender,
            _target,
            msg.value,
            0,
            _message
        );

        require(
            successfulMessages[versionedHash] == false,
            "Provided message has already been received."
        );

        // Prevent calls to WITHDRAWER, which would enable
        // an attacker to maliciously craft the _message to spoof
        // a call from any L2 account.
        // Todo: evaluate if this attack is still relevant
        if (_target == Lib_BedrockPredeployAddresses.WITHDRAWER) {
            // Write to the successfulMessages mapping and return immediately.
            successfulMessages[versionedHash] = true;
            return;
        }

        xDomainMsgSender = _sender;
        // slither-disable-next-line reentrancy-no-eth, reentrancy-events, reentrancy-benign
        (bool success, ) = _target.call{ value: msg.value }(_message);
        // slither-disable-next-line reentrancy-benign
        xDomainMsgSender = Lib_DefaultValues.DEFAULT_XDOMAIN_SENDER;

        // Mark the message as received if the call was successful. Ensures that a message can be
        // relayed multiple times in the case that the call reverted.
        if (success == true) {
            // slither-disable-next-line reentrancy-no-eth
            successfulMessages[versionedHash] = true;
            // slither-disable-next-line reentrancy-events
            emit RelayedMessage(versionedHash);
        } else {
            // slither-disable-next-line reentrancy-events
            emit FailedRelayedMessage(versionedHash);
        }
    }
}
