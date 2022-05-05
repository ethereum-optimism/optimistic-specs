//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import {
    Lib_CrossDomainUtils
} from "@eth-optimism/contracts/libraries/bridge/Lib_CrossDomainUtils.sol";

library CrossDomainHashing {
    function addVersionToNonce(uint256 _nonce, uint16 _version)
        internal
        pure
        returns (uint256 nonce)
    {
        assembly {
            nonce := or(shl(240, _version), _nonce)
        }
    }

    function getVersionFromNonce(uint256 _nonce) internal pure returns (uint16 version) {
        assembly {
            version := shr(240, _nonce)
        }
    }

    function getVersionedEncoding(
        uint256 _nonce,
        address _sender,
        address _target,
        uint256 _value,
        uint256 _gasLimit,
        bytes memory _data
    ) internal pure returns (bytes memory) {
        uint16 version = getVersionFromNonce(_nonce);
        if (version == 0) {
            return getEncodingV0(_target, _sender, _data, _nonce);
        } else if (version == 1) {
            return getEncodingV1(_nonce, _sender, _target, _value, _gasLimit, _data);
        }

        revert("Unknown version.");
    }

    function getVersionedHash(
        uint256 _nonce,
        address _sender,
        address _target,
        uint256 _value,
        uint256 _gasLimit,
        bytes memory _data
    ) internal pure returns (bytes32) {
        uint16 version = getVersionFromNonce(_nonce);
        if (version == 0) {
            return getHashV0(_target, _sender, _data, _nonce);
        } else if (version == 1) {
            return getHashV1(_nonce, _sender, _target, _value, _gasLimit, _data);
        }

        revert("Unknown version.");
    }

    function getEncodingV0(
        address _target,
        address _sender,
        bytes memory _data,
        uint256 _nonce
    ) internal pure returns (bytes memory) {
        return Lib_CrossDomainUtils.encodeXDomainCalldata(_target, _sender, _data, _nonce);
    }

    function getEncodingV1(
        uint256 _nonce,
        address _sender,
        address _target,
        uint256 _value,
        uint256 _gasLimit,
        bytes memory _data
    ) internal pure returns (bytes memory) {
        return
            abi.encodeWithSignature(
                "relayMessage(uint256,address,address,uint256,uint256,bytes)",
                _nonce,
                _sender,
                _target,
                _value,
                _gasLimit,
                _data
            );
    }

    function getHashV0(
        address _target,
        address _sender,
        bytes memory _data,
        uint256 _nonce
    ) internal pure returns (bytes32) {
        return keccak256(getEncodingV0(_target, _sender, _data, _nonce));
    }

    function getHashV1(
        uint256 _nonce,
        address _sender,
        address _target,
        uint256 _value,
        uint256 _gasLimit,
        bytes memory _data
    ) internal pure returns (bytes32) {
        return keccak256(getEncodingV1(_nonce, _sender, _target, _value, _gasLimit, _data));
    }
}
