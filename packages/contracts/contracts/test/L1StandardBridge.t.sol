//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import { CommonTest } from "./CommonTest.sol";
import { L2OutputOracle_Initializer } from "./L2OutputOracle.t.sol";

import { L2StandardBridge } from "../L2/messaging/L2StandardBridge.sol";
import { L1StandardBridge } from "../L1/messaging/L1StandardBridge.sol";
import { OptimismPortal } from "../L1/OptimismPortal.sol";

contract L1StandardBridge_Test is CommonTest, L2OutputOracle_Initializer {
    OptimismPortal op;

    L1StandardBridge L1Bridge ;
    L2StandardBridge L2Bridge ;

    function setUp() external {
        L1Bridge = new L1StandardBridge();
        L2Bridge = new L2StandardBridge(address(L1Bridge));
        op = new OptimismPortal(oracle, 100);

        L1Bridge.initialize(op, address(L2Bridge));
    }

    function test_L1BridgeSetsPortalAndL2Bridge() external {
        OptimismPortal portal = L1Bridge.optimismPortal();
        address bridge = L1Bridge.l2TokenBridge();

        assertEq(address(portal), address(op));
        assertEq(bridge, address(L2Bridge));
    }

    // receive
    // - can accept ETH
    // depositETH
    // - emits ETHDepositInitiated
    // - calls optimismPortal.depositTransaction
    // - only EOA
    // - ETH ends up in the optimismPortal
    // depositETHTo
    // - emits ETHDepositInitiated
    // - calls optimismPortal.depositTransaction
    // - EOA or contract can call
    // - ETH ends up in the optimismPortal
    // depositERC20
    // - updates bridge.deposits
    // - emits ERC20DepositInitiated
    // - calls optimismPortal.depositTransaction
    // - only callable by EOA
    // depositERC20To
    // - updates bridge.deposits
    // - emits ERC20DepositInitiated
    // - calls optimismPortal.depositTransaction
    // - reverts if called by EOA
    // - callable by a contract
    // finalizeETHWithdrawal
    // - emits ETHWithdrawalFinalized
    // - only callable by L2 bridge
    // finalizeERC20Withdrawal
    // - updates bridge.deposits
    // - emits ERC20WithdrawalFinalized
    // - only callable by L2 bridge
    // donateETH
    // - can send ETH to the contract
}
