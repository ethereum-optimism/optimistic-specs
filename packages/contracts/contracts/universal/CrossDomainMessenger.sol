// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

// solhint-disable max-line-length

/* Library Imports */
import {
    Lib_DefaultValues
} from "@eth-optimism/contracts/libraries/constants/Lib_DefaultValues.sol";
import {
    Lib_CrossDomainUtils
} from "@eth-optimism/contracts/libraries/bridge/Lib_CrossDomainUtils.sol";
import { CrossDomainHashing } from "../libraries/Lib_CrossDomainHashing.sol";

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
 * @title CrossDomainMessenger
 * @dev The CrossDomainMessenger contract delivers messages between two layers.
 */
abstract contract CrossDomainMessenger is
    OwnableUpgradeable,
    PausableUpgradeable,
    ReentrancyGuardUpgradeable
{
    /**********
     * Events *
     **********/

    event SentMessage(
        address indexed target,
        address sender,
        bytes message,
        uint256 messageNonce,
        uint256 gasLimit
    );

    event RelayedMessage(
        bytes32 indexed msgHash
    );

    event FailedRelayedMessage(
        bytes32 indexed msgHash
    );

    /*************
     * Constants *
     *************/

    uint16 constant HASH_VERSION = 1;

    /*************
     * Variables *
     *************/

    // blockedMessages in old L1CrossDomainMessenger
    bytes32 internal REMOVED_VARIABLE_SPACER_1;

    // relayedMessages in old L1CrossDomainMessenger
    bytes32 internal REMOVED_VARIABLE_SPACER_2;

    /// @notice Mapping of message hash to boolean success value.
    mapping(bytes32 => bool) public successfulMessages;

    /// @notice Current x-domain message sender.
    address internal xDomainMsgSender;

    /// @notice Nonce for the next message to be sent.
    uint256 internal msgNonce;

    /// @notice Address of the CrossDomainMessenger on the other chain.
    address public otherMessenger;

    /********************
     * Public Functions *
     ********************/

    /**
     * Pause relaying.
     */
    function pause() external onlyOwner {
        _pause();
    }

    /**
     * Retrieves the address of the x-domain message sender. Will throw an error if the sender is
     * not currently set (equal to the default sender).
     *
     * @return Address of the x-domain message sender.
     */
    function xDomainMessageSender() external view returns (address) {
        require(
            xDomainMsgSender != Lib_DefaultValues.DEFAULT_XDOMAIN_SENDER,
            "xDomainMessageSender is not set"
        );

        return xDomainMsgSender;
    }

    /**
     * Retrieves the next message nonce. Adds the hash version to the nonce.
     *
     * @return Next message nonce with added hash version.
     */
    function messageNonce() public view returns (uint256) {
        return CrossDomainHashing.addVersionToNonce(
            msgNonce,
            HASH_VERSION
        );
    }

    /**
     *
     * @param _target Target contract address.
     * @param _message Message to send to the target.
     * @param _gasLimit Gas limit for the provided message.
     */
    function sendMessage(
        address _target,
        bytes memory _message,
        uint32 _gasLimit
    ) external payable {
        _sendMessage(
            otherMessenger,
            address(this),
            msg.value,
            uint64(_gasLimit),
            false,
            Lib_CrossDomainUtils.encodeXDomainCalldata(
                _target,
                msg.sender,
                _message,
                messageNonce()
            )
        );

        emit SentMessage(_target, msg.sender, _message, messageNonce(), _gasLimit);

        unchecked {
            ++msgNonce;
        }
    }

    function relayMessage(
        uint256 _nonce,
        address _sender,
        address _target,
        uint256 _value,
        uint256 _gasLimit,
        bytes calldata _data,
        bytes calldata _proof
    ) external nonReentrant whenNotPaused payable {
        bytes32 versionedHash = CrossDomainHashing.getVersionedHash(
            _nonce,
            _sender,
            _target,
            _value,
            _gasLimit,
            _data
        );

        require(
            _target != address(this),
            "Cannot send messages to self."
        );

        require(
            successfulMessages[versionedHash] == false,
            "Message has already been relayed."
        );

        require(
            _verifyMessageProof(
                versionedHash,
                _nonce,
                _sender,
                _target,
                _value,
                _gasLimit,
                _data,
                _proof
            ),
            "Message could not be authenticated."
        );

        xDomainMsgSender = _sender;
        (bool success, ) = _target.call{ value: _value, gas: _gasLimit }(_data);
        xDomainMsgSender = Lib_DefaultValues.DEFAULT_XDOMAIN_SENDER;

        if (success == true) {
            successfulMessages[versionedHash] = true;
            emit RelayedMessage(versionedHash);
        } else {
            emit FailedRelayedMessage(versionedHash);
        }
    }

    /**********************
     * Internal Functions *
     **********************/

    /**
     * Initializes the contract.
     */
    function _initialize(
        address _otherMessenger
    )
        internal
        initializer
    {
        xDomainMsgSender = Lib_DefaultValues.DEFAULT_XDOMAIN_SENDER;
        otherMessenger = _otherMessenger;

        // TODO: ensure we know what these are doing and why they are here
        // Initialize upgradable OZ contracts
        __Context_init_unchained();
        __Ownable_init_unchained();
        __Pausable_init_unchained();
        __ReentrancyGuard_init_unchained();
    }

    function _verifyMessageProof(
        bytes32 _versionedHash,
        uint256 _nonce,
        address _sender,
        address _target,
        uint256 _value,
        uint256 _gasLimit,
        bytes calldata _data,
        bytes calldata _proof
    ) internal view virtual returns (bool);

    function _sendMessage(
        address _to,
        address _from,
        uint256 _value,
        uint64 _gasLimit,
        bool _isCreation,
        bytes memory _data
    ) internal virtual;
}
