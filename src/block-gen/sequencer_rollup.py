import pickle
from typing import List, NewType
from block_gen import Block, Transaction, Event

# A more complex rollup that allows for privilaged block
# production. This is achieved by allowing a "sequencer"
# to squeeze blocks into historical L1 blocks after the fact.

# This expands on the "simplest possible rollup" by, in addition
# to blocks being generated based on the first event, the sequencer
# may submit transactions in later blocks which contain many L2 blocks.

#################### Types ####################

class SequencerBlock(Block):
    # An optional field which specifies the epoch that
    # this sequencer block is intended for.
    target_epoch: int

SequencerBatch = NewType("SequencerBatch", List[SequencerBlock])

#################### Functions ####################

# Generate the full rullup chain
def generate_rollup_chain(l1_chain: List[Block], sequencer_timeout) -> List[Block]:
    l2_chain: List[Block] = []
    for i in range(0, len(l1_chain) - sequencer_timeout):
        root_block = l1_chain[i]
        subsequent_blocks = l1_chain[i+1:i+1+sequencer_timeout]
        l2_blocks = generate_rollup_epoch(root_block, subsequent_blocks, sequencer_timeout)
        l2_chain += l2_blocks
    return l2_chain

# Generate a single epoch of the rollup chain. There is 1 epoch for every L1 block.
# Epochs have 1 deposit block and can have variable numbers of sequencer blocks.
# We must wait until the `sequencer_timeout` to determine an epoch's blocks.
def generate_rollup_epoch(
            root_block: Block,
            previous_blocks: List[Block],
            sequencer_timeout) -> List[Block]:
    assert len(previous_blocks) == min(sequencer_timeout, root_block["block_number"] - 1), \
        "Cannot generate epoch without all preceding blocks up to sequencer_timeout."
    deposit_block = generate_deposit_block(previous_blocks[0])
    l2_chain: List[Block] = [deposit_block]
    # Determine all sequencer blocks
    last_target_epoch = -1
    for block in previous_blocks:
        batch: SequencerBatch = extract_batch(block)
        if batch == None:
            continue
        for seq_block in batch:
            # Update the last_target_epoch
            if seq_block["target_epoch"] > last_target_epoch:
                last_target_epoch = seq_block["target_epoch"]
            # Ignore the block if it is targeting the wrong epoch
            if last_target_epoch != root_block["block_number"]:
                continue
            # We've found a sequencer block for this epoch so append it!
            l2_chain.append(seq_block)

    return l2_chain

def extract_batch(l1_block: Block) -> SequencerBatch:
    encoded_seq_batch = l1_block["txs"][0]["data"]
    try:
        seq_batch: SequencerBatch = pickle.loads(encoded_seq_batch)
    except:
        seq_batch: SequencerBatch = None
    return seq_batch

def generate_deposit_block(l1_block: Block) -> Block:
    deposits = l1_block["events"]
    events: List[Event] = []
    txs: List[Transaction] = deposits
    deposit_block: Block = {
        "block_hash": 'deposit blockhash' + str(l1_block["block_hash"]),
        "base_fee": 'deposit basefee' + str(l1_block["base_fee"]),
        "block_number": l1_block["block_number"],
        "timestamp": l1_block["timestamp"],
        "events": events,
        "txs": txs
    }
    return deposit_block