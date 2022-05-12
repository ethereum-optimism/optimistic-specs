//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

library SlotPacking128x128 {
    function set(
        uint128 a,
        uint128 b
    )
    internal
    pure
    returns (bytes32 slot)
    {
        assembly {
            slot := or(slot, a)
            slot := shl(128, slot)
            slot := or(slot, b)
        }
        return slot;
    }

    function get(bytes32 slot)
    internal
    pure
    returns (uint128 a, uint128 b)
    {
        assembly {
            a := shr(128, slot)
            b := and(0xffffffffffffffffffffffffffffffff, slot)
        }
    }
}