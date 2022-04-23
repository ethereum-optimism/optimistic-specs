//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

/* Testing utilities */
import { CommonTest } from "./CommonTest.sol";
import { L2OutputOracle_Initializer } from "./L2OutputOracle.t.sol";

import {
    Lib_DefaultValues
} from "@eth-optimism/contracts/libraries/constants/Lib_DefaultValues.sol";
import {
    Lib_PredeployAddresses
} from "@eth-optimism/contracts/libraries/constants/Lib_PredeployAddresses.sol";
import {
    Lib_CrossDomainUtils
} from "@eth-optimism/contracts/libraries/bridge/Lib_CrossDomainUtils.sol";

/* Target contract dependencies */
import { L2OutputOracle } from "../L1/L2OutputOracle.sol";
import { OptimismPortal } from "../L1/OptimismPortal.sol";

/* Target contract */
import { L1CrossDomainMessenger } from "../L1/messaging/L1CrossDomainMessenger.sol";

contract L1CrossDomainMessenger_Test is CommonTest, L2OutputOracle_Initializer {
    // Dependencies
    OptimismPortal op;
    // 'L2OutputOracle oracle' is declared in L2OutputOracle_Initializer

    // Contract under test
    L1CrossDomainMessenger messenger;

    // Receiver address for testing
    address recipient = address(0xabbaacdc);

    function setUp() external {
        // Oracle value is zero, but this test does not depend on it.
        op = new OptimismPortal(oracle, 7 days);
        messenger = new L1CrossDomainMessenger();
        messenger.initialize(op);
    }

    // pause: should pause the contract when called by the current owner
    function test_pause() external {
        messenger.pause();
        assert(messenger.paused());
    }

    // pause: should not pause the contract when called by account other than the owner
    function testCannot_pause() external {
        vm.expectRevert('Ownable: caller is not the owner');
        vm.prank(address(0xABBA));
        messenger.pause();
    }

    // sendMessage: should be able to send a single message
    function test_sendMessage() external {
        uint256 messageNonce = messenger.messageNonce();
        bytes memory xDomainCalldata = Lib_CrossDomainUtils.encodeXDomainCalldata(
            recipient,
            address(this),
            NON_ZERO_DATA,
            messageNonce
        );
        vm.expectCall(
            address(op),
            abi.encodeWithSignature(
                "depositTransaction(address,uint256,uint256,bool,bytes)",
                Lib_PredeployAddresses.L2_CROSS_DOMAIN_MESSENGER, 0, NON_ZERO_GASLIMIT, false, xDomainCalldata
            )
        );
        messenger.sendMessage(recipient, NON_ZERO_DATA, uint32(NON_ZERO_GASLIMIT));
    }

    // sendMessage: should be able to send the same message twice
    function test_sendMessageTwice() external {
        messenger.sendMessage(recipient, NON_ZERO_DATA, uint32(NON_ZERO_GASLIMIT));
        messenger.sendMessage(recipient, NON_ZERO_DATA, uint32(NON_ZERO_GASLIMIT));
    }



// xDomainMessageSender: should return the xDomainMsgSender address


// relayMessage: should revert if still inside the fraud proof window
// relayMessage: should revert if attempting to relay a message sent to an L1 system contract
// relayMessage: should revert if provided an invalid output root proof
// relayMessage: should revert if provided an invalid storage trie witness
// relayMessage: should send a successful call to the target contract
// relayMessage: the xDomainMessageSender is reset to the original value
// relayMessage: should revert if trying to send the same message twice
// relayMessage: should revert if paused

// blockMessage and allowMessage: should revert if called by an account other than the owner
// blockMessage and allowMessage: should revert if the message is blocked
// blockMessage and allowMessage: should succeed if the message is blocked, then unblocked
