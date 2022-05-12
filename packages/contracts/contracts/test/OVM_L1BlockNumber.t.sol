//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import { Test } from "forge-std/Test.sol";
import { L1Block } from "../L2/L1Block.sol";
import { OVM_L1BlockNumber } from "../L2/OVM_L1BlockNumber.sol";

contract OVM_L1BlockNumberTest is Test {
    L1Block lb;
    OVM_L1BlockNumber bn;
    bytes32 immutable NON_ZERO_HASH = keccak256(abi.encode(1));

    function setUp() external {
        lb = new L1Block();
        bn = new OVM_L1BlockNumber();
        vm.prank(lb.DEPOSITOR_ACCOUNT());
        lb.setL1BlockValues(uint64(999), uint64(2), 3, NON_ZERO_HASH, uint64(4));
        bn.initialize(address(lb));
    }

    function test_initializeCalledOnce() external {
        vm.expectRevert(abi.encodeWithSignature("AlreadyInitialized()"));
        bn.initialize(address(1));
    }

    function test_getL1BlockNumber() external {
        assertEq(bn.getL1BlockNumber(), 999);
    }
}
