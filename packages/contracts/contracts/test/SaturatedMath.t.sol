//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import { Test } from "forge-std/Test.sol";
import { SaturatedMath } from "../libraries/SaturatedMath.sol";

contract SaturatedMath_Test is Test {
    function test_add(uint256 a, uint256 b) external {
        uint256 result = SaturatedMath.add(a, b);

        unchecked {
            uint256 c = a + b;
            if (c != result) {
                assertEq(result, type(uint256).max);
            } else {
                assertEq(c, result);
            }
        }
    }

    function test_sub(uint256 a, uint256 b) external {
        uint256 result = SaturatedMath.sub(a, b);

        unchecked {
            uint256 c = a - b;
            if (c != result) {
                assertEq(result, 0);
            } else {
                assertEq(c, result);
            }
        }
    }

    function test_mul(uint256 a, uint256 b) external {
        uint256 result = SaturatedMath.mul(a, b);

        unchecked {
            uint256 c = a * b;
            if (a == 0) {
                assertEq(result, 0);
            } else if (c / a != b) {
                assertEq(result, type(uint256).max);
            } else {
                assertEq(result, c);
            }
        }
    }

    function test_div(uint256 a, uint256 b) external {
        uint256 result = SaturatedMath.div(a, b);

        unchecked {
            if (b == 0) {
                assertEq(result, 0);
            } else {
                uint256 c = a / b;
                assertEq(result, c);
            }
        }
    }
}
