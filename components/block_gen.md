# Block Generation

The logic which is used to generate the rollup chain from L1.

## Glossary

- **Block Inputs**: All information in an Ethereum block required to generate the full block contents.
- **Rollup Blockchain**: The rollup blocks generated by performing the rollup's state transition function over all block inputs.

## Summary

The rollup chain can be deterministically generated given an L1 Ethereum chain. The fact that the entire rollup chain can be generated based on L1 blocks is _what makes OE a rollup_. This process can be represented as:

```python
f(l1_blockchain) -> rollup_blockchain
```

In this document we define a block generation function which is designed to be:

1. Minimal.
2. Support sequencers and sequencer consensus.
3. Resilient to the sequencer losing connection to L1.

## Simple Sequencer Block Input Generation
There are two types of blocks in the simple sequencer block generation algorithm:

1. Deposit block
2. Sequencer block

### Deposit Blocks
For every L1 block (after the deployment of the rollup) an L2 deposit block is created. These deposit blocks contain both `UserDeposit`s and `BlockDeposits`. User deposits are L1 user initiated actions which are carried out on L2, providing the rollup liveness guarantees. Block deposits deposit contextual information about the L1 block (eg. `blockhash` and `timestamp`).

```python
class Deposit:
    feedIndex: uint64
    GasLimit:  uint64

class UserDeposit(Deposit):
    isEOA:       bool
    l1TxOrigin:  Address
    target:      Address
    data:        bytes

class BlockDeposit(Deposit):
    blockHash:   bytes32
    blockNumber: uint64
    timestamp:   uint64
    baseFee:     uint64
```

Every time a deposit block is added to the rollup chain it marks the beginning of a new `epoch`. Each epoch contains **one** deposit block, and zero to many **sequencer blocks**.

### Sequencer Blocks
The sequencer is able to submit blocks which target a particular rollup epoch. They can submit up to `MAX_SEQUENCER_BLOCKS_PER_EPOCH` number of sequencer blocks every epoch. Additionally, they must assign a `target_epoch` to their blocks which satisfies:

```python
assert target_epoch > current_l1_block_number - sequencer_timeout, \
    "Sequencer must submit their blocks before the sequencer_timeout."
assert target_epoch < current_l1_block_number \
    "Sequencer cannot target future epochs."
```

### Epoch Block Input Generation
Each epoch's block inputs can be independently generated using the following function:

```python
# Generate a single epoch of the rollup chain. There is 1 epoch for every L1 block.
# Epochs have 1 deposit block and can have variable numbers of sequencer blocks.
# In the worst case you must wait until the `sequencer_timeout` to determine an
# epoch's blocks.
def generate_rollup_epoch(
            root_block: Block,
            subsequent_blocks: List[Block],
            sequencer_timeout) -> List[Block]:
    assert len(subsequent_blocks) >= sequencer_timeout, \
        "Cannot determine epoch blocks until sequencer timeout has passed"
    deposit_block = generate_deposit_block(root_block)
    l2_chain: List[Block] = [deposit_block]
    # Determine all sequencer blocks
    last_target_epoch = 0
    for block in subsequent_blocks:
        batch: SequencerBatch = extract_batch(block)
        if batch == None:
            continue
        for seq_block in batch:
            # Update the last_target_epoch
            if (seq_block["target_epoch"] > last_target_epoch and
                seq_block["target_epoch"] < block["block_number"]):
                last_target_epoch = seq_block["target_epoch"]
            # Ignore the block if it is targeting the wrong epoch
            if last_target_epoch != root_block["block_number"]:
                continue
            # We've found a sequencer block for this epoch so append it!
            l2_chain.append(seq_block)

    return l2_chain
```

After having generated each epoch it is possible to stich all epochs together to form the full rollup chain.
```python
# Generate the full rullup chain
def generate_rollup_chain(l1_chain: List[Block], sequencer_timeout) -> List[Block]:
    l2_chain: List[Block] = []
    for i in range(0, len(l1_chain) - sequencer_timeout):
        root_block = l1_chain[i]
        subsequent_blocks = l1_chain[i+1:i+1+sequencer_timeout]
        l2_blocks = generate_rollup_epoch(root_block, subsequent_blocks, sequencer_timeout)
        l2_chain += l2_blocks
    return l2_chain
```

## Simple Sequencer Full Block Generation
Now that we have all block inputs, it is possible to run the Ethereum state transition function to generate all remaining fields. Transforming block inputs into full rollup blocks requires processing transactions in the EVM, storing state, etc. Everything we are used to when running a node.

```python
def block_inputs_to_rollup_blocks(l2_chain: List[BlockInput]) -> RollupChain:
    chain = RollupChain()
    for block_input in l2_chain:
        chain.apply_block_input(block_input)
    return chain
```

For a full description of the rollup chain block processing function, see the [Execution Engine]("./exec_engine.md") section of the docs.

#### A Note on Fraud Proof Block Generation
Inside of the fraud proof we use the `generate_rollup_epoch(..)` function to narrow in on a single L2 block state transition as the first step of the fraud proof. After that we evaluate a single L2 block. (This is because the fraud proof witness generation is slow so we need to split it up).