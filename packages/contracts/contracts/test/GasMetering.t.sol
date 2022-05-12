//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import { console } from "forge-std/console.sol";
import { CommonTest } from "./CommonTest.t.sol";
import { GasMetering } from "../L1/GasMetering.sol";

contract A is GasMetering {
    function meter(uint256 requestedGas, uint256 mint)
        gasMetered(requestedGas, mint)
        public
    {}
}

contract GasMetering_Test is CommonTest {
    A internal a;

    function setUp() external {
        a = new A();
    }

    function foo() external {
        console.log("lol");
    }
}
