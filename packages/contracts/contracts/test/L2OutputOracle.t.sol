//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

/* Testing utilities */
import { CommonTest } from "./CommonTest.t.sol";

import {
    Lib_PredeployAddresses
} from "@eth-optimism/contracts/libraries/constants/Lib_PredeployAddresses.sol";

import { Lib_BedrockPredeployAddresses } from "../libraries/Lib_BedrockPredeployAddresses.sol";
import { L2OutputOracle } from "../L1/L2OutputOracle.sol";
import { LibRLP } from "./Lib_RLP.t.sol";

import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract L2OutputOracle_Initializer is CommonTest {
    // Utility variables
    uint256 appendedTimestamp;

    // Test target
    L2OutputOracle oracle;

    // Constructor arguments
    address sequencer = 0x000000000000000000000000000000000000AbBa;
    uint256 submissionInterval = 1800;
    uint256 l2BlockTime = 2;
    bytes32 genesisL2Output = keccak256(abi.encode(0));
    uint256 historicalTotalBlocks = 100;

    // Cache of the initial L2 timestamp
    uint256 startingBlockTimestamp;

    // By default the first block has timestamp zero, which will cause underflows in the tests
    uint256 initTime = 1000;

    constructor() {
        // Move time forward so we have a non-zero starting timestamp
        vm.warp(initTime);
        // Deploy the L2OutputOracle and transfer owernship to the sequencer
        oracle = new L2OutputOracle(
            submissionInterval,
            l2BlockTime,
            genesisL2Output,
            historicalTotalBlocks,
            initTime,
            sequencer
        );
        startingBlockTimestamp = block.timestamp;
    }
}

// Define this test in a standalone contract to ensure it runs immediately after the constructor.
contract L2OutputOracleTest_Constructor is L2OutputOracle_Initializer {
    function test_constructor() external {
        /*
        assertEq(oracle.owner(), sequencer);
        assertEq(oracle.SUBMISSION_INTERVAL(), submissionInterval);
        assertEq(oracle.L2_BLOCK_TIME(), l2BlockTime);
        assertEq(oracle.HISTORICAL_TOTAL_BLOCKS(), historicalTotalBlocks);
        assertEq(oracle.latestBlockTimestamp(), startingBlockTimestamp);
        assertEq(oracle.STARTING_BLOCK_TIMESTAMP(), startingBlockTimestamp);
        assertEq(oracle.getL2Output(startingBlockTimestamp), genesisL2Output);
        */
    }
}

