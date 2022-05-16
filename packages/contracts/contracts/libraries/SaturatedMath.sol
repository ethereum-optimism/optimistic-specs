//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

library SaturatedMath {
    function add(uint256 a, uint256 b) internal pure returns (uint256) {
        unchecked {
            uint256 c = a + b;
            if (c < a) {
                return type(uint256).max;
            }
            return c;
        }
    }

    function sub(uint256 a, uint256 b) internal pure returns (uint256) {
        unchecked {
            if (b > a) {
                return 0;
            }
            return a - b;
        }
    }

    function mul(uint256 a, uint256 b) internal pure returns (uint256) {
        unchecked {
            if (a == 0) {
                return 0;
            }
            uint256 c = a * b;
            if (c / a != b) {
                return type(uint256).max;
            }
            return c;
        }
    }

    function div(uint256 a, uint256 b) internal pure returns (uint256) {
        unchecked {
            if (b == 0) {
                return 0;
            }
            return a / b;
        }
    }

    function add_u128(uint128 a, uint128 b) internal pure returns (uint128) {
        unchecked {
            uint128 c = a + b;
            if (c < a) {
                return type(uint128).max;
            }
            return c;
        }
    }

    function sub_u128(uint128 a, uint128 b) internal pure returns (uint128) {
        unchecked {
            if (b > a) {
                return 0;
            }
            return a - b;
        }
    }

    function mul_u128(uint128 a, uint128 b) internal pure returns (uint128) {
        unchecked {
            if (a == 0) {
                return 0;
            }
            uint128 c = a * b;
            if (c / a != b) {
                return type(uint128).max;
            }
            return c;
        }
    }

    function add_u64(uint64 a, uint64 b) internal pure returns (uint64) {
        unchecked {
            uint64 c = a + b;
            if (c < a) {
                return type(uint64).max;
            }
            return c;
        }
    }
}
