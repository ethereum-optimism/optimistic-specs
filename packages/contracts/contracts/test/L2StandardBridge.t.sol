//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import { CommonTest } from "./CommonTest.sol";
import { L2OutputOracle_Initializer } from "./L2OutputOracle.t.sol";

import { L2StandardBridge } from "../L2/messaging/L2StandardBridge.sol";
import { L1StandardBridge } from "../L1/messaging/L1StandardBridge.sol";
import { OptimismPortal } from "../L1/OptimismPortal.sol";

contract L2StandardBridge_Test is CommonTest, L2OutputOracle_Initializer {
    OptimismPortal op;

    L1StandardBridge L1Bridge ;
    L2StandardBridge L2Bridge ;

    function setUp() external {
        L1Bridge = new L1StandardBridge();
        L2Bridge = new L2StandardBridge(address(L1Bridge));
        op = new OptimismPortal(oracle, 100);

        L1Bridge.initialize(op, address(L2Bridge));
    }

    function test_L2BridgeCorrectL1Bridge() external {
        address l1Bridge = L2Bridge.l1TokenBridge();
        assertEq(address(L1Bridge), l1Bridge);
    }

    // l1TokenBridge is correct
    // withdraw
    // - token is burned
    // - emits WithdrawalInitiated
    // - calls Withdrawer.initiateWithdrawal
    // withdrawTo
    // - token is burned
    // - emits WithdrawalInitiated w/ correct recipient
    // - calls Withdrawer.initiateWithdrawal
    // finalizeDeposit
    // - only callable by l1TokenBridge
    // - supported token pair emits DepositFinalized
    // - invalid deposit emits DepositFailed
    // - invalid deposit calls Withdrawer.initiateWithdrawal
}

