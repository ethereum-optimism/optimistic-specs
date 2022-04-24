//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

/* Testing utilities */
import { CommonTest } from "./CommonTest.sol";

/* Target contract dependencies */
import { L2OutputOracle } from "../L1/L2OutputOracle.sol";

/* Target contract */
import { OptimismPortal } from "../L1/OptimismPortal.sol";

contract OptimismPortal_Test is CommonTest {
    event TransactionDeposited(
        address indexed from,
        address indexed to,
        uint256 mint,
        uint256 value,
        uint256 gasLimit,
        bool isCreation,
        bytes data
    );

    // Dependencies
    L2OutputOracle oracle;

    OptimismPortal op;

    function setUp() external {
        // Oracle value is zero, but this test does not depend on it.
        op = new OptimismPortal(oracle, 7 days);
    }

    function test_receive_withEthValueFromEOA() external {
        // EOA emulation
        vm.prank(address(this), address(this));

        vm.expectEmit(true, true, false, true);
        emit TransactionDeposited(address(this), address(this), 100, 100, 30_000, false, hex"");

        (bool s, ) = address(op).call{ value: 100 }(hex"");
        s; // Silence the compiler's "Return value of low-level calls not used" warning.

        assertEq(address(op).balance, 100);
    }
}
