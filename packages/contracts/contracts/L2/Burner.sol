//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

/**
 * @title Burner
 */
contract Burner {
    constructor() payable {
        selfdestruct(payable(address(this)));
    }
}