contract L2OutputOracleTest is L2OutputOracle_Initializer {
    bytes32 appendedOutput1 = keccak256(abi.encode(1));

    constructor() {
        appendedTimestamp = oracle.nextTimestamp();

        // Warp to after the timestamp we'll append
        vm.warp(appendedTimestamp + 1);
        vm.prank(sequencer);
        oracle.appendL2Output(appendedOutput1, appendedTimestamp, 0, 0);
    }

    /****************
     * Getter Tests *
     ****************/

    // Test: latestBlockTimestamp() should return the correct value
    function test_latestBlockTimestamp() external {
        assertEq(oracle.latestBlockTimestamp(), appendedTimestamp);
    }

    // Test: getL2Output() should return the correct value
    function test_getL2Output() external {
        assertEq(true, false);
        // assertEq(oracle.getL2Output(appendedTimestamp), appendedOutput1);
        // assertEq(oracle.getL2Output(appendedTimestamp + 1), 0);

    }

    // Test: nextTimestamp() should return the correct value
    function test_nextTimestamp() external {
        assertEq(
            oracle.nextTimestamp(),
            // The return value should match this arithmetic
            initTime + submissionInterval * 2
        );
    }

    // Test: computeL2BlockNumber() should return the correct value
    function test_computeL2BlockNumber() external {
        // Test with the timestamp of the very first appended block
        uint256 argTimestamp = startingBlockTimestamp;
        uint256 expected = historicalTotalBlocks + 1;
        assertEq(oracle.computeL2BlockNumber(argTimestamp), expected);

        // Test with an integer multiple of the l2BlockTime
        argTimestamp = startingBlockTimestamp + 20;
        expected = historicalTotalBlocks + 1 + (20 / l2BlockTime);
        assertEq(oracle.computeL2BlockNumber(argTimestamp), expected);

        // Test with a remainder
        argTimestamp = startingBlockTimestamp + 33;
        expected = historicalTotalBlocks + 1 + (33 / l2BlockTime);
        assertEq(oracle.computeL2BlockNumber(argTimestamp), expected);
    }
    // Test: computeL2BlockNumber() fails with a blockNumber from before the startingBlockTimestamp
    function testCannot_computePreHistoricalL2BlockNumber() external {
        bytes memory expectedError = "Timestamp prior to startingBlockTimestamp";
        uint256 argTimestamp = startingBlockTimestamp - 1;
        vm.expectRevert(expectedError);
        oracle.computeL2BlockNumber(argTimestamp);
    }

    /*****************************
     * Append Tests - Happy Path *
     *****************************/

    // Test: appendL2Output succeeds when given valid input, and no block hash and number are
    // specified.
    function test_appendingAnotherOutput() external {
        assertEq(false, true);
        /*
        bytes32 appendedOutput2 = keccak256(abi.encode(2));
        uint256 nextTimestamp = oracle.nextTimestamp();

        // Ensure the submissionInterval is enforced
        assertEq(nextTimestamp, appendedTimestamp + submissionInterval);

        vm.warp(nextTimestamp + 1);
        vm.prank(sequencer);
        oracle.appendL2Output(appendedOutput2, nextTimestamp, 0, 0);
        */
    }

    // Test: appendL2Output succeeds when given valid input, and when a block hash and number are
    // specified for reorg protection.
    // This tests is disabled (w/ skip_ prefix) because all blocks in Foundry currently have a
    // blockhash of zero.
    function skip_test_appendWithBlockhashAndHeight() external {
        assertEq(false, true);
        /*
        // Move ahead to block 100 so that we can reference historical blocks
        vm.roll(100);

        // Get the number and hash of a previous block in the chain
        uint256 l1BlockNumber = block.number - 1;
        bytes32 l1BlockHash = blockhash(l1BlockNumber);

        uint256 nextTimestamp = oracle.nextTimestamp();
        vm.warp(nextTimestamp + 1);
        vm.prank(sequencer);

        // Changing the l1BlockNumber argument should break this tests, however it does not
        // per the comment preceding this test.
        oracle.appendL2Output(nonZeroHash, nextTimestamp, l1BlockHash, l1BlockNumber);
        */
    }

    /***************************
     * Append Tests - Sad Path *
     ***************************/

    // Test: appendL2Output fails if called by a party that is not the sequencer.
    function testCannot_appendOutputIfNotSequencer() external {
        assertEq(false, true);
        /*
        uint256 nextTimestamp = oracle.nextTimestamp();

        vm.warp(nextTimestamp + 1);
        vm.expectRevert("Ownable: caller is not the owner");
        oracle.appendL2Output(nonZeroHash, nextTimestamp, 0, 0);
        */
    }

    // Test: appendL2Output fails given a zero blockhash.
    function testCannot_appendEmptyOutput() external {
        assertEq(false, true);
        /*
        bytes32 outputToAppend = bytes32(0);
        uint256 nextTimestamp = oracle.nextTimestamp();
        vm.warp(nextTimestamp + 1);
        vm.prank(sequencer);
        vm.expectRevert("Cannot submit empty L2 output");
        oracle.appendL2Output(outputToAppend, nextTimestamp, 0, 0);
        */
    }

    // Test: appendL2Output fails if the timestamp doesn't match the next expected timestamp.
    function testCannot_appendUnexpectedTimestamp() external {
        uint256 nextTimestamp = oracle.nextTimestamp();
        vm.warp(nextTimestamp + 1);
        vm.prank(sequencer);
        vm.expectRevert("Timestamp not equal to next expected timestamp");
        oracle.appendL2Output(nonZeroHash, nextTimestamp - 1, 0, 0);
    }

    // Test: appendL2Output fails if the timestamp is equal to the current L1 timestamp.
    function testCannot_appendCurrentTimestamp() external {
        uint256 nextTimestamp = oracle.nextTimestamp();
        vm.warp(nextTimestamp + 1);
        vm.prank(sequencer);
        vm.expectRevert("Cannot append L2 output in future");
        oracle.appendL2Output(nonZeroHash, block.timestamp, 0, 0);
    }

    // Test: appendL2Output fails if the timestamp is in the future.
    function testCannot_appendFutureTimestamp() external {
        uint256 nextTimestamp = oracle.nextTimestamp();
        vm.warp(nextTimestamp + 1);
        vm.prank(sequencer);
        vm.expectRevert("Cannot append L2 output in future");
        oracle.appendL2Output(nonZeroHash, block.timestamp + 1, 0, 0);
    }

    // Test: appendL2Output fails when given valid input, but the block hash and number do not
    // match.
    // This tests is disabled (w/ skip_ prefix) because all blocks in Foundry currently have a
    // blockhash of zero.
    function skip_testCannot_AppendWithUnmatchedBlockhash() external {
        // Move ahead to block 100 so that we can reference historical blocks
        vm.roll(100);

        // Get the number and hash of a previous block in the chain
        uint256 l1BlockNumber = block.number - 1;
        bytes32 l1BlockHash = blockhash(l1BlockNumber);

        uint256 nextTimestamp = oracle.nextTimestamp();
        vm.warp(nextTimestamp + 1);
        vm.prank(sequencer);

        // This will fail when foundry no longer returns zerod block hashes
        oracle.appendL2Output(nonZeroHash, nextTimestamp, l1BlockHash, l1BlockNumber - 1);
    }

    /****************
     * Delete Tests *
     ****************/

    event l2OutputDeleted(bytes32 indexed _l2Output, uint256 indexed _l2timestamp);
    function test_deleteL2Output() external {
        /*
        uint256 latestBlockTimestamp = oracle.latestBlockTimestamp();
        bytes32 outputToDelete = oracle.getL2Output(latestBlockTimestamp);
        bytes32 newLatestOutput = oracle.getL2Output(latestBlockTimestamp - submissionInterval);

        vm.prank(sequencer);
        vm.expectEmit(true, true, false, false);
        emit l2OutputDeleted(outputToDelete, latestBlockTimestamp);
        oracle.deleteL2Output(outputToDelete);

        // validate latestBlockTimestamp has been reduced
        uint256 latestBlockTimestampAfter = oracle.latestBlockTimestamp();
        assertEq(
            latestBlockTimestamp - submissionInterval,
            latestBlockTimestampAfter
        );

        // validate that the new latest output is as expected.
        assertEq(
            newLatestOutput,
            oracle.getL2Output(latestBlockTimestampAfter)
        );
        */
    }

    function testCannot_deleteL2Output_ifNotSequencer() external {
        /*
        uint256 latestBlockTimestamp = oracle.latestBlockTimestamp();
        bytes32 outputToDelete = oracle.getL2Output(latestBlockTimestamp);

        vm.expectRevert("Ownable: caller is not the owner");
        oracle.deleteL2Output(outputToDelete);
    }

    function testCannot_deleteL2Output_ifWrongOutput() external {
        uint256 previousBlockTimestamp = oracle.latestBlockTimestamp() - submissionInterval;
        bytes32 outputToDelete = oracle.getL2Output(previousBlockTimestamp);

        vm.prank(sequencer);
        vm.expectRevert("Can only delete the most recent output.");
        oracle.deleteL2Output(outputToDelete);
        */
    }
}
