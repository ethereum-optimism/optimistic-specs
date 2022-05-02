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

    // TODO: test withdrawing ETH
    // with L2Bridge.withdraw

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

