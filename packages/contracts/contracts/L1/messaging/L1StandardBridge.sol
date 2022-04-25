// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

/* Interface Imports */
import { IL1StandardBridge } from "@eth-optimism/contracts/L1/messaging/IL1StandardBridge.sol";
import { IL1ERC20Bridge } from "@eth-optimism/contracts/L1/messaging/IL1ERC20Bridge.sol";
import { IL2ERC20Bridge } from "@eth-optimism/contracts/L2/messaging/IL2ERC20Bridge.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { OptimismPortal } from "../OptimismPortal.sol";

/* Library Imports */
import {
    Lib_PredeployAddresses
} from "@eth-optimism/contracts/libraries/constants/Lib_PredeployAddresses.sol";
import { Address } from "@openzeppelin/contracts/utils/Address.sol";
import { SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

/**
 * @title L1StandardBridge
 * @dev The L1 ETH and ERC20 Bridge is a contract which stores deposited L1 funds and standard
 * tokens that are in use on L2. It synchronizes a corresponding L2 Bridge, informing it of deposits
 * and listening to it for newly finalized withdrawals.
 *
 */
contract L1StandardBridge is IL1StandardBridge {
    using SafeERC20 for IERC20;

    /********************************
     * External Contract References *
     ********************************/

    OptimismPortal public optimismPortal;
    address public l2TokenBridge;

    // Maps L1 token to L2 token to balance of the L1 token deposited
    mapping(address => mapping(address => uint256)) public deposits;

    /***************
     * Constructor *
     ***************/

    // This contract lives behind a proxy, so the constructor parameters will go unused.
    constructor() {}

    /******************
     * Initialization *
     ******************/

    /**
     * @param _optimismPortal OptimismPortal address.
     * @param _l2TokenBridge L2 standard bridge address.
     */
    // slither-disable-next-line external-function
    function initialize(OptimismPortal _optimismPortal, address _l2TokenBridge) public {
        require(address(optimismPortal) == address(0), "Contract has already been initialized.");
        optimismPortal = _optimismPortal;
        l2TokenBridge = _l2TokenBridge;
    }

    /**************
     * Depositing *
     **************/

    /** @dev Modifier requiring sender to be EOA. This check could be bypassed by a malicious
     *  contract via initcode, but it takes care of the user error we want to avoid.
     */
    modifier onlyEOA() {
        // Used to stop deposits from contracts (avoid accidentally lost tokens)
        require(!Address.isContract(msg.sender), "Account not EOA");
        _;
    }

    /**
     * @dev This function can be called with no data
     * to deposit an amount of ETH to the caller's balance on L2.
     * Since the receive function doesn't take data, a conservative
     * default amount is forwarded to L2.
     */
    receive() external payable onlyEOA {
        _initiateETHDeposit(msg.sender, msg.sender, 200_000, bytes(""));
    }

    /**
     * @inheritdoc IL1StandardBridge
     */
    function depositETH(uint32 _l2Gas, bytes calldata _data) external payable onlyEOA {
        _initiateETHDeposit(msg.sender, msg.sender, _l2Gas, _data);
    }

    /**
     * @inheritdoc IL1StandardBridge
     */
    function depositETHTo(
        address _to,
        uint32 _l2Gas,
        bytes calldata _data
    ) external payable {
        _initiateETHDeposit(msg.sender, _to, _l2Gas, _data);
    }

    /**
     * @dev Performs the logic for deposits by storing the ETH and informing the L2 ETH Gateway of
     * the deposit.
     * @param _from Account to pull the deposit from on L1.
     * @param _to Account to give the deposit to on L2.
     * @param _l2Gas Gas limit required to complete the deposit on L2.
     * @param _data Optional data to forward to L2. This data is provided
     *        solely as a convenience for external contracts. Aside from enforcing a maximum
     *        length, these contracts provide no guarantees about its content.
     */
    function _initiateETHDeposit(
        address _from,
        address _to,
        uint32 _l2Gas,
        bytes memory _data
    ) internal {
        // Construct calldata for finalizeDeposit call
        bytes memory message = abi.encodeWithSelector(
            IL2ERC20Bridge.finalizeDeposit.selector,
            address(0),
            Lib_PredeployAddresses.OVM_ETH,
            _from,
            _to,
            msg.value,
            _data
        );

        emit ETHDepositInitiated(_from, _to, msg.value, _data);

        // Send calldata into L2
        optimismPortal.depositTransaction(
            Lib_PredeployAddresses.L2_STANDARD_BRIDGE,
            0,
            _l2Gas,
            false,
            message
        );
    }

    /**
     * @inheritdoc IL1ERC20Bridge
     */
    function depositERC20(
        address _l1Token,
        address _l2Token,
        uint256 _amount,
        uint32 _l2Gas,
        bytes calldata _data
    ) external virtual onlyEOA {
        _initiateERC20Deposit(_l1Token, _l2Token, msg.sender, msg.sender, _amount, _l2Gas, _data);
    }

    /**
     * @inheritdoc IL1ERC20Bridge
     */
    function depositERC20To(
        address _l1Token,
        address _l2Token,
        address _to,
        uint256 _amount,
        uint32 _l2Gas,
        bytes calldata _data
    ) external virtual {
        _initiateERC20Deposit(_l1Token, _l2Token, msg.sender, _to, _amount, _l2Gas, _data);
    }

    /**
     * @dev Performs the logic for deposits by informing the L2 Deposited Token
     * contract of the deposit and calling a handler to lock the L1 funds. (e.g. transferFrom)
     *
     * @param _l1Token Address of the L1 ERC20 we are depositing
     * @param _l2Token Address of the L1 respective L2 ERC20
     * @param _from Account to pull the deposit from on L1
     * @param _to Account to give the deposit to on L2
     * @param _amount Amount of the ERC20 to deposit.
     * @param _l2Gas Gas limit required to complete the deposit on L2.
     * @param _data Optional data to forward to L2. This data is provided
     *        solely as a convenience for external contracts. Aside from enforcing a maximum
     *        length, these contracts provide no guarantees about its content.
     */
    function _initiateERC20Deposit(
        address _l1Token,
        address _l2Token,
        address _from,
        address _to,
        uint256 _amount,
        uint32 _l2Gas,
        bytes calldata _data
    ) internal {
        // When a deposit is initiated on L1, the L1 Bridge transfers the funds to itself for future
        // withdrawals. safeTransferFrom also checks if the contract has code, so this will fail if
        // _from is an EOA or address(0).
        // slither-disable-next-line reentrancy-events, reentrancy-benign
        IERC20(_l1Token).safeTransferFrom(_from, address(this), _amount);

        // Construct calldata for _l2Token.finalizeDeposit(_to, _amount)
        bytes memory message = abi.encodeWithSelector(
            IL2ERC20Bridge.finalizeDeposit.selector,
            _l1Token,
            _l2Token,
            _from,
            _to,
            _amount,
            _data
        );

        // slither-disable-next-line reentrancy-benign
        deposits[_l1Token][_l2Token] = deposits[_l1Token][_l2Token] + _amount;

        // slither-disable-next-line reentrancy-events
        emit ERC20DepositInitiated(_l1Token, _l2Token, _from, _to, _amount, _data);

        // Send calldata into L2
        optimismPortal.depositTransaction(
            Lib_PredeployAddresses.L2_STANDARD_BRIDGE,
            0,
            _l2Gas,
            false,
            message
        );
    }

    /*************************
     * Cross-chain Functions *
     *************************/

    /**
     * @notice Ensures that the caller is the portal, and that it has the l2Sender value
     * set to the address of the L2 Token Bridge.
     */
    modifier onlyL2Bridge() {
        require(
            msg.sender == address(optimismPortal),
            "Messages must be relayed by first calling the Optimism Portal"
        );
        require(
            optimismPortal.l2Sender() == l2TokenBridge,
            "Message must be sent from the L2 Token Bridge"
        );
        _;
    }

    /**
     * @inheritdoc IL1StandardBridge
     */
    function finalizeETHWithdrawal(
        address _from,
        address _to,
        uint256 _amount,
        bytes calldata _data
    ) external onlyL2Bridge {
        emit ETHWithdrawalFinalized(_from, _to, _amount, _data);

        (bool success, ) = _to.call{ value: _amount }(new bytes(0));
        require(success, "TransferHelper::safeTransferETH: ETH transfer failed");
    }

    /**
     * @inheritdoc IL1ERC20Bridge
     */
    function finalizeERC20Withdrawal(
        address _l1Token,
        address _l2Token,
        address _from,
        address _to,
        uint256 _amount,
        bytes calldata _data
    ) external onlyL2Bridge {
        deposits[_l1Token][_l2Token] = deposits[_l1Token][_l2Token] - _amount;
        emit ERC20WithdrawalFinalized(_l1Token, _l2Token, _from, _to, _amount, _data);

        // When a withdrawal is finalized on L1, the L1 Bridge transfers the funds to the withdrawer
        IERC20(_l1Token).safeTransfer(_to, _amount);
    }

    /*****************************
     * Temporary - Migrating ETH *
     *****************************/

    /**
     * @dev Adds ETH balance to the account. This is meant to allow for ETH
     * to be migrated from an old gateway to a new gateway.
     * NOTE: This is left for one upgrade only so we are able to receive the migrated ETH from the
     * old contract
     */
    // Bedrock upgrade note: we may use this if this contract gets merged into the portal,
    // or even if we just start using a new proxy.
    function donateETH() external payable {}
}
