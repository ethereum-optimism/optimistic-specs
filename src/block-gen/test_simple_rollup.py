import copy
import pickle
from block_gen import Block, Event, gen_dummy_block
from typing import List
from simple_rollup import generate_rollup

def test_generate_rollup():
    l1_chain: List[Block] = []
    expected_l2_chain: List[Block] = []

    for i in range(100):
        block = gen_dummy_block(i)
        expected_l2_chain.append(copy.deepcopy(block))
        l2_block_encoded = pickle.dumps(block)
        block["events"][0]["data"] = l2_block_encoded
        l1_chain.append(block)
    l2_chain = generate_rollup(l1_chain)

    for i in range(len(expected_l2_chain)):
        assert l2_chain[i]["timestamp"] == expected_l2_chain[i]["timestamp"]

    # Woot! We pulled the rollup out of the L1 chain!