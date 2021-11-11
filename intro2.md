# The Optimistic Ethereum Specification Overview

Optimistic Ethereum is an _EVM equivalent_, _optimistic rollup_ protocol
designed to _scale Ethereum_ while remaining maximally compatible with existing
Ethereum infrastructure. This document provides an overview of the protocol to
provide context for the rest of the specification.

## Table of Contents

1. [Foundations](#foundations)
1. [Network Participants](#network-participants)
1. [User Transactions and The Sequencer](#user-transactions-and-the-sequencer)
1. [L2 Batches and Blocks](#l2-batches-and-blocks)
1. [L2 Block Properties](#l2-block-properties)
1. [L1 Components Overview](#l1-components-overview)
   - [Data Feeds](#data-feeds)
   - [Fraud Proof Manager](#fraud-proof-manager)

## Foundations

### What is Ethereum scalability?

Ethereum's limited resources, specifically bandwidth, computation, and storage,
constrain the number of transactions which can be processed on the network,
leading to extremely high fees. Scaling Ethereum means increasing the number of
useful transactions the Ethereum network can process, by increasing the supply
of these limited resources. This also means the fees can be made much lower.

You can [follow this link][rollup-scale] to learn more about how rollups help
solve Ethereum scalability.

[rollup-scale]: https://hackmd.io/@norswap/rollups

### What is Optimistic Rollup?

[Optimistic rollup](https://vitalik.ca/general/2021/01/05/rollup.html) is a
layer 2 scalability technique which increases the computation & storage capacity
of Ethereum without sacrificing security or decentralization. Transaction data
is submitted on-chain but executed off-chain. If there is an error in the
off-chain execution, a fraud proof can be submitted on-chain to correct the
error and protect user funds. In the same way you don't go to court unless there
is a dispute, you don't execute transactions on on-chain unless there is an
error. The rollup is *optimistic* because the execution results are assumed to
be correct until a fraud proof proves them wrong.

### What is EVM Equivalence?

[EVM
Equivalence](https://medium.com/ethereum-optimism/introducing-evm-equivalence-5c2021deb306)
is complete compliance with the state transition function described in the
Ethereum yellow paper, the formal definition of the protocol. By conforming to
the Ethereum standard across EVM equivalent rollups, smart contract developers
can write once and deploy anywhere.

### 🎶 All together now 🎶

#### Optimistic Ethereum is an _EVM equivalent_, _optimistic rollup_ protocol designed to _scale Ethereum_.

## Network Participants

There are three actors in Optimistic Ethereum: users, sequencers, and verifiers.

<!--
![Network Overview](./assets/network-participants-overview.svg)
-->

![Network Overview](https://raw.githubusercontent.com/ethereum-optimism/optimistic-specs/main/assets/network-participants-overview.svg)

### Users

At the heart of the network are users (us!). Users can:

1. Deposit or withdraw tokens by sending transactions to Ethereum mainnet.
2. Send transactions to the sequencer, to use EVM smart contracts on L2, or to
   send ETH to other L2 users.
3. View the status of transactions using block explorers provided by network
   verifiers.

### Sequencer

As a first approximation, you can see the sequencer as the L2 block producer. It
accrues user transactions into "batches" and submits these batches to contracts
on L1 in order to make L2 blocks out of them (beware however, that there isn't a
1-1 mapping between batches and blocks). Unlike on L1, on L2 transactions can be
confirmed before a full block (or even a full batch) is created!

We are going to give a lot more detail about the operation of the sequencer in
the rest of this document.

### Verifiers

Verifiers monitor L1 for rollup data. They serve three purpose:

1. Serving rollup data to users; and
2. Verifying rollup integrity and disputing invalid assertions.
3. Propagate the L2 state among validators (\*)

In order for the network to remain secure there must be at least one honest
verifier who is able to verify the integrity of the rollup chain & serve
blockchain data to users.

(\*) It's important to note that a validator does not rely on this data to
validate the rollup — only access to the L1 chain is required. However, getting
the most recent L2 state allows validators to serve the state to users, and to
pre-validate transactions that should soon be posted to L1. L2 state propagation
uses the same mechanism as L1 state sync, but must additionally include a
signature from the sequencer to ensure the legitimacy of the data.

## User Transactions and The Sequencer

As a user, once you have [bridged] (\*) some ETH over to Optimistic Ethereum,
you will probably interact with it though your wallet, such as Metamask. Add the
network (chainID 10, [JSON-RPC] endpoint `https://mainnet.optimism.io/`) and
you're good to go.

(\*) L1 -> L2 deposits will be explained later.

[bridged]: https://www.optimism.io/apps/bridges
[JSON-RPC]: https://github.com/ethereum/execution-apis

When you send a transaction, it is sent to the node called the sequencer. The
sequencer will run verify your transaction, execute it if valid, and confirm it.
This happens much faster than on Ethereum mainnet (around ~1s).

There is currently a single sequencer operated by Optimism PBC. We expect future
versions of this specification to introduce sequencer decentralization. However,
despite the sequencer being decentralized, trust is not required. As we will
see, the sequencer posts its results on Ethereum mainnet (henceforth: layer 1 or
L1), where they can be permissionlessly challenged with a fraud proof. If the
sequencer temporarily goes down, Optimism will remain live because users are
able to submit L2 transactions on L1 (of course, this implies the loss of the
increased throughput and lower fees).

L2 transactions are identical to L1 transactions, except that they use the
Optimistic L2 state instead of the L1 state. This includes account balances,
deployed contracts and contract storage.

## L2 Batches and Blocks

Optimistic Ethereum has blocks, however these work slightly differently from L1
blocks, which are mined by proof-of-work and will soon be decided by
proof-of-stake.

Optimistic Ethereum has two kind of blocks:
1. L2 deposit blocks
2. L2 sequencer blocks

As it confirms transactions, the sequencer accumulates these transaction in a
"batch". The transactions in such batches will become **_L2 sequencer block_**
once posted to the Optimistic Ethereum contracts on L1. Please note that there
is no simple mapping between batches and sequencer blocks: a batch may "contain"
multiple blocks or even partial blocks. Batches are simply a means of grouping
transactions together for submitting transactions to L1.

The sequencer decides which L@ sequencer block a transaction belongs to,
although this decision is constrained by the protocol. For instance, block
numbers assigned to transactions increase monotonically. See the section [on the
`NUMBER` opcode](#NUMBER) for more details.

Just like batches, L1 blocks can contain multiple L2 blocks, or even partial
blocks. They can also have received zero, one, or many batches from the
sequencer.

**_L2 deposit blocks_**, on the other hand, arise from L1 blocks. There is one
L2 deposit block per L1 block. These deposit blocks are implicitly created by
the sequencer whenever it posts a batch to a new L1 block.

**TODO:** Implicitly? Needs a link to an explanation of how data feeds are materialized.

L2 deposit blocks comprise two types of data:
1. L1 block properties, namely
   - block hash
   - block number
   - timestamp
   - base fee
2. L2 transactions submitted on L1

These L2 transactions that have been "enqueued" on L1 are called *deposits*.

These transactions have two main uses:
- Ensure that the rollup remains live even if the sequencer goes down or starts
  censoring L2 transactions. Consequently, funds can never remain stuck on the
  rollup.
- Deposit ETH and ERC-20 tokens onto the rollup. This is achieved by sending the
  token to a contract which locks it, then submits a L2 transaction (on L1)
  instructing a L2 contract to mint an equivalent L2 token. The transaction
  records the address of the L1 contract that submitted it, and the L2 contract
  only allows minting if the transaction was submitted by an authorized L1
  contract.

  Anybody can operate such a pair of contracts, but Optimism PBC operates the
  bridge for ETH and some blue chip tokens ([the gateway]). See also other
  [trusted bridge providers].

[the gateway]: https://gateway.optimism.io/
[trusted bridge providers]: https://www.optimism.io/apps/bridges

## L2 Block Properties

Because Optimistic Ethereum is 100% EVM-equivalent, all EVM opcodes are
available. However, because blocks work differently, the semantics of a few
opcodes must be made precise, namely `TIMESTAMP`, `NUMBER`, `BLOCKHASH`,
`BASEFEE`, `GASLIMIT` and `COINBASE`.

#### `NUMBER`

The sequencer numbers L2 blocks sequentially. If there were only sequencer
blocks, this would be trivial. The fact that the sequencer must "insert" deposit
blocks in the L2 block stream makes this a little bit more complicated.

To "insert" a deposit block in the block stream, the sequencer *skips over it*
by skipping a block number. For instance, if a batch contains a transaction
whose assigned block number is `X`, followed by a transaction whose assigned
block number is `X+2`, this indicates that:

- The first transacton is the final transaction of the L2 sequencer block with
  number `X`.
- The second transaction is the first transaction of the L2 sequencer block with
  number `X+2`.
- There is a deposit block with number `X+1`, which is the oldest deposit block
  that hasn't yet been included in the L2 block stream.

**TODO:** is this true? or does the sequencer signal the insertion point for deposit blocks more explicitly

To see the kind of problem that can occur, consider the naive procedure where
the smart contract handling new batches enforces that the batch "skips over"
every deposit block associated with L1 blocks coming before `B`, and otherwise
rejects the batch. There are a few issues with this. This protocol is both too
strict and too loose.

First, the procedure is too strict: the sequencer doesn't directly control which
block the batch will be posted on. This makes it very easy for the batch to land
on an L1 block whose parent was unknown to the sequencer. This could occur for a
variety of reason, including poor network conditions and malicious L1 block
withholding. This could lead to a lot of wasted batch submission transactions.

Second, the procedure is not strict enough: under this protocol, the sequencer
is allowed to infinitely postpone the inclusion of deposit blocks, as long as it
also does not post a batch to L1.

To solve these issues, we introduce for each L1 block a *sequencing window* of
size `S > 1` during which the sequencer **must** include the deposit block.

The *sequencing window* is also called the *sequencer block submission window*
(for obvious reasons), but also the *force inclusion period*. This is because if
the sequence fails to include the deposit block generated by the L1 block `B`
within the sequencing window `[B+1, B+S]`, then validators will consider that
deposit block to be forcefully included, and will assign it the block number
`X+1` where `X` is the last known L2 block.

It's interesting to consider what this last block `X` might be. It might be the
highest-number L2 sequencer block for which a transaction was posted by the
sequencer within the `[B,B+S]` L1 block range (not a typo, this is `S+1` sized
range). However, it could also be another force-included deposit block. For
instance, if the `[B,B+S-1]` L1 block range does not include the deposit block
generated by the `B-1` L1 block (via a sequencer block skipping over it), then
the previous L2 block `X` is the forcefully included deposit block for L1 block
`B-1`.

In any case, it is sufficient to look a the `[B,B+S]` L1 block range to
determine the block number (and the execution result) of the deposit block
generated by the L1 block `B`.

Also note that the sequencer not submitting any batches during the sequencing
window `[B,B+S]` is merely a special case of the force inclusion mechanism,
which ensures that the L2 chain remains live, even if the sequencer goes down.
Even if the sequencer could never go down, the force inclusion mechanism is
required to prevent the sequencer from censoring L2 transactions, which can now
be forcefully included via L1.

If the sequencer violates these conditions, for instance if it keeps posting
batches after the L1 block `B+S` without skipping a block number for the forced
inclusion of the deposit block, then this is considered to be fraud, and can be
proved as such via a [fraud proof].

[fraud proof]: TODO

**TODO:** link fraud proof index page

Finally, a bit of jargon. We call an *epoch* the sequence of L2 block starting
with a deposit block and containing zero or more L2 sequencing block until the
next deposit block, which marks the start of the next epoch. Epoch `E` is the
epoch staring with the deposit block generated by the L1 block with number `E`.

**TODO:** what do we pick as the actual sequencing window?

#### `TIMESTAMP`

The timestamp should be **approximately** equivalent to the time at which the
block was "conceived".

- For deposit blocks, this is the timestamp of the L1 block it belongs to.
- For sequencer blocks, this is (approximately) the time at which the decided to
  assign newly received transactions to a new sequencer block

**TODO:** I flat out made this up. How is it supposed to work?

Optimistic Ethereum makes the following guarantees when it comes to block
timestamps:

- The timestamp of each block is higher or equal to that of the block that
  precedes it.

**TODO:** is that so?

#### `BLOCKHASH`

**TODO:** how does this work? Obviously can't be a block hash because we run and
confirm transactions before knowing all the transactions in the block (this
section should mention this)

#### `BASEFEE`

**TODO**

#### `GASLIMIT`

**TODO:** numbers

#### `COINBASE`

**TODO:** address of the sequencer?

## L1 Components

**TODO:** design ongoing

Before digging further into the operation of the sequencer and the verifier,
let's give an overview of the Optimistic Ethereum infrastructure on L1.

### Data Feeds

First we have **data feeds**. Conceptually, a data feed is an append-only log of
a certain kind of data. Smarts contracts may be used to help implement these
feeds, but the only requirement is that they these feeds can be
deterministically retrieved from the L1 state and L1 block data.

**TODO:** The [overview] says "retrievable in a bounded number of steps" but I'm
not sure if the precision is useful - what would a counter-example be? Also, as
the L1 chain grows, the number of operations to retrieve a log entry naturally
increases, so I'm not sure we can talk of bounds in the first place.

The **deposit feed** logs all L2 transaction submitted on L1, as well as all
L1->L2 deposits submitted on L1.

The **sequencer feed** logs all L2 transactions which are submitted (in batches)
by the sequencer to L1. These transactions are ordered, and have associated
timestamp and block information.

**TODO:** I added the part about block information, that's correct right?

The **L2 block feed** logs all L2 blocks inputs. Block inputs comprise
transactions submitted on L1 (with associated block information), as well as L1
block properties from deposit blocks, but excludes ["block header items"] such
as the Merkle roots for the state, receipts, gas information, ...

**TODO:** what's the relationship between the L2 block feed and the depsoit feed & sequencer feed? it just seems like an ordered merge of boths?

["block header items"]: https://github.com/norswap/nanoeth/blob/cdc9867ed553847f3b0b7787fd03f0ed091c6b7f/src/com/norswap/nanoeth/blocks/BlockHeader.java#L22-L156

**TODO:** I expanded the description compared to the [overview],
confirm that it's right

[overview]: https://github.com/ethereum-optimism/optimistic-specs/blob/main/overview.md

#### Fraud Proof Manager

**TODO:** I'm not a fan of the "L2 block oracle" term - what is it an oracle
for? You'd never guess what it does given just the name. What about "Fraud proof
manager" instead?

The fraud proof manager has two responsibilities:

1. It records *proposals*, which are summarized outcomes of the execution of an
   L2 block — in particular, the Merkle root of the L2 state and the Merkle root
   of receipts (yellowpaper §4.3.1) resulting from the block's execution.
2. It *finalizes* L2 blocks after the dispute period (currently 7 days) has
   passed, after which the L2 blocks can no longer be challenged.

**TODO:** things to add to the definition of "proposal"?

The fraud proof manager has the following components:

The **proposal manager** stores proposals submitted both by proposers (the
sequencer and verifiers who wish to challenge another proposer's proposal). It
also ensure that proposers are sufficiently bonded.

Each proposer must lock up a bond that will be forfeit if their proposal is
proved to be invalid by the operation of the fraud proof manager, in which case
the bond will be given to the challenger.

**TODO:** this will be fleshed out when explaining the dispute game, but it
should be made clearer what can be challenged (i.e. not just the sequencer - a
challenger can make a wrong challenge and forfeit its bond too)

The **the k-section game manager**

TODO

- [Dispute game specification(https://statechannels.notion.site/Draft-dispute-game-specification-2eee37cd8cc943759405a9ef97885411)
- [Dispute game contract](https://github.com/statechannels/dispute-game/blob/main/sol-prototype/dispute-manager.sol)

**TODO:** decide k

--------------------------------------------------------------------------------
