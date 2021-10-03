import pickle
from typing import List
from block_gen import Block

# The simplest possible rollup where the rollup blocks are
# encoded as the first event of every L1 block.

def generate_rollup(l1_chain: List[Block]) -> List[Block]:
    l2_chain: List[Block] = []
    for block in l1_chain:
        l2_block = generate_rollup_block(block)
        l2_chain.append(l2_block)
    return l2_chain

# This function is used in the fraud proof. It allows for an
# easy stateless transformation that determines the L2 blocks
# at a particular L1 block number.
def generate_rollup_block(l1_block: Block) -> Block:
    first_tx = l1_block["events"][0]["data"]
    l2_block = pickle.loads(first_tx)
    return l2_block