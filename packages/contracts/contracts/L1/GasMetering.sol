//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import { console } from "forge-std/console.sol";

import { SlotPacking128x64x64 } from "../libraries/SlotPacking128x64x64.sol";
import { SlotPacking128x128 } from "../libraries/SlotPacking128x128.sol";

// TODO: update the spec with nicer variable names and then
// use the variable names 1:1 in the logic so that its easy
// to verify
abstract contract GasMetering {
    // 128 bits basefee
    // 64 bites blocknumber
    // 64 bits boughtGas

    // basefee, blocknumber, boughtGas
    bytes32 internal storage1;

    // gaslimit, sane gas limit
    // TODO: need a method for setting these values
    bytes32 internal storage2;

    uint8 internal constant ELASTICITY_MULTIPLIER = 2;
    uint8 internal constant BASE_FEE_MAX_CHANGE_DENOMINATOR = 8;

    function _initialize(
        uint128 prevBaseFee,
        uint64 prevNum,
        uint64 prevBoughtGas,
        uint128 gasTargetLimit,
        uint128 gasSanityLimit
    ) internal {
        require(storage1 == bytes32(0) && storage2 == bytes32(0));

        storage1 = SlotPacking128x64x64.set(prevBaseFee, prevNum, prevBoughtGas);

        storage2 = SlotPacking128x128.set(gasTargetLimit, gasSanityLimit);
    }

    function _burnGas(uint256 amount) internal view {
        uint256 i = 0;
        uint256 gas = gasleft();
        while (gas - amount < gasleft()) {
            ++i;
        }
    }

    // TODO: rename variables + update spec with renamed variables
    function gasMetered(uint64 requestedGas, uint256 mint) internal {
        (uint128 prevBaseFee, uint64 prevNum, uint64 prevBoughtGas) = SlotPacking128x64x64.get(
            storage1
        );

        (uint128 gasTargetLimit, uint128 gasSanityLimit) = SlotPacking128x128.get(storage2);
        uint128 gasTarget = gasTargetLimit / ELASTICITY_MULTIPLIER;

        // business logic below
        uint128 nowBaseFee = prevBaseFee;
        uint64 nowBoughtGas = prevBoughtGas + requestedGas;
        uint64 nowNum = uint64(block.number);

        if (nowNum != prevNum) {
            nowBoughtGas = requestedGas;
            if (prevBoughtGas > gasTarget) {
                uint128 gasUsedDelta = prevBoughtGas - gasTarget;
                uint128 baseFeePerGasDelta = (prevBaseFee * gasUsedDelta) /
                    gasTarget /
                    BASE_FEE_MAX_CHANGE_DENOMINATOR;
                if (baseFeePerGasDelta == 0) {
                    baseFeePerGasDelta = 1;
                }
                nowBaseFee = prevBaseFee - baseFeePerGasDelta;
            } else if (prevBoughtGas < gasTarget) {
                uint128 gasUsedDelta = gasTarget - prevBoughtGas;
                uint128 baseFeePerGasDelta = (prevBaseFee * gasUsedDelta) /
                    gasTarget /
                    BASE_FEE_MAX_CHANGE_DENOMINATOR;
                nowBaseFee = prevBaseFee - baseFeePerGasDelta;
            }
        }

        require(nowBoughtGas < gasSanityLimit, "Cannot buy more L2 gas.");
        uint256 requiredLockup = mint + (requestedGas * nowBaseFee);

        // in payable version:
        // require(msg.value > requiredLockup);

        _burnGas(requiredLockup);

        // TODO: these type conversions are unsafe
        storage1 = SlotPacking128x64x64.set(nowBaseFee, nowNum, nowBoughtGas);
    }
}
