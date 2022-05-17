# Guaranteed Gas Fee Market

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Gas Stipend](#gas-stipend)
- [Limiting Guaranteed Gas](#limiting-guaranteed-gas)
- [1559 Fee Market](#1559-fee-market)
  - [Exponent Based Fee Reduction](#exponent-based-fee-reduction)
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

Guaranteed gas on L2 is bought in the following manner. An L2 gas price is calculated
via a 1559 style gas market. The total amount of ETH required to buy that gas is then
calculated (`guaranteed gas * L2 deposit basefee`). The contract then accepts that
amount of ETH (in a future upgrade) or (only method right now), burns an amount of
L1 gas that corresponds to the L2 cost (`L2 cost / L1 Basefee`).

## Gas Stipend

To offset the gas spent on the deposit event, we credit (TODO AMOUNT) gas times the
current basefee to the cost of the L2 gas. The amount of gas is selected to represent
the cost to the user. If the ETH price of the gas (gas times current L1 baseefee) is
greater than the requested guaranteed gas times the L2 gas price, no L1 gas is burnt.

## Limiting Guaranteed Gas

The total amount of guaranteed gas that can be bought in a single L1 block must
be limited to prevent a denial of service attack against L2 as well as allow the
total amount of guaranteed gas to be below the L2 block gas limit.

We set a guaranteed gas limit of 2,500,000 gas per L1 block. This corresponds to the
L2 gas target that we expect.

## 1559 Fee Market

To reduce [Priority Gas Auctions](./glossary.md#priority-gas-auction) and accurately price gas,
we implement a 1559 style fee market on L1 with the following pseudocode. We also use this
opporunity to place a hard limit on the amount of guaranteed gas that is provided.

```python
# Pseudocode to update the L2 Deposit Basefee and cap the amount of guaranteed gas
# bought in a block. Calling code must handle the gas burn and validity checks on
# the ability of the account to afford this gas.
BASE_FEE_MAX_CHANGE_DENOMINATOR = 8
ELASTICITY_MULTIPLIER = 2
GUARANTEED_GAS_LIMIT = 2,500,000
GAS_TARGET = GUARANTEED_GAS_LIMIT / ELASTICITY_MULTIPLIER
    
# All values are uint64s
prev_basefee, prev_num, prev_bought_gas = load_and_unpack_storage()
now_num = block.number

# Clamp the full basefee to a specific range. The minimum value in the range should be around 100-1000
# to enable faster responses in the basefee. This replaces the `max` mechanism in the ethereum 1559
# implementation (it also serves to enable the basefee to increase if it is very small).
def clamp(v: i256, min: u64, max: u64) -> u64:
    if v < i256(min):
        return min
    elif v > i256(max):
        return max
    else:
        return u64(v)


if prev_num == now_num:
    now_basefee = prev_basefee
    now_bought_gas = prev_bought_gas + requested_gas
elif prev_num == now_num + 1:
    # New formula
    # Width extension and conversion to signed integer math
    gas_used_delta = int128(prev_bought_gas) - int128(GAS_TARGET)
    # Use truncating (round to 0) division - solidity's default.
    # Sign extend gas_used_delta & prev_basefee to 256 bits to avoid overflows here.
    base_fee_per_gas_delta = prev_basefee * gas_used_delta / GAS_TARGET / BASE_FEE_MAX_CHANGE_DENOMINATOR
    now_basefee_wide = prev_basefee + base_fee_per_gas_delta

    now_basefee = clamp(now_basefee_wide, min=1000, max=UINT_64_MAX_VALUE)
    now_bought_gas =  requested_gas
else:
    # Skipped multiple blocks. Use an approximation to do constant time gas updating
    n = now_num - prev_num
    # Apply 7/8 reduction to prev_basefee for the n empty blocks in a row.
    base_fee_per_gas_delta = prev_basefee * 7**n / 8**n
    now_basefee_wide = prev_basefee + base_fee_per_gas_delta

    now_basefee = clamp(now_basefee_wide, min=1000, max=UINT_64_MAX_VALUE)
    now_bought_gas =  requested_gas

require(now_bought_gas < GUARANTEED_GAS_LIMIT)


pack_and_store(now_basefee, now_num, now_bought_gas)
```

### Exponent Based Fee Reduction

When there are stretches where no deposits are executed on L1, the basefee should be decaying, but is not.
If there is the case that the basefee spiked, this mechanism is needed to enable a more accurate decay. It
uses exponentiation to run in constant (relative to the number of missed blocks) gas.

With the current elasticty and change denominator values, if `n` is greater than 2^32, the exponention will overflow.

## Rationale for burning L1 Gas

If we collect ETH directly we need to add the payable selector. Some projects are not
able to do this. The alternative is to burn L1 gas. Unfortunately this is quite wastefull.
As such, we provide two options to buy L2 gas:

1. Burn L1 Gas
2. Send ETH to the Optimism Portal

The payable version (Option 2) will likely have discout applied to it (or conversly, #1 has a premium
applied to it).

For the initial release of bedrock, only #1 is supported.
