// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

// solhint-disable max-line-length
/* Library Imports */
import { AddressAliasHelper } from "@eth-optimism/contracts/standards/AddressAliasHelper.sol";
import {
    Lib_DefaultValues
} from "@eth-optimism/contracts/libraries/constants/Lib_DefaultValues.sol";
import { CrossDomainHashing } from "../libraries/Lib_CrossDomainHashing.sol";

/* Interface Imports */
import { IL2CrossDomainMessenger } from "../interfaces/IL2CrossDomainMessenger.sol";

/* Interaction imports */
import { Burner } from "./Burner.sol";

// solhint-enable max-line-length

/**
 * @title L2CrossDomainMessenger
 * @notice The L2CrossDomainMessenger contract facilitates sending both
 * ETH value and data from L2 to L1.
 * It is predeployed in the L2 state at address TODO.
 */
contract L2CrossDomainMessenger is IL2CrossDomainMessenger {
    /*************
     * Variables *
     *************/

    // TODO: this event is newly defined as part of bedrock, do we
    // still need it?
    /**
     * @notice Emitted any time a withdrawal is initiated.
     * @param nonce Unique value corresponding to each withdrawal.
     * @param sender The L2 account address which initiated the withdrawal.
     * @param target The L1 account address the call will be send to.
     * @param value The ETH value submitted for withdrawal, to be forwarded to the target.
     * @param gasLimit The minimum amount of gas that must be provided when withdrawing on L1.
     * @param data The data to be forwarded to the target on L1.
     */
    event WithdrawalInitiated(
        uint256 indexed nonce,
        address indexed sender,
        address indexed target,
        uint256 value,
        uint256 gasLimit,
        bytes data
    );

    /// @notice Emitted when the balance of this contract is burned.
    event WithdrawerBalanceBurnt(uint256 indexed amount);

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
        uint256 nonce = CrossDomainHashing.addVersionToNonce(messageNonce, HASH_VERSION);

        bytes32 versionedHash = CrossDomainHashing.getVersionedHash(
            nonce,
            msg.sender,
            _target,
            msg.value,
            _gasLimit,
            _message
        );

        require(sentMessages[versionedHash] == false, "");
        sentMessages[versionedHash] = true;

        // Emit an event before we bump the nonce or the nonce will be off by one.
        emit SentMessage(_target, msg.sender, _message, nonce, _gasLimit);

        // TODO(tynes): I don't think we need this event anymore
        emit WithdrawalInitiated(nonce, msg.sender, _target, msg.value, _gasLimit, _message);
        unchecked {
            ++messageNonce;
        }
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
        bytes32 versionedHash = CrossDomainHashing.getVersionedHash(
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
        // TODO: this could be imported via a library instead of
        // address(this)
        if (_target == address(this)) {
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

    /**
     * @notice Removes all ETH held in this contract from the state, by deploying a contract which
     * immediately self destructs.
     * For simplicity, this call is not incentivized as it costs very little to run.
     * Inspired by https://etherscan.io/address/0xb69fba56b2e67e7dda61c8aa057886a8d1468575#code
     */
    function burn() external {
        uint256 balance = address(this).balance;
        new Burner{ value: balance }();
        emit WithdrawerBalanceBurnt(balance);
    }
}
