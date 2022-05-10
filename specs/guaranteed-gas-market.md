# Guaranteed Gas Fee Market

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Gas Stipend](#gas-stipend)
- [Limiting Guaranteed Gas](#limiting-guaranteed-gas)
- [1559 Fee Market](#1559-fee-market)
- [Rationale for burning L1 Gas](#rationale-for-burning-l1-gas)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

[Deposited transaction](./glossary.md#deposited-transaction) are transactions on L2
that are initiated on L1. The gas that they use on L2 is bought on L1 via a gas burn
or a direct payment. We maintain a fee market and hard cap on the amount of gas provided
to all deposits in a single L1 block.

The gas provided to deposited transactions is sometimes called "guaranteed gas". The
gas provided to deposited transactions is unique in the regard that it is not
refundable. It cannot be refunded as it is sometimes paid for with a gas burn and
there may not be any ETH left to refund.

The **guaranteed gas** is composed of a gas stipend, and of any guaranteed gas the user
would like to purchase (on L1) on top of that.

## Gas Stipend

Because there is some cost to submitting the transaction and updating the basefee,
we provide transactions with a small amount of free gas.

If the user requests more `guaranteedGas` than the `gasStipend`, that gas will
be bought with L1 ETH via a gas burn or by buying it directly. If they request
less gas than the stipend, they will not be charged.

## Limiting Guaranteed Gas

The total amount of guaranteed gas that can be bought in a single L1 block must
be limited to prevent a denial of service attack against L2 as well as allow the
total amount of guaranteed gas to be below the L2 block gas limit.

We set limit the total amount of gas buyable via a contract method. It will initially
be controlled by the Optimism Multisig before being handed over to governance.
TODO - check that this is the actual plan.

## 1559 Fee Market

To reduce Priority Gas Auctions (PGAS - TODO LINK/DEFINITION) and accurately price gas, we implement a 1559
style fee market on L1 with the following pseudocode. We also use this opporunity to
place a hard limit on the amount of guaranteed gas that is provided.

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
    # Fun fact, geth doesn't let the new_basefee get below 0 and while not in the EIP spec, we should add this as well.
        new_basefee   := curr_basefee - basefee_delta
    }
    curr_basefee := new_basefee
    curr_number := block.number
    curr_bought_gas := 0
   
}

curr_bought_gas += required_gas
require(curr_bought_gas <= min(GUARANTEED_GAS_LIMIT, SANITY_GAS_LIMIT)
gas_cost = requested_gas * curr_basefee

burn(gas_cost) OR pay_to_contract(gas_cost) # Depends if payable or non-payable version

pack_and_store(curr_basefee, curr_number, curr_bought_gas)
```

TODO: Python pseudo-code

## Rationale for burning L1 Gas

If we collect ETH directly we need to add the payable selector. Some projects are not
able to do this. The alternative is to burn L1 gas. Unfortunately this is quite wastefull.
As such, we provide two options to buy L2 gas:

1. Burn L1 Gas
2. Send ETH to the Optimism Portal

The payable version (Option 2) will have a TODO discout applied to it (or conversly, #1 has a premium
applied to it).
