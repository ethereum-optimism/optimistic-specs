pragma solidity 0.8.10;

 interface DepositFeed {
     function depositTransaction(
         address _to,
         uint256 _value,
         uint256 _gasLimit,
         bool _isCreation,
         bytes memory _data
     ) external payable;
 }

 contract MultiDepositor {
     DepositFeed df = DepositFeed(0xDeaDDEaDDeAdDeAdDEAdDEaddeAddEAdDEAd0001);

     function deposit() external payable {
         for (uint i = 0; i < 3; i++) {
             df.depositTransaction(
                 0x7770000000000000000000000000000000000000,
                 1000,
                 3000000,
                 false,
                 ""
             );
         }
     }
 }