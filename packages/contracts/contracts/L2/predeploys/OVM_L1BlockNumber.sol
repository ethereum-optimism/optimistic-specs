// SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import { iOVM_L1BlockNumber } from "./iOVM_L1BlockNumber.sol";
import { L1Block } from "../L1Block.sol";

contract OVM_L1BlockNumber is iOVM_L1BlockNumber {
    address public l1BlockAddress;

    constructor(address _l1BlockAddress) {
        l1BlockAddress = _l1BlockAddress;
    }

    function getL1BlockNumber() external view returns (uint256) {
        return L1Block(l1BlockAddress).number();
    }
}
