//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import { DSTest } from "forge-std/Test.sol";
import { Vm } from "forge-std/Vm.sol";
import { L1Block } from "../L2/L1Block.sol";
import { OVM_L1BlockNumber } from "../L2/predeploys/OVM_L1BlockNumber.sol";

contract OVM_L1BlockNumberTest is DSTest {
    Vm vm = Vm(HEVM_ADDRESS);
    L1Block lb;
    OVM_L1BlockNumber bn;
    bytes32 immutable NON_ZERO_HASH = keccak256(abi.encode(1));

    function setUp() external {
        lb = new L1Block();
        bn = new OVM_L1BlockNumber(address(lb));
        vm.prank(lb.DEPOSITOR_ACCOUNT());
        lb.setL1BlockValues(uint64(999), uint64(2), 3, NON_ZERO_HASH, uint64(4));
    }

    function test_getL1BlockNumber() external {
        assertEq(bn.getL1BlockNumber(), 999);
    }
}
