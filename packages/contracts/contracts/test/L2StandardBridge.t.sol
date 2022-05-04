//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import {
    Lib_PredeployAddresses
} from "@eth-optimism/contracts/libraries/constants/Lib_PredeployAddresses.sol";

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
import { IL1ERC20Bridge } from "../L1/messaging/IL1ERC20Bridge.sol";
import { AddressAliasHelper } from "@eth-optimism/contracts/standards/AddressAliasHelper.sol";

import { console } from "forge-std/console.sol";
import { stdStorage, StdStorage } from "forge-std/Test.sol";

contract L2StandardBridge_Test is CommonTest, BridgeInitializer  {
    using stdStorage for StdStorage;

    function setUp() external {
        // put some tokens in the bridge, give them to alice on L2
        uint256 slot = stdstore
            .target(address(L1Bridge))
            .sig("deposits(address,address)")
            .with_key(address(token))
            .with_key(address(L2Token))
            .find();

        vm.store(address(L1Bridge), bytes32(slot), bytes32(uint256(100000)));
        deal(address(L2Token), alice, 100000, true);
    }

    function test_L2BridgeCorrectL1Bridge() external {
        address l1Bridge = L2Bridge.l1TokenBridge();
        assertEq(address(L1Bridge), l1Bridge);
    }

    // withdraw
    // - token is burned
    // - emits WithdrawalInitiated
    // - calls Withdrawer.initiateWithdrawal
    function test_L2BridgeWithdraw() external {
        vm.expectEmit(true, true, true, true);
        emit WithdrawalInitiated(
            address(token),
            address(L2Token),
            alice,
            alice,
            100,
            hex""
        );

        uint256 aliceBalance = L2Token.balanceOf(alice);

        vm.expectCall(
            Lib_BedrockPredeployAddresses.WITHDRAWER,
            abi.encodeWithSelector(
                IWithdrawer.initiateWithdrawal.selector,
                address(L1Bridge),
                10000,
                abi.encodeWithSelector(
                    IL1ERC20Bridge.finalizeERC20Withdrawal.selector,
                    address(token),
                    address(L2Token),
                    alice,
                    alice,
                    100,
                    hex""
                )
            )
        );

        vm.expectCall(
            address(L2Token),
            abi.encodeWithSelector(
                IL2StandardERC20.burn.selector,
                alice,
                100
            )
        );

        vm.prank(address(alice));
        L2Bridge.withdraw(
            address(L2Token),
            100,
            10000,
            hex""
        );

        assertEq(L2Token.balanceOf(alice), aliceBalance - 100);
        assertEq(L2Token.totalSupply(), L2Token.balanceOf(alice));
    }

    function test_L2BridgeRevertWithdraw() external {
        vm.expectRevert(abi.encodeWithSignature("InvalidWithdrawalAmount()"));
        vm.prank(address(alice));
        L2Bridge.withdraw{ value: 100 }(
            address(L2Token),
            100,
            10000,
            hex""
        );
    }

    function test_L2BridgeWithdrawETH() external {
        vm.expectCall(
            Lib_BedrockPredeployAddresses.WITHDRAWER,
            abi.encodeWithSelector(
                Withdrawer.initiateWithdrawal.selector,
                address(L1Bridge),
                10000,
                abi.encodeWithSelector(
                    L1StandardBridge.finalizeETHWithdrawal.selector,
                    alice,
                    alice,
                    100,
                    hex""
                )
            )
        );

        vm.expectEmit(true, true, true, true);
        emit WithdrawalInitiated(
            address(0),
            Lib_PredeployAddresses.OVM_ETH,
            alice,
            alice,
            100,
            hex""
        );

        uint256 aliceBalance = alice.balance;

        vm.prank(address(alice));
        L2Bridge.withdraw{ value: 100 }(
            Lib_PredeployAddresses.OVM_ETH,
            100,
            10000,
            hex""
        );

        uint256 aliceBalancePost = alice.balance;

        // Alice's balance should go down
        assertEq(aliceBalance - 100, aliceBalancePost);
    }

    function test_L2BridgeRevertWithdrawETH() external {
        vm.expectRevert(abi.encodeWithSignature("InvalidWithdrawalAmount()"));
        vm.prank(address(alice));
        L2Bridge.withdraw(
            Lib_PredeployAddresses.OVM_ETH,
            100,
            10000,
            hex""
        );
    }

    // withdrawTo
    // - token is burned
    // - emits WithdrawalInitiated w/ correct recipient
    // - calls Withdrawer.initiateWithdrawal
    function test_L2BridgeWithdrawTo() external {
        assertEq(L2Token.balanceOf(bob), 0);

        uint256 aliceBalance = L2Token.balanceOf(alice);

        vm.expectCall(
            Lib_BedrockPredeployAddresses.WITHDRAWER,
            abi.encodeWithSelector(
                Withdrawer.initiateWithdrawal.selector,
                address(L1Bridge),
                200000,
                abi.encodeWithSelector(
                    IL1ERC20Bridge.finalizeERC20Withdrawal.selector,
                    address(token),
                    address(L2Token),
                    alice,
                    bob,
                    200,
                    hex""
                )
            )
        );

        vm.expectEmit(true, true, true, true);
        emit WithdrawalInitiated(
            address(token),
            address(L2Token),
            alice,
            bob,
            200,
            hex""
        );

        vm.prank(alice);
        L2Bridge.withdrawTo(
            address(L2Token),
            bob,
            200,
            200000,
            hex""
        );

        // alice's balance should go down
        uint256 aliceBalancePost = L2Token.balanceOf(alice);
        assertEq(aliceBalance - 200, aliceBalancePost);
    }

    // TODO: the eth functions

    // finalizeDeposit
    // - only callable by l1TokenBridge
    // - supported token pair emits DepositFinalized
    // - invalid deposit emits DepositFailed
    // - invalid deposit calls Withdrawer.initiateWithdrawal
    function test_L2BridgeFinalizeDeposit() external {
        uint256 aliceBalance = L2Token.balanceOf(alice);

        vm.expectCall(
            address(L2Token),
            abi.encodeWithSelector(
                IL2StandardERC20.mint.selector,
                alice,
                100
            )
        );

        vm.expectEmit(true, true, true, true);
        emit DepositFinalized(
            address(token),
            address(L2Token),
            alice,
            alice,
            100,
            hex""
        );

        vm.prank(AddressAliasHelper.applyL1ToL2Alias(address(L1Bridge)));
        L2Bridge.finalizeDeposit(
            address(token),
            address(L2Token),
            alice,
            alice,
            100,
            hex""
        );

        uint256 aliceBalancePost = L2Token.balanceOf(alice);
        assertEq(aliceBalance + 100, aliceBalancePost);
    }

    function test_L2BridgeBadDeposit() external {
        vm.expectEmit(true, true, true, true);
        emit DepositFailed(
            address(10),
            address(L2Token),
            alice,
            alice,
            100,
            hex""
        );

        vm.prank(AddressAliasHelper.applyL1ToL2Alias(address(L1Bridge)));
        L2Bridge.finalizeDeposit(
            address(10),
            address(L2Token),
            alice,
            alice,
            100,
            hex""
        );
    }

    function test_L2BridgeFinalizeETHDeposit() external {
        uint256 aliceBalance = alice.balance;

        vm.expectEmit(true, true, true, true);
        emit DepositFinalized(
            address(0),
            Lib_PredeployAddresses.OVM_ETH,
            alice,
            alice,
            100,
            hex""
        );

        hoax(AddressAliasHelper.applyL1ToL2Alias(address(L1Bridge)));
        L2Bridge.finalizeDeposit{ value: 100 }(
            address(0),
            Lib_PredeployAddresses.OVM_ETH,
            alice,
            alice,
            100,
            hex""
        );

        uint256 aliceBalancePost = alice.balance;
        assertEq(aliceBalance + 100, aliceBalancePost);
    }

    // when the values do not match up
    function test_L2BridgeFinalizeETHDepositWrongAmount() external {
        uint256 aliceBalance = alice.balance;

        vm.expectEmit(true, true, true, true);
        emit DepositFailed(
            address(0),
            Lib_PredeployAddresses.OVM_ETH,
            alice,
            alice,
            100,
            hex""
        );

        vm.expectCall(
            Lib_BedrockPredeployAddresses.WITHDRAWER,
            abi.encodeWithSelector(
                Withdrawer.initiateWithdrawal.selector,
                address(L1Bridge),
                0,
                abi.encodeWithSelector(
                    L1StandardBridge.finalizeETHWithdrawal.selector,
                    alice,
                    alice,
                    200,
                    hex""
                )
            )
        );

        hoax(AddressAliasHelper.applyL1ToL2Alias(address(L1Bridge)));
        L2Bridge.finalizeDeposit{ value: 200 }(
            address(0),
            Lib_PredeployAddresses.OVM_ETH,
            alice,
            alice,
            100,
            hex""
        );

        uint256 aliceBalancePost = alice.balance;
        assertEq(aliceBalance, aliceBalancePost);
        assertEq(address(L2Bridge).balance, 0);
        assertEq(address(Lib_BedrockPredeployAddresses.WITHDRAWER).balance, 200);
    }

    function test_L2BridgeFinalizeDepositRevertsOnCaller() external {
        vm.expectRevert("Can only be called by a the l1TokenBridge");
        vm.prank(alice);
        L2Bridge.finalizeDeposit(
            address(0),
            Lib_PredeployAddresses.OVM_ETH,
            alice,
            alice,
            100,
            hex""
        );
    }
}

