// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

/* Library Imports */
import { Lib_OVMCodec } from "@eth-optimism/contracts/libraries/codec/Lib_OVMCodec.sol";

/* Interface Imports */
import { ICrossDomainMessenger } from "./ICrossDomainMessenger.sol";
import { WithdrawalVerifier } from "../libraries/Lib_WithdrawalVerifier.sol";

/**
 * @title IL1CrossDomainMessenger
 */
interface IL1CrossDomainMessenger is ICrossDomainMessenger {
    /**********
     * Events *
     **********/

    /**
     * @notice Emitted when a Transaction is deposited from L1 to L2. The parameters of this
     * event are read by the rollup node and used to derive deposit transactions on L2.
     */
    event TransactionDeposited(
        address indexed from,
        address indexed to,
        uint256 mint,
        uint256 value,
        uint64 gasLimit,
        bool isCreation,
        bytes data
    );

    /*******************
     * Data Structures *
     *******************/

    struct L2MessageInclusionProof {
        bytes32 stateRoot;
        Lib_OVMCodec.ChainBatchHeader stateRootBatchHeader;
        Lib_OVMCodec.ChainInclusionProof stateRootProof;
        bytes stateTrieWitness;
        bytes storageTrieWitness;
    }

    /********************
     * Public Functions *
     ********************/

    /**
     * Relays a cross domain message to a contract.
     */
    function relayMessage(
        uint256 _nonce,
        address _sender,
        address _target,
        uint256 _value,
        uint256 _gasLimit,
        bytes calldata _data,
        uint256 _l2Timestamp,
        WithdrawalVerifier.OutputRootProof calldata _outputRootProof,
        bytes calldata _withdrawalProof
    ) external payable;
}
