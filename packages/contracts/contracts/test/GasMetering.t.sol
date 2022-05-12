//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import { console } from "forge-std/console.sol";
import { CommonTest } from "./CommonTest.t.sol";
import { GasMetering } from "../L1/GasMetering.sol";

contract A is GasMetering {
    function meter(uint64 requestedGas, uint256 mint) public {
        gasMetered(requestedGas, mint);
    }

    function initialize() public {
        uint128 prevBaseFee = 8;
        uint64 prevNum = 0;
        uint64 prevBoughtGas = 100;
        uint128 gasTargetLimit = 2_000_000;
        uint128 gasSanityLimit = 5_000_000;

        _initialize(
            prevBaseFee,
            prevNum,
            prevBoughtGas,
            gasTargetLimit,
            gasSanityLimit
        );
    }
}

contract GasMetering_Test is CommonTest {
    A internal a;

    function setUp() external {
        a = new A();
        a.initialize();
    }

    // TODO: needs more testing
    function test_meter() external {
        a.meter(100, 100);
    }
}
