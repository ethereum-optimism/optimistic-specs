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

    function getVersionFromNonce(uint256 _nonce)
        internal
        pure
        returns (uint16 version)
    {
        assembly {
            version := shr(240, _nonce)
        }
    }

    function getVersionedHash(
        uint256 _nonce,
        address _sender,
        address _target,
        uint256 _value,
        uint256 _gasLimit,
        bytes memory _data
    )
        internal
        pure
        returns (bytes32)
    {
        uint16 version = getVersionFromNonce(_nonce);

        if (version == 0) {
            return getHashV0(
                _target,
                _sender,
                _data,
                _nonce
            );
        } else if (version == 1) {
            return getHashV1(
                _nonce,
                _sender,
                _target,
                _value,
                _gasLimit,
                _data
            );
        } else {
            require(false, "unknown version");
        }
    }

    function getHashV0(
        address _target,
        address _sender,
        bytes memory _data,
        uint256 _nonce
    )
        internal
        pure
        returns (bytes32)
    {
        bytes memory xDomainCalldata = Lib_CrossDomainUtils.encodeXDomainCalldata(
            _target,
            _sender,
            _data,
            _nonce
        );

        return keccak256(xDomainCalldata);
    }

    function getHashV1(
        uint256 _nonce,
        address _sender,
        address _target,
        uint256 _value,
        uint256 _gasLimit,
        bytes memory _data
    )
        internal
        pure
        returns (bytes32)
    {
        return keccak256(abi.encode(_nonce, _sender, _target, _value, _gasLimit, _data));
    }
}
