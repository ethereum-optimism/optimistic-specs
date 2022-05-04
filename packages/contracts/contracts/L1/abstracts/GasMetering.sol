//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

abstract contract GasMetering {

    // TODO: i think solc is smart enough to pack
    // things into a single slot when they are
    // smaller than uint256 and next to each other
    // but not sure what will happen if there
    // are getters for them or if they are public
    // https://github.com/cronus-finance/cronus-core/blob/0ad76ef2498766954f2494700418160732f69c24/contracts/CronusPair.sol#L23

    uint256 internal storage1;
    uint256 internal storage2;

    function getBaseFee() public view returns (uint128) {
        (uint128 basefee , ,) = getState();
        return basefee;
    }

    function getBlockNumber() public view returns (uint64) {
        (, uint64 blockNumber, ) = getState();
        return blockNumber;
    }

    function getBoughtGas() public view returns (uint64) {
        (, , uint64 boughtGas) = getState();
        return boughtGas;
    }

    function getTargetGasLimit() public view returns (uint64) {
        (uint64 targetGasLimit, ) = getGasLimits();
        return targetGasLimit;
    }

    function getSanityGasLimit() public view returns (uint64) {
        (, uint64 sanityGasLimit) = getGasLimits();
        return sanityGasLimit;
    }

    function getState() internal view returns (uint128 a, uint64 b, uint64 c) {
        uint256 s = storage1;
        assembly {
            a := shr(128, s)
            b := and(0xffffffffffffffff, shr(64, s))
            c := and(0xffffffffffffffff, s)
        }
        return (a, b, c);
    }

    function getGasLimits() internal view returns (uint64 a, uint64 b) {
        uint256 s = storage2;
        assembly {
            a := and(0xffffffffffffffff, s)
            b := and(0xffffffffffffffff, shr(64, s))
        }
    }

    modifier gasMetered() {
        (uint128 baseFee, uint64 blockNumber, uint64 boughtGas) = getState();
        (uint64 gasTarget, uint64 gasSanity) = getGasLimits();

        // TODO: spec is inconsistent
        /*
        if blockNumber != block.number {
            if boughtGas == gas_target {
                new_basefee := curr_basefee
            } else if curr_bought_gas > gas_target {
                gas_delta     := curr_bought_gas - gas_target
                basefee_delta := gas_delta * curr_basefee // gas_target // BASE_FEE_MAX_CHANGE_DENOMINATOR
                basefee_delta := max(basefee_delta, 1) # TODO: Why does 1559 have this asymmetry?
                new_basefee   := curr_basefee + basefee_delta
            } else {
                gas_delta     := gas_target - curr_bought_gas
                basefee_delta := gas_delta * curr_basefee // gas_target // BASE_FEE_MAX_CHANGE_DENOMINATOR
                // Fun fact, geth doesn't let the new_basefee get below 0.
                new_basefee   := curr_basefee - basefee_delta
            }
            curr_basefee := new_basefee
            curr_number := block.number
            curr_bought_gas := 0
            pack_and_store(curr_basefee, curr_number, curr_bought_gas)
        }
        */
        _;
    }
}
