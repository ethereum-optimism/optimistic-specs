// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

/* Library Imports */
import {
    Lib_SecureMerkleTrie
} from "@eth-optimism/contracts/libraries/trie/Lib_SecureMerkleTrie.sol";

/**
 * @title TestLib_SecureMerkleTrie
 */
contract TestLib_SecureMerkleTrie {
    function verifyInclusionProof(
        bytes memory _key,
        bytes memory _value,
        bytes memory _proof,
        bytes32 _root
    ) public pure returns (bool) {
        return Lib_SecureMerkleTrie.verifyInclusionProof(_key, _value, _proof, _root);
    }

    function update(
        bytes memory _key,
        bytes memory _value,
        bytes memory _proof,
        bytes32 _root
    ) public pure returns (bytes32) {
        return Lib_SecureMerkleTrie.update(_key, _value, _proof, _root);
    }

    function get(
        bytes memory _key,
        bytes memory _proof,
        bytes32 _root
    ) public pure returns (bool, bytes memory) {
        return Lib_SecureMerkleTrie.get(_key, _proof, _root);
    }

    function getSingleNodeRootHash(bytes memory _key, bytes memory _value)
        public
        pure
        returns (bytes32)
    {
        return Lib_SecureMerkleTrie.getSingleNodeRootHash(_key, _value);
    }
}
