//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import { console } from "forge-std/console.sol";

import { SlotPacking128x64x64 } from "../libraries/SlotPacking128x64x64.sol";
import { SlotPacking128x128 } from "../libraries/SlotPacking128x128.sol";
import { SaturatedMath } from "../libraries/SaturatedMath.sol";

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

    /**
     * @notice
     * TODO: rename variables + update spec with renamed variables
     */
    function gasMetered(uint64 requestedGas, uint256 mint) internal {
        (uint128 prevBaseFee, uint64 prevNum, uint64 prevBoughtGas) = SlotPacking128x64x64.get(
            storage1
        );

        (uint128 gasTargetLimit, uint128 gasSanityLimit) = SlotPacking128x128.get(storage2);
        uint128 gasTarget = gasTargetLimit / ELASTICITY_MULTIPLIER;

        // If the gas target is exactly hit, maintain the previous base fee
        uint128 nowBaseFee = prevBaseFee;
        uint64 nowBoughtGas = SaturatedMath.add_u64(prevBoughtGas, requestedGas);
        uint64 nowNum = uint64(block.number);

        // Check to see if there has been a deposit
        if (nowNum != prevNum) {
            nowBoughtGas = requestedGas;
            if (prevBoughtGas > gasTarget) {
                uint128 gasUsedDelta = SaturatedMath.sub_u128(prevBoughtGas, gasTarget);

                // math.max(baseFeePerGasDelta, 1);

                uint128 baseFeePerGasDelta = (prevBaseFee * gasUsedDelta) /
                    gasTarget /
                    BASE_FEE_MAX_CHANGE_DENOMINATOR;

                if (baseFeePerGasDelta == 0) {
                    baseFeePerGasDelta = 1;
                }
                nowBaseFee = SaturatedMath.sub_u128(prevBaseFee, baseFeePerGasDelta);
            } else if (prevBoughtGas < gasTarget) {
                uint128 gasUsedDelta = SaturatedMath.sub_u128(gasTarget, prevBoughtGas);

                uint128 baseFeePerGasDelta = (prevBaseFee * gasUsedDelta) /
                    gasTarget /
                    BASE_FEE_MAX_CHANGE_DENOMINATOR;

                nowBaseFee = SaturatedMath.sub_u128(prevBaseFee, baseFeePerGasDelta);
            }
        }

        require(nowBoughtGas < gasSanityLimit, "Cannot buy more L2 gas.");
        uint256 requiredLockup = SaturatedMath.mul_u128(requestedGas, nowBaseFee);

        _burnGas(requiredLockup);

        storage1 = SlotPacking128x64x64.set(nowBaseFee, nowNum, nowBoughtGas);
    }
}
