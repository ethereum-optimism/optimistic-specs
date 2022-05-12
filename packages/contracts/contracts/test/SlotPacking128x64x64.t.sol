//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import { Test } from "forge-std/Test.sol";
import { SlotPacking128x64x64 } from "../libraries/SlotPacking128x64x64.sol";

contract SlotPacking128x64x64_Test is Test {
    function test_set(uint128 _a, uint64 _b, uint64 _c) external {
        bytes32 filled = SlotPacking128x64x64.set(_a, _b, _c);
        (uint128 a, uint64 b, uint64 c) = SlotPacking128x64x64.get(filled);
        assertEq(a, _a);
        assertEq(b, _b);
        assertEq(c, _c);
    }
}
