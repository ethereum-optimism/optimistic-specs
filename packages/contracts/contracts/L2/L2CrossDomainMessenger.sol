// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

// solhint-disable max-line-length
/* Library Imports */
import { AddressAliasHelper } from "@eth-optimism/contracts/standards/AddressAliasHelper.sol";
import {
    Lib_PredeployAddresses
} from "@eth-optimism/contracts/libraries/constants/Lib_PredeployAddresses.sol";

/* Interaction imports */
import { CrossDomainMessenger } from "../universal/CrossDomainMessenger.sol";
import { L2ToL1MessagePasser } from "./L2ToL1MessagePasser.sol";

// solhint-enable max-line-length

/**
 * @title L2CrossDomainMessenger
 * @notice The L2CrossDomainMessenger contract facilitates sending both ETH value and data from L2
 * to L1. It is predeployed in the L2 state at address 0x4200000000000000000000000000000000000016.
 */
contract L2CrossDomainMessenger is CrossDomainMessenger {
    /********************
     * Public Functions *
     ********************/

    function initialize(address _l1CrossDomainMessenger) external initializer {
        address[] memory blockedSystemAddresses = new address[](2);
        blockedSystemAddresses[0] = address(this);
        blockedSystemAddresses[1] = Lib_PredeployAddresses.L2_TO_L1_MESSAGE_PASSER;

        _initialize(_l1CrossDomainMessenger, blockedSystemAddresses);
    }

    /**********************
     * Internal Functions *
     **********************/

    function _isSystemMessageSender() internal view override returns (bool) {
        return AddressAliasHelper.undoL1ToL2Alias(msg.sender) == otherMessenger;
    }

    function _sendMessage(
        address _to,
        uint64 _gasLimit,
        uint256 _value,
        bytes memory _data
    ) internal override {
        L2ToL1MessagePasser(payable(Lib_PredeployAddresses.L2_TO_L1_MESSAGE_PASSER))
            .initiateWithdrawal{ value: _value }(_to, _gasLimit, _data);
    }
}
