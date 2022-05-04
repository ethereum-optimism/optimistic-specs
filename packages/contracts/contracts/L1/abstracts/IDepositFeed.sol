pragma solidity ^0.8.10;

interface IDepositFeed {
    function depositTransaction(
        address _to,
        uint256 _value,
        uint256 _additionalGasPrice,
        uint64 _additionalGasLimit,
        uint64 _guaranteedGas,
        bool _isCreation,
        bytes memory _data
    ) external payable;
}
