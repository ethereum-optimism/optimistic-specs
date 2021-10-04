# Fraud Proof Algorithm for Simple Sequencer
This is a WIP -- mostly just a few notes that I'm jotting down to not forget.

## Requirements
1. Not all blockhashes are posted up front.
    - Blockhashes should be posted roughly every 100 L1 blocks.
2. Only one block is evaluated during the fraud proof.

## Fraud Proof Structure

Fraud proofs are initialized with some constant size input. In our case this is:

```python
def init_fraud_proof(latest_blockhash: bytes, invalid_epoch_commitment: int):
    ''' Invalid epoch commitment is the index of the invalid blockhash '''
    # ...
```

What makes rollups special is that all of the preimages to the `latest_blockhash` are known. This means that from within the fraud proof VM, all blockhash preimages (including the state root, pre blockhash, etc) can be queried using the `preimage oracle`. Fundamentally this means that all historical blocks, events, transactions, and state can are accessible within the fraud proof VM.

### Preimage Oracle

The fraud proof VM has access to the following function (called the "preimage oracle"):

```python
def get_preimage(hash: bytes) -> bytes:
    # The preimage is looked up from a local database held by the fraud prover
    preimage = fetch_preimage(hash)
    assert keccak256(preimage) == hash
    return preimage
```

This seemingly simple functionality has extreme consequences. All publically available preimages can be operated on from within the fraud proof -- a pretty amazing property.

### Step 1: Make available the blockhash disagreement
- Requirement (1) reminder: the proposer does not submit all intermediate blockhashes.
- Requirement (2) reminder: The fraud proof VM must only execute a single blockhash.

The only way for us to satisfy these two requirements is to break our fraud proof into two distinct steps. 1) make all the blockhashes that were skipped available, and then 2) use the blockhashes & other Ethereum chain data inside the fraud proof VM to determine the invalid blockhash and execute the transition.

So for step 1 all we need to do is force the proposer to post and commit to all intermediate blockhashes.

### Step 2: Determine first invalid execution step

The fraud proof function is executed in a Fraud Proof VM with access to the `preimage oracle`. The function is roughly:

```python
# The fraud proof function determines whether or not the supplied blockhash is
# incorrect. This is not executed on chain but instead inside of the Fraud Proof VM
# and then a single step of the fraud proof VM is executed on chain.
def is_invalid_blockhash(latest_l1_blockhash: str, proposed_l2_blockhash: int)->bool:
    # Function is roughly as follows:
    #
    # Use the preimage oracle + `lastest_l1_blockhash` to search historical L1 blocks
    # and pull out the following information:
    #
    # 1. The previous L2 blockhash right before the potentially invalid `proposed_l2_blockhash`.
    # 2. All L1 blocks from the `root_block` to `root_block+sequencer_timeout`
    #    (for info on what the root block is see the block generation section.)
    #
    # With all of this information, generate all L2 blocks for the epoch. And then evaluate
    # the state transition between proposed_l2_blockhash-1 -> proposed_l2_blockhash
    pass
```