//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import {
    Lib_PredeployAddresses
} from "@eth-optimism/contracts/libraries/constants/Lib_PredeployAddresses.sol";
import { IL2ERC20Bridge } from "@eth-optimism/contracts/L2/messaging/IL2ERC20Bridge.sol";

import { IWithdrawer } from "../L2/IWithdrawer.sol";
import { Withdrawer } from "../L2/Withdrawer.sol";
import { L2StandardBridge } from "../L2/messaging/L2StandardBridge.sol";
import { L1StandardBridge } from "../L1/messaging/L1StandardBridge.sol";
import { OptimismPortal } from "../L1/OptimismPortal.sol";
import { Lib_BedrockPredeployAddresses } from "../libraries/Lib_BedrockPredeployAddresses.sol";
import { L2StandardTokenFactory } from "../L2/messaging/L2StandardTokenFactory.sol";
import { IL2StandardTokenFactory } from "../L2/messaging/IL2StandardTokenFactory.sol";
import { L2StandardERC20 } from "../L2/tokens/L2StandardERC20.sol";
import { IL2StandardERC20 } from "../L2/tokens/IL2StandardERC20.sol";

import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { CommonTest } from "./CommonTest.t.sol";
import { L2OutputOracle_Initializer, BridgeInitializer } from "./L2OutputOracle.t.sol";
import { LibRLP } from "./Lib_RLP.t.sol";

import { stdStorage, StdStorage } from "forge-std/Test.sol";

