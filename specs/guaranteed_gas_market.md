# Guaranteed Gas Fee Market

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Gas Stipend](#gas-stipend)
- [1559 Fee Market](#1559-fee-market)
- [Rationale for burning L1 Gas](#rationale-for-burning-l1-gas)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

Deposited transactions (TODO: LINK) are transactions on L2 that are initiated on L1. The
gas that they use on L2 is bought on L1 via a gas burn or a direct payment. We maintain
a fee market and hard cap on the amount of gas provided to all deposits in a single L1
block.

The gas provided to deposited transactions is sometimes called "guaranteed gas". The
gas provided to deposited transactions is unqiue in the regards that it is not
refundable as it is sometimes paid for with a gas burn and there may not be any
ETH to refund to it.

The **guaranteed gas** is composed of a gas stipend, and of any guaranteed gas the user
would like to purchase (on L1) on top of that.

## Gas Stipend

Because there is some cost to submitting the transaction and updating the basefee,
we provide transactions with a small amount of free gas.

If the user requests more `guaranteedGas` than the `gasStipend`, that gas will
be bought with L1 ETH via a gas burn or by buying it directly.
It is not refundable if the transaction uses less gas than the gas limit.

Can also provide a stipend to the guaranteed gas of every transaction.
The stipend is bought with gas that was spent executing the deposit logic;
however, the stipend still needs to be included in the sum of bought
guaranteed gas.

TODO: How much / do we actually need this?

If a gas stipend is provided, the user is only required to buy the amount of
guaranteed gas in excess of the gas stipend.

## 1559 Fee Market

When

The deposit feed contract must limit the total amount of guaranteed gas in a
single L1 block. This is to limit the amount of gas that is used on L2.

This is done with a hard limit on the amount of guaranteed gas that can be
bought. To reduce PGAs (TODO: LINK TO DEFINITION) if the deposit mechanism is congested, we also implement
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

burn(gas_cost) # Via gas or ETH.

pack_and_store(curr_basefee, curr_number, curr_bought_gas)
```

TODO: Python pseudo-code

## Rationale for burning L1 Gas

If we burn ETH (or collect it), we need to add the payable selector everywhere.
Adding it everywhere is not feasible and really bad UX.
We will have a payable version and offere a discout against the gas burning version.
