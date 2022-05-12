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
    bytes32 internal storage2;

    uint256 constant internal ELASTICITY_MULTIPLIER = 2;
    uint256 constant internal BASE_FEE_MAX_CHANGE_DENOMINATOR = 8;

    function _burnGas(uint256 amount) internal view {
        uint256 i = 0;
        uint256 gas = gasleft();
        uint256 target = gas - amount;
        while (target < gasleft()) {
            i++;
        }
    }

    // TODO: rename variables + update spec with renamed variables
    modifier gasMetered(uint256 requestedGas, uint256 mint) {
        // prev_basefee, prev_num, prev_bought_gas = load_and_unpack_storage()
        (uint128 prevBaseFee, uint64 prevNum, uint64 prevBoughtGas) =
            SlotPacking128x64x64.get(storage1);

        // constants
        // gas_target_limit, gas_sanity_limit = load_and_unpack_storage2()
        (uint128 gasTargetLimit, uint128 gasSanityLimit) =
            SlotPacking128x128.get(storage2);
        uint256 gasTarget = gasTargetLimit / ELASTICITY_MULTIPLIER;

        // business logic below

        uint256 gasCost = requestedGas * prevBaseFee;

        uint256 nowBaseFee = prevBaseFee;
        uint256 nowNum = block.number;
        uint256 nowBoughtGas =  prevBoughtGas + requestedGas;

        if (nowNum != prevNum) {
            // TODO: this is in the spec
            // uint256 nowBoughtGas = requestedGas;
            if (prevBoughtGas < gasTarget) {
                uint256 gasUsedDelta = prevBoughtGas - gasTarget;
                uint256 baseFeePerGasDelta =
                    prevBaseFee * gasUsedDelta / gasTarget / BASE_FEE_MAX_CHANGE_DENOMINATOR;
                if (baseFeePerGasDelta == 0) {
                    baseFeePerGasDelta = 1;
                }
                nowBaseFee = prevBaseFee - baseFeePerGasDelta;
            } else if (prevBoughtGas > gasTarget) {
                uint256 gasUsedDelta = gasTarget - prevBoughtGas;
                uint256 baseFeePerGasDelta =
                    prevBaseFee * gasUsedDelta / gasTarget / BASE_FEE_MAX_CHANGE_DENOMINATOR;
                nowBaseFee = prevBaseFee - baseFeePerGasDelta;
            }
        }

        require(nowBoughtGas < gasSanityLimit);
        uint256 requiredLockup = mint + (requestedGas * nowBaseFee);

        // in payable version:
        // require(msg.value > requiredLockup);

        _burnGas(requiredLockup);

        // TODO: these type conversions are unsafe
        storage1 = SlotPacking128x64x64.set(
            uint128(nowBaseFee),
            uint64(nowNum),
            uint64(nowBoughtGas)
        );

        _;
    }
}
