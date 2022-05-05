// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

// solhint-disable max-line-length
/* Library Imports */
import {
    Lib_PredeployAddresses
} from "@eth-optimism/contracts/libraries/constants/Lib_PredeployAddresses.sol";

/* Interface Imports */
import { OptimismPortal } from "./OptimismPortal.sol";
import { CrossDomainMessenger } from "../universal/CrossDomainMessenger.sol";

// solhint-enable max-line-length

/**
 * @title L1CrossDomainMessenger
 * @dev The L1 Cross Domain Messenger contract sends messages from L1 to L2, and relays messages
 * from L2 onto L1.
 */
contract L1CrossDomainMessenger is CrossDomainMessenger {
    /*************
     * Variables *
     *************/

    /// @notice Address of the OptimismPortal.
    OptimismPortal public portal;

    /********************
     * Public Functions *
     ********************/

    function initialize(OptimismPortal _portal) external initializer {
        portal = _portal;

        address[] memory blockedSystemAddresses = new address[](1);
        blockedSystemAddresses[0] = address(this);

        _initialize(Lib_PredeployAddresses.L2_CROSS_DOMAIN_MESSENGER, blockedSystemAddresses);
    }

    /**********************
     * Internal Functions *
     **********************/

    function _isSystemMessageSender() internal view override returns (bool) {
        return msg.sender == address(portal) && portal.l2Sender() == otherMessenger;
    }

    function _sendMessage(
        address _to,
        uint64 _gasLimit,
        uint256 _value,
        bytes memory _data
    ) internal override {
        portal.depositTransaction{ value: _value }(_to, _value, _gasLimit, false, _data);
    }
}