contract L1StandardBridge_Test is CommonTest, BridgeInitializer  {
    using stdStorage for StdStorage;

    function test_L1BridgeSetsPortalAndL2Bridge() external {
        OptimismPortal portal = L1Bridge.optimismPortal();
        address bridge = L1Bridge.l2TokenBridge();

        assertEq(address(portal), address(op));
        assertEq(bridge, address(L2Bridge));
    }

    // receive
    // - can accept ETH
    function test_L1BridgeReceiveETH() external {
        vm.expectEmit(true, true, true, true);
        emit ETHDepositInitiated(alice, alice, 100, hex"");

        vm.expectCall(
            address(op),
            abi.encodeWithSelector(
                op.depositTransaction.selector,
                Lib_PredeployAddresses.L2_STANDARD_BRIDGE,
                100,
                200000,
                false,
                abi.encodeWithSelector(
                    IL2ERC20Bridge.finalizeDeposit.selector,
                    address(0),
                    Lib_PredeployAddresses.OVM_ETH,
                    alice,
                    alice,
                    100,
                    hex""
                )
            )
        );

        vm.prank(alice);
        (bool success, bytes memory data) = address(L1Bridge).call{ value: 100 }(hex"");
        assertEq(success, true);
        assertEq(data, hex"");
        assertEq(address(op).balance, 100);
    }

    // depositETH
    // - emits ETHDepositInitiated
    // - calls optimismPortal.depositTransaction
    // - only EOA
    // - ETH ends up in the optimismPortal

    // TODO: this now goes through the bridge
    function test_L1BridgeDepositETH() external {
        vm.expectEmit(true, true, true, true);
        emit ETHDepositInitiated(alice, alice, 1000, hex"ff");

        vm.expectCall(
            address(op),
            abi.encodeWithSelector(
                op.depositTransaction.selector,
                Lib_PredeployAddresses.L2_STANDARD_BRIDGE,
                1000,
                10000,
                false,
                abi.encodeWithSelector(
                    IL2ERC20Bridge.finalizeDeposit.selector,
                    address(0),
                    Lib_PredeployAddresses.OVM_ETH,
                    alice,
                    alice,
                    1000,
                    hex"ff"
                )
            )
        );

        vm.prank(alice);
        L1Bridge.depositETH{ value: 1000 }(10000, hex"ff");
        assertEq(address(op).balance, 1000);
    }

    function test_L1BridgeOnlyEOADepositETH() external {
        vm.etch(alice, address(token).code);

        vm.expectRevert("Account not EOA");
        vm.prank(alice);
        L1Bridge.depositETH{ value: 1000 }(10000, hex"ff");
        assertEq(address(op).balance, 0);
    }

    // depositETHTo
    // - emits ETHDepositInitiated
    // - calls optimismPortal.depositTransaction
    // - EOA or contract can call
    // - ETH ends up in the optimismPortal
    function test_L1BridgeDepositETHTo() external {
        vm.expectEmit(true, true, true, true);
        emit ETHDepositInitiated(alice, bob, 1000, hex"ff");

        vm.expectCall(
            address(op),
            abi.encodeWithSelector(
                op.depositTransaction.selector,
                Lib_PredeployAddresses.L2_STANDARD_BRIDGE,
                1000,
                10000,
                false,
                abi.encodeWithSelector(
                    IL2ERC20Bridge.finalizeDeposit.selector,
                    address(0),
                    Lib_PredeployAddresses.OVM_ETH,
                    alice,
                    bob,
                    1000,
                    hex"ff"
                )
            )
        );

        vm.prank(alice);
        L1Bridge.depositETHTo{ value: 1000 }(bob, 10000, hex"ff");
        assertEq(address(op).balance, 1000);
    }

    // depositERC20
    // - updates bridge.deposits
    // - emits ERC20DepositInitiated
    // - calls optimismPortal.depositTransaction
    // - only callable by EOA
    function test_L1BridgeDepositERC20() external {
        vm.expectEmit(true, true, true, true);
        emit ERC20DepositInitiated(
            address(token),
            address(L2Token),
            alice,
            alice,
            100,
            hex""
        );

        deal(address(token), alice, 100000, true);

        vm.prank(alice);
        token.approve(address(L1Bridge), type(uint256).max);

        vm.expectCall(
            address(token),
            abi.encodeWithSelector(
                ERC20.transferFrom.selector,
                alice,
                address(L1Bridge),
                100
            )
        );

        vm.expectCall(
            address(op),
            abi.encodeWithSelector(
                op.depositTransaction.selector,
                Lib_PredeployAddresses.L2_STANDARD_BRIDGE,
                0,
                10000,
                false,
                abi.encodeWithSelector(
                    IL2ERC20Bridge.finalizeDeposit.selector,
                    address(token),
                    address(L2Token),
                    alice,
                    alice,
                    100,
                    hex""
                )
            )
        );

        vm.prank(alice);
        L1Bridge.depositERC20(
            address(token),
            address(L2Token),
            100,
            10000,
            hex""
        );

        assertEq(L1Bridge.deposits(address(token), address(L2Token)), 100);
    }

    function test_L1BridgeOnlyEOADepositERC20() external {
        vm.etch(alice, address(token).code);

        vm.expectRevert("Account not EOA");
        vm.prank(alice);
        L1Bridge.depositERC20(
            address(token),
            address(L2Token),
            100,
            10000,
            hex""
        );
        assertEq(L1Bridge.deposits(address(token), address(L2Token)), 0);
    }

    // depositERC20To
    // - updates bridge.deposits
    // - emits ERC20DepositInitiated
    // - calls optimismPortal.depositTransaction
    // - callable by a contract
    function test_L1BridgeDepositERC20To() external {
        vm.expectEmit(true, true, true, true);
        emit ERC20DepositInitiated(
            address(token),
            address(L2Token),
            alice,
            bob,
            1000,
            hex""
        );

        deal(address(token), alice, 100000, true);

        vm.prank(alice);
        token.approve(address(L1Bridge), type(uint256).max);

        vm.expectCall(
            address(token),
            abi.encodeWithSelector(
                ERC20.transferFrom.selector,
                alice,
                address(L1Bridge),
                1000
            )
        );

        vm.expectCall(
            address(op),
            abi.encodeWithSelector(
                op.depositTransaction.selector,
                Lib_PredeployAddresses.L2_STANDARD_BRIDGE,
                0,
                10000,
                false,
                abi.encodeWithSelector(
                    IL2ERC20Bridge.finalizeDeposit.selector,
                    address(token),
                    address(L2Token),
                    alice,
                    bob,
                    1000,
                    hex""
                )
            )
        );

        vm.prank(alice);
        L1Bridge.depositERC20To(
            address(token),
            address(L2Token),
            bob,
            1000,
            10000,
            hex""
        );

        assertEq(L1Bridge.deposits(address(token), address(L2Token)), 1000);
    }

    // finalizeETHWithdrawal
    // - emits ETHWithdrawalFinalized
    // - only callable by L2 bridge
    function test_L1BridgeFinalizeETHWithdrawal() external {
        vm.deal(address(op), 100);
        vm.store(address(op), 0, bytes32(abi.encode(L2Bridge)));

        vm.expectEmit(true, true, true, true);
        emit ETHWithdrawalFinalized(
            alice,
            alice,
            100,
            hex""
        );

        vm.expectCall(
            alice,
            hex""
        );

        vm.prank(address(op));
        L1Bridge.finalizeETHWithdrawal{ value: 100 }(
            alice,
            alice,
            100,
            hex""
        );

        assertEq(address(op).balance, 0);
    }

    function test_L1BridgeOnlyPortalFinalizeETHWithdrawal() external {
        vm.expectRevert("Messages must be relayed by first calling the Optimism Portal");
        L1Bridge.finalizeETHWithdrawal{ value: 100 }(
            alice,
            alice,
            100,
            hex""
        );
    }

    function test_L1BridgeOnlyL2BridgeFinalizeETHWithdrawal() external {
        vm.deal(address(op), 100);

        vm.expectRevert("Message must be sent from the L2 Token Bridge");
        vm.prank(address(op));
        L1Bridge.finalizeETHWithdrawal{ value: 100 }(
            alice,
            alice,
            100,
            hex""
        );
    }

    // finalizeERC20Withdrawal
    // - updates bridge.deposits
    // - emits ERC20WithdrawalFinalized
    // - only callable by L2 bridge
    function test_L1BridgeFinalizeERC20Withdrawal() external {
        deal(address(token), address(L1Bridge), 100, true);

        uint256 slot = stdstore
            .target(address(L1Bridge))
            .sig("deposits(address,address)")
            .with_key(address(token))
            .with_key(address(L2Token))
            .find();

        // Give the L1 bridge some ERC20 tokens
        vm.store(address(L1Bridge), bytes32(slot), bytes32(uint256(100)));
        assertEq(L1Bridge.deposits(address(token), address(L2Token)), 100);

        vm.expectEmit(true, true, true, true);
        emit ERC20WithdrawalFinalized(
            address(token),
            address(L2Token),
            alice,
            alice,
            100,
            hex""
        );

        vm.expectCall(
            address(token),
            abi.encodeWithSelector(
                ERC20.transfer.selector,
                alice,
                100
            )
        );

        vm.store(address(op), 0, bytes32(abi.encode(L2Bridge)));
        vm.prank(address(op));
        L1Bridge.finalizeERC20Withdrawal(
            address(token),
            address(L2Token),
            alice,
            alice,
            100,
            hex""
        );

        assertEq(token.balanceOf(address(L1Bridge)), 0);
        assertEq(token.balanceOf(address(alice)), 100);
    }

    function test_L1BridgeOnlyPortalFinalizeERC20Withdrawal() external {
        vm.expectRevert("Messages must be relayed by first calling the Optimism Portal");
        L1Bridge.finalizeERC20Withdrawal(
            address(token),
            address(L2Token),
            alice,
            alice,
            100,
            hex""
        );
    }

    function test_L1BridgeOnlyL2BridgeFinalizeERC20Withdrawal() external {
        vm.expectRevert("Message must be sent from the L2 Token Bridge");
        vm.prank(address(op));
        L1Bridge.finalizeERC20Withdrawal(
            address(token),
            address(L2Token),
            alice,
            alice,
            100,
            hex""
        );
    }

    // donateETH
    // - can send ETH to the contract
    function test_L1BridgeDonateETH() external {
        assertEq(address(L1Bridge).balance, 0);
        vm.prank(alice);
        L1Bridge.donateETH{ value: 1000 }();
        assertEq(address(L1Bridge).balance, 1000);
    }
}
