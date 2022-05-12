//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import { Test } from "forge-std/Test.sol";
import { SlotPacking128x128 } from "../libraries/SlotPacking128x128.sol";

contract SlotPacking128x128_Test is Test {
    function test_set(uint128 _a, uint128 _b) external {
        bytes32 filled = SlotPacking128x128.set(_a, _b);
        (uint128 a, uint128 b) = SlotPacking128x128.get(filled);
        assertEq(a, _a);
        assertEq(b, _b);
    }
}
