import random
import pickle
from block_gen import Block, Event, gen_dummy_block, gen_dummy_block_with_deposit
from typing import List
from sequencer_rollup import (
    generate_deposit_block,
    generate_rollup_chain,
    generate_rollup_epoch,
    extract_batch,
    SequencerBatch,
    SequencerBlock
)

def test_extract_batch_with_only_deposit():
    block = gen_dummy_block(1)
    block["events"][0]["data"] = "I am a deposit!"
    batch: SequencerBatch = extract_batch(block)
    assert batch is None


def test_extract_batch_with_single_sequencer_block():
    sequencer_block = gen_dummy_sequencer_block(0)
    block = gen_dummy_block(0)
    block["events"][0]["data"] = "I am a deposit!"
    expected_batch: SequencerBatch = [sequencer_block]
    block["txs"][0]["data"] = pickle.dumps(expected_batch)
    got_batch: SequencerBatch = extract_batch(block)
    assert expected_batch == got_batch


def test_extract_batch_with_single_sequencer_block():
    sequencer_block = gen_dummy_sequencer_block(0)
    block = gen_dummy_block_with_deposit(0)
    expected_batch: SequencerBatch = [sequencer_block]
    block["txs"][0]["data"] = pickle.dumps(expected_batch)
    got_batch: SequencerBatch = extract_batch(block)
    assert expected_batch == got_batch


def test_generate_rollup_epoch_with_single_sequencer_block():
    # Create the start block of the epoch
    start_block = gen_dummy_block_with_deposit(0)
    # Next create a sequencer block and put it in the root block
    sequencer_block = gen_dummy_sequencer_block(1)
    root_block = gen_dummy_block(1)
    batch: SequencerBatch = [sequencer_block]
    root_block["txs"][0]["data"] = pickle.dumps(batch)
    # Finally let's generate the l2 epoch which should contain a deposit & sequencer block
    got_epoch = generate_rollup_epoch([start_block, root_block])
    # Check it against what we expect
    deposit_block = generate_deposit_block(start_block)
    expected_epoch = [deposit_block, sequencer_block]
    assert got_epoch == expected_epoch


def test_generate_rollup_chain():
    l1_chain: List[Block] = []
    l2_chain: List[Block] = []
    pending_batch: SequencerBatch = []
    # sequencer_timeout: int = 10
    sequencer_timeout: int = 2
    num_l1_blocks: int = 2
    for i in range(num_l1_blocks):
        block = gen_dummy_block_with_deposit(i)
        if len(pending_batch) != 0:
            block["txs"][0]["data"] = pickle.dumps(pending_batch)
        l1_chain.append(block)
        # Append a deposit block for the newly added block
        l2_chain.append(generate_deposit_block(block))
        # Generate a pending batch that we will append to the next block
        pending_batch = make_pending_batch(block, sequencer_timeout)
        # pending_batch = make_random_pending_batch(block, l2_chain, sequencer_timeout)
        l2_chain += pending_batch
    # Generate the whole rollup chain
    got_l2_chain = generate_rollup_chain(l1_chain, sequencer_timeout)
    expected_l2_chain: List[Block] = []
    for rollup_block in l2_chain:
        if rollup_block["block_number"] > l1_chain[-1]["block_number"]:
            break
        expected_l2_chain.append(rollup_block)
    # Now that we've filtered out the un-finalized l2 blocks, let's make
    # sure that we generated the rollup we expected!
    assert expected_l2_chain == got_l2_chain


######### Helpers #########

def make_pending_batch(
            latest_l1_block: Block,
            sequencer_timeout: int
            ) -> SequencerBatch:
    epoch = latest_l1_block["block_number"] + sequencer_timeout // 2
    sequencer_block = gen_dummy_sequencer_block(epoch)
    return [sequencer_block]

def make_random_pending_batch(
            latest_l1_block: Block,
            l2_chain: List[Block],
            sequencer_timeout: int
            ) -> SequencerBatch:
    latest_l2_block_num = l2_chain[-1]["block_number"]
    # The epoch can not be earlier than:
    # 1) The latest epoch we referenced in L2; or
    # 2) The **next** block after the latest block on L1
    smallest_possible_epoch = max(latest_l2_block_num, latest_l1_block["block_number"] + 1)
    largest_possible_epoch = latest_l1_block["block_number"] + sequencer_timeout
    # We'll have some fun and make the sequencer blocks random.
    epoch = random.randint(smallest_possible_epoch, largest_possible_epoch)
    sequencer_block = gen_dummy_sequencer_block(epoch)
    return [sequencer_block]

def gen_dummy_sequencer_block(target_epoch: int) -> SequencerBlock:
    # target_epoch = l1_block_number
    events: List[Event] = [{ "data": "seq event 1 data" }, { "data": "seq event 2 data" }]
    txs: List[Event] = [{ "data": "seq tx 1 data" }]
    block: SequencerBlock = {
        "target_epoch": target_epoch,
        "block_hash": 'blockhash' + str(target_epoch),
        "base_fee": 'basefee' + str(target_epoch),
        "block_number": target_epoch,
        "timestamp": target_epoch,
        "events": events,
        "txs": txs
    }
    return block