// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

import {
    Lib_PredeployAddresses
} from "@eth-optimism/contracts/libraries/constants/Lib_PredeployAddresses.sol";
import { StandardBridge } from "../universal/StandardBridge.sol";
import { OptimismMintableERC20 } from "../universal/OptimismMintableERC20.sol";

/**
 * @title L2StandardBridge
 * @dev This contract is an L2 predeploy that is responsible for facilitating
 * deposits of tokens from L1 to L2.
 * TODO: ensure that this has 1:1 backwards compatibility
 */
contract L2StandardBridge is StandardBridge {
    /**********
     * Events *
     **********/

    event WithdrawalInitiated(
        address indexed _l1Token,
        address indexed _l2Token,
        address indexed _from,
        address _to,
        uint256 _amount,
        bytes _data
    );

    event DepositFinalized(
        address indexed _l1Token,
        address indexed _l2Token,
        address indexed _from,
        address _to,
        uint256 _amount,
        bytes _data
    );

    event DepositFailed(
        address indexed _l1Token,
        address indexed _l2Token,
        address indexed _from,
        address _to,
        uint256 _amount,
        bytes _data
    );

    /********************
     * Public Functions *
     ********************/

    /**
     * @notice Initialize the L2StandardBridge. This must only be callable
     * once. `_initialize` ensures this.
     */
    function initialize(address payable _otherBridge) public {
        _initialize(payable(Lib_PredeployAddresses.L2_CROSS_DOMAIN_MESSENGER), _otherBridge);
    }

    /**
     * @notice Withdraw tokens to self on L1
     * @param _l2Token The L2 token address to withdraw
     * @param _amount The amount of L2 token to withdraw
     * @param _minGasLimit The min gas limit in the withdrawing call
     * @param _data Additional calldata to pass along
     */
    function withdraw(
        address _l2Token,
        uint256 _amount,
        uint32 _minGasLimit,
        bytes calldata _data
    ) external payable virtual {
        _initiateWithdrawal(_l2Token, msg.sender, msg.sender, _amount, _minGasLimit, _data);
    }

    /**
     * @notice Withdraw tokens to an address on L1
     * @param _l2Token The L2 token address to withdraw
     * @param _to The L1 account to withdraw to
     * @param _amount The amount of L2 token to withdraw
     * @param _minGasLimit The min gas limit in the withdrawing call
     * @param _data Additional calldata to pass along
     */
    function withdrawTo(
        address _l2Token,
        address _to,
        uint256 _amount,
        uint32 _minGasLimit,
        bytes calldata _data
    ) external payable virtual {
        _initiateWithdrawal(_l2Token, msg.sender, _to, _amount, _minGasLimit, _data);
    }

    /**
     * @notice Finalize the L1 to L2 deposit. This should only be callable by
     * a deposit through the L1StandardBridge.
     * @param _l1Token The L1 token address
     * @param _l2Token The corresponding L2 token address
     * @param _from The sender of the tokens
     * @param _to The recipient of the tokens
     * @param _amount The amount of tokens
     * @param _data Additional calldata
     */
    function finalizeDeposit(
        address _l1Token,
        address _l2Token,
        address _from,
        address _to,
        uint256 _amount,
        bytes calldata _data
    ) external payable virtual {
        // Check to see if the bridge is being used to deposit ETH.
        // The `msg.value` must match the `_amount` to prevent
        // ETH from getting stuck in the contract
        if (
            _l1Token == address(0) &&
            _l2Token == Lib_PredeployAddresses.OVM_ETH &&
            msg.value == _amount
        ) {
            // An ETH deposit is being made via the Token Bridge.
            // We simply forward it on. If this call fails, ETH will be stuck, but the L1Bridge
            // uses onlyEOA on the receive function, so anyone sending to a contract knows
            // what they are doing.
            finalizeBridgeETH(_from, _to, _amount, _data);
            emit DepositFinalized(_l1Token, _l2Token, _from, _to, _amount, _data);
        } else if (
            _isOptimismMintable(_l2Token, _l1Token)
        ) // Check the target token is compliant and
        // verify the deposited token on L1 matches the L2 deposited token representation here
        // slither-disable-next-line reentrancy-events
        {
            // When a deposit is finalized, we credit the account on L2 with the same amount of
            // tokens.
            // slither-disable-next-line reentrancy-events
            finalizeBridgeERC20(_l2Token, _l1Token, _from, _to, _amount, _data);
            // slither-disable-next-line reentrancy-events
            emit DepositFinalized(_l1Token, _l2Token, _from, _to, _amount, _data);
        } else {
            // Either the L2 token which is being deposited-into disagrees about the correct address
            // of its L1 token, or does not support the correct interface.
            // This should only happen if there is a  malicious L2 token, or if a user somehow
            // specified the wrong L2 token address to deposit into.
            // In either case, we stop the process here and construct a withdrawal
            // message so that users can get their funds out in some cases.
            // There is no way to prevent malicious token contracts altogether, but this does limit
            // user error and mitigate some forms of malicious contract behavior.
            emit DepositFailed(_l1Token, _l2Token, _from, _to, _amount, _data);

            // Withdraw ETH in the case that the user submitted a bad ETH
            // deposit to prevent ETH from getting stuck
            // TODO: can this be wrapped into _initiateWithdrawal?
            // need to handle using `msg.value` here instead of `_value`
            if (_l2Token == Lib_PredeployAddresses.OVM_ETH) {
                _initiateBridgeETH(_from, _to, msg.value, 0, _data);
            } else {
                _initiateBridgeERC20(_l2Token, _l1Token, _from, _to, _amount, 0, _data);
            }
        }
    }

    /**********************
     * Internal Functions *
     **********************/

    /**
     * @notice Handle withdrawals, taking into account the legacy form of ETH
     * when it was represented as an ERC20 at the OVM_ETH contract.
     * TODO: require(msg.value == _value) for OVM_ETH case?
     */
    function _initiateWithdrawal(
        address _l2Token,
        address _from,
        address _to,
        uint256 _amount,
        uint32 _minGasLimit,
        bytes calldata _data
    ) internal {
        address l1Token = OptimismMintableERC20(_l2Token).l1Token();
        emit WithdrawalInitiated(l1Token, _l2Token, msg.sender, _to, _amount, _data);
        if (_l2Token == Lib_PredeployAddresses.OVM_ETH) {
            _initiateBridgeETH(_from, _to, _amount, _minGasLimit, _data);
        } else {
            _initiateBridgeERC20(_l2Token, l1Token, _from, _to, _amount, _minGasLimit, _data);
        }
    }
}
