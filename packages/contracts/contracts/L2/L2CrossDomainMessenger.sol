// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

// solhint-disable max-line-length
/* Library Imports */
import { AddressAliasHelper } from "@eth-optimism/contracts/standards/AddressAliasHelper.sol";
import {
    Lib_DefaultValues
} from "@eth-optimism/contracts/libraries/constants/Lib_DefaultValues.sol";
import { CrossDomainHashing } from "../libraries/Lib_CrossDomainHashing.sol";

/* Interface Imports */
import { IL2CrossDomainMessenger } from "../interfaces/IL2CrossDomainMessenger.sol";

/* Interaction imports */
import { Burner } from "./Burner.sol";
import { CrossDomainMessenger } from "../universal/CrossDomainMessenger.sol";
// solhint-enable max-line-length

/**
 * @title L2CrossDomainMessenger
 * @notice The L2CrossDomainMessenger contract facilitates sending both ETH value and data from L2 to L1.
 * It is predeployed in the L2 state at address 0x4200000000000000000000000000000000000016.
 */
contract L2CrossDomainMessenger is
    CrossDomainMessenger
{
    /**********
     * Events *
     **********/

    /// @notice Emitted when the balance of this contract is burned.
    event WithdrawerBalanceBurnt(uint256 indexed amount);

    /********************
     * Public Functions *
     ********************/

    function initialize(
        address _l1CrossDomainMessenger
    )
        external
        initializer
    {
        _initialize(_l1CrossDomainMessenger);
    }

    /**
     * Removes all ETH held in this contract from the state, by deploying a contract which
     * immediately self destructs. For simplicity, this call is not incentivized as it costs very
     * little to run. Inspired by:
     * https://etherscan.io/address/0xb69fba56b2e67e7dda61c8aa057886a8d1468575#code
     */
    function burn() external {
        uint256 balance = address(this).balance;
        new Burner{ value: balance }();
        emit WithdrawerBalanceBurnt(balance);
    }

    /**********************
     * Internal Functions *
     **********************/

    function _verifyMessageProof(
        bytes32 _versionedHash,
        uint256 _nonce,
        address _sender,
        address _target,
        uint256 _value,
        uint256 _gasLimit,
        bytes calldata _data,
        bytes calldata _proof
    ) internal view override returns (bool) {
        // TODO
    }

    function _sendMessage(
        address _to,
        address _from,
        uint256 _value,
        uint64 _gasLimit,
        bool _isCreation,
        bytes memory _data
    ) internal override {
        require(
            !_isCreation,
            "Contract creations not allowed for L2 to L1 messages."
        );

        // TODO: Send message to L2ToL1MessagePasser
    }
}
