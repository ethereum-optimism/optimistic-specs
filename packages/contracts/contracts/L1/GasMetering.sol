//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

// TODO: update the spec with nicer variable names and then
// use the variable names 1:1 in the logic so that its easy
// to verify
abstract contract GasMetering {
    uint256 internal storage1;
    uint256 internal storage2;

    // TODO: remove these getters?
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

    // 128 bits basefee
    // 64 bites blocknumber
    // 64 bits boughtGas

    // basefee, blocknumber, boughtGas

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

    // TODO
    function _packAndStore(
        uint256 baseFee,
        uint256 num,
        uint256 boughtGas
    ) internal {
        uint256 slot;
        // 0 1 1 1
        // 0 0 0 0
        assembly {
            slot := or(slot, shl())
        }
        storage1 = slot;
    }

    function _burn(uint256 amount) internal {
        uint256 i = 0;
        uint256 gas = gasleft();
        uint256 target = gas - amount;
        while (target < gasleft()) {
            i++;
        }
    }

    modifier gasMetered(uint256 requestedGas, uint256 mint) {
        // prev_basefee, prev_num, prev_bought_gas = load_and_unpack_storage()
        (uint128 prevBaseFee, uint64 prevNum, uint64 prevBoughtGas) = getState();
        // gas_target_limit, gas_sanity_limit = load_and_unpack_storage2()
        (uint64 gasTargetLimit, uint64 gasSanityLimit) = getGasLimits();

        // constants
        uint256 ELASTICITY_MULTIPLIER = 2;
        uint256 BASE_FEE_MAX_CHANGE_DENOMINATOR = 8;
        uint256 gasTarget = gasTargetLimit / ELASTICITY_MULTIPLIER;

        uint256 gasCost = requestedGas * baseFee;

        uint256 nowBaseFee = prevBaseFee;
        uint256 nowNum = block.number;
        uint256 nowBoughtGas =  prevBoughtGas + requestedGas;

        if (nowNum != prevNum) {
            uint256 nowBoughtGas = requestedGas;
            if (prevBoughtGas < gasTarget) {
                uint256 gasUsedDelta = prevBoughtGas - gasTarget;
                uint256 baseFeePerGasDelta =
                    prevBaseFee * gasUsedDelta / gasTarget / BASE_FEE_MAX_CHANGE_DENOMINATOR
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

        _burn(requiredLockup);
        _packAndStore(nowBaseFee, nowNum, nowBoughtGas);

        _;
    }
}
