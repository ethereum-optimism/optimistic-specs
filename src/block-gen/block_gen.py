from typing import List

class Event:
    data: str

class Transaction:
    data: str

class Block:
    block_hash: str
    base_fee: int
    block_number: int
    timestamp: int
    events: List[Event]
    txs: List[Transaction]

def gen_dummy_block(i: int) -> Block:
    events: List[Event] = [{ "data": "event 1 data" }, { "data": "event 2 data" }]
    txs: List[Transaction] = [{ "data": "tx 1 data" }]
    block: Block = {
        "block_hash": 'blockhash' + str(i),
        "base_fee": 'basefee' + str(i),
        "block_number": i,
        "timestamp": i,
        "events": events,
        "txs": txs
    }
    return block

def gen_dummy_block_with_deposit(i: int) -> Block:
    block = gen_dummy_block(i)
    block["events"][0]["data"] = "I am a deposit!"
    return block