pragma solidity ^0.8.10;

// TODO: this should be removed as its a duplicate
// of IL2ERC20Bridge. I believe its only used
// in tests now
interface IL2StandardBridge {
    event DepositFailed(
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
    event WithdrawalInitiated(
        address indexed _l1Token,
        address indexed _l2Token,
        address indexed _from,
        address _to,
        uint256 _amount,
        bytes _data
    );

    function finalizeDeposit(
        address _l1Token,
        address _l2Token,
        address _from,
        address _to,
        uint256 _amount,
        bytes memory _data
    ) external payable;

    function l1TokenBridge() external view returns (address);

    function withdraw(
        address _l2Token,
        uint256 _amount,
        uint32 _l1Gas,
        bytes memory _data
    ) external payable;

    function withdrawTo(
        address _l2Token,
        address _to,
        uint256 _amount,
        uint32 _l1Gas,
        bytes memory _data
    ) external payable;
}
