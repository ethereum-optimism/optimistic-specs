# Deposit Fee Spec

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Deposits on L1](#deposits-on-l1)
  - [Guaranteed Gas](#guaranteed-gas)
  - [Gas Stipend](#gas-stipend)
  - [Additional Gas](#additional-gas)
  - [Limiting Guaranteed Gas](#limiting-guaranteed-gas)
  - [Rationale for Guaranteed vs Additional Gas](#rationale-for-guaranteed-vs-additional-gas)
  - [Rationale for burning L1 Gas](#rationale-for-burning-l1-gas)
- [Deposits on L2](#deposits-on-l2)
  - [Guaranteed Gas On L2](#guaranteed-gas-on-l2)
  - [Additional Gas On L2](#additional-gas-on-l2)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Deposits on L1

L1 Deposits are augmented with the following fields:

- `guaranteedGas`
- `gasPrice`
- `additionalGas`

`guaranteedGas` and `additionalGas` must fit into a 64 bit unsigned integer.  
`gasPrice` must fit into a 256 bit unsigned integer.

### Guaranteed Gas

The **guaranteed gas** is composed of a gas stipend and any additional guaranteed
gas the user would like to buy on L1.

If the user requests more `guaranteedGas` than the `gasStipend`, that gas will
be bought with L1 Eth via a gas burn or by buying it directly.
It is not refundable if the transaction uses less gas than the gas limit.

### Gas Stipend

Can also provide a stipend to the guaranteed gas of every transaction.
The stipend is bought with gas that was spent executing the deposit logic;
however, the stipend still needs to be included in the sum of bought
guaranteed gas.

TODO: How much / do we actually need this?

If a gas stipend is provided, we only buy the amount of guaranteed gas
in excess of the gas stipend.

### Additional Gas

Users can request additional gas for their L2 transaction. This gas is
refundable is less than the gas limit is used. However, if there is not
enough available gas on L2 or the gas price not high enough, the additional
gas will not be provided on L2.

Additional gas is bought with the following:

- `gasPrice`: L2 gas price to buy.
- `additionalGas`: Amount of additional gas requested

The additional gas will be bought with the `from` account's balance. We do
not require that the user mint enough L2 eth to cover the additional gas fee,
but do suggest that tooling checks that the user has enough balance for it.

### Limiting Guaranteed Gas

The deposit feed contract must limit the total amount of guaranteed gas in a
single L1 block. This is to limit the amount of gas that is used on L2.

This is done with a hard limit on the amount of guaranteed gas that can be
bought. To reduce PGAs if the deposit mechanism is congested, we also implement
an EIP-1559 style fee market with the following pseudo code:

```text
BASE_FEE_MAX_CHANGE_DENOMINATOR = 8
ELASTICITY_MULTIPLIER = 2

curr_basefee: u128, curr_num: u64, curr_bought_gas: u64 = load_and_unpack_storage()
GUARANTEED_GAS_LIMIT: u64, SANITY_GAS_LIMIT: u64 = load_and_unpack_storage2()
gas_target = GUARANTEED_GAS_LIMIT // ELASTICITY_MULTIPLIER

# // implies floor division, however because gas_delta is always positive, it is the same as truncating (aka round to 0) division
# If first deposit of this block, calculate the new basefee and store other info as well.
if curr_num != block.number {
    if curr_bought_gas == gas_target {
        new_basefee := curr_basefee
    } else if curr_bought_gas > gas_target {
        gas_delta     := curr_bought_gas - gas_target
        basefee_delta := gas_delta * curr_basefee // gas_target // BASE_FEE_MAX_CHANGE_DENOMINATOR
        basefee_delta := max(basefee_delta, 1) # TODO: Why does 1559 have this asymmetry?
        new_basefee   := curr_basefee + basefee_delta
    } else {
        gas_delta     := gas_target - curr_bought_gas
        basefee_delta := gas_delta * curr_basefee // gas_target // BASE_FEE_MAX_CHANGE_DENOMINATOR
        // Fun fact, geth doesn't let the new_basefee get below 0 and while not in the EIP spec, we should add this as well.
        new_basefee   := curr_basefee - basefee_delta
    }
    curr_basefee := new_basefee
    curr_number := block.number
    curr_bought_gas := 0
   
}

curr_bought_gas += required_gas
require(curr_bought_gas <= GUARANTEED_GAS_LIMIT)
require(curr_bought_gas <= SANITY_GAS_LIMIT)
gas_cost = requested_gas * curr_basefee

burn(gas_cost) # Via gas or eth.

pack_and_store(curr_basefee, curr_number, curr_bought_gas)
```

### Rationale for Guaranteed vs Additional Gas

Need to have this additional gas mechanism to enable gas refunds.  
Guaranteed gas cannot be refunded without opening up a DOS vector.

### Rationale for burning L1 Gas

If we burn Eth (or collect it), we need to add the payable selector everywhere.
Adding it everywhere is not feasible and really bad UX.

## Deposits on L2

### Guaranteed Gas On L2

When processing L2 blocks, do the following:

1. Sum up all guaranteed gas in deposits.
2. Subtract the guaranteed gas from the gas pool.
3. Don't later subtract guaranteed gas when processing deposits individually.

The reason for subtracting out guaranteed gas at the start is to ensure that we have
enough gas for all deposits in the presence of additional gas.

### Additional Gas On L2

Additional gas is only bought given the following conditions:

- `gasPrice` >= the L2 gas price (however it is set)
- There is enough `additionalGas` remaining in the gas pool
- The account has enough balance to cover `gasPrice * additionalGas`

If `additionalGas` is bought, the `gasLimit` of the transaction is
`guaranteedGas + additionalGas` else it is `guaranteedGas`

When the transaction is done, if `gasUsed > guaranteedGas` a refund of
`(gasLimit - (gasUsed - guaranteedGas)) * gasPrice` is sent to
`from`. It is important to not refund any `guaranteedGas`
