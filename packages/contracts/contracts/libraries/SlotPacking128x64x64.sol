//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

library SlotPacking128x64x64 {
    function set(
        uint128 a,
        uint64 b,
        uint64 c
    ) internal pure returns (bytes32 slot) {
        assembly {
            slot := or(slot, shl(128, a))
            slot := or(slot, shl(64, b))
            slot := or(slot, c)
        }
        return slot;
    }

    function get(bytes32 slot)
        internal
        pure
        returns (
            uint128 a,
            uint64 b,
            uint64 c
        )
    {
        assembly {
            a := shr(128, slot)
            b := and(0xffffffffffffffff, shr(64, slot))
            c := and(0xffffffffffffffff, slot)
        }
    }
}
