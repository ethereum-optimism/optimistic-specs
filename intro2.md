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

Ethereum's limited resources, specifically bandwidth, computation, and storage, constrain the number of transactions which
can be processed on the network, leading to extremely high fees. Scaling
Ethereum means increasing the number of useful transactions the Ethereum network can process, by increasing the supply of these limited resources. This also means the fees can be made much lower.

You can [follow this link][rollup-scale] to learn more about how rollups help solve Ethereum scalability.

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

As a user, once you have [bridged] (\*) some ETH over to Optimistic Ethereum, you will
probably interact with it though your wallet, such as Metamask. Add the network
(chainID 10, [JSON-RPC] endpoint `https://mainnet.optimism.io/`) and you're good
to go.

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
"batch". The transactions in such batches will become *L2 sequencer block* once
posted to the Optimistic Ethereum contracts on L1. Please note that there is no
simple mapping between batches and sequencer blocks: a batch may "contain"
multiple blocks or even partial blocks. Batches are simply a means of grouping
transactions together for submitting transactions to L1.

*L2 deposit blocks*, on the other hand, arise from L1 blocks. There is one L2
deposit block per L1 block. These deposit blocks are implicitly created by the
sequencer whenever it posts a batch to a new L1 block.

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

Indicates the L2 block number, which is the number of the L2 block that this
transaction will end up in.

Blocks are numbered sequentially by the sequencer. Given the presence of both
sequencer and deposit blocks, some explanation on block numbering is in order.

At any given point, the sequencer knows the last block on which it has
successfully submitted a batch, which we'll refer to as `last_batch_L1_block`.

From this block, it is possible to derive the highest L2 block number associated
with a L2 transaction posted on L1. We'll call this
`last_batch_sequencer_L2_block_number`.

Assuming the sequencer just submitted this batch, it can then proceed to confirm
transactions, assigning them a block number B such that`B >=
last_batch_L2_block_number`.

Just like batches, L1 blocks can contain multiple L2 blocks, or even partial
blocks. They can also have received zero, one, or many batches from the
sequencer.

Periodically, the sequencer sends a batch of transactions to L1. The sequencer
may only delay sending a batch for up to `S` blocks, where `S` is the
*sequencing window*, also known as *sequencer block submission window* and
*force inclusion window*.

If the sequencer does not post any batch for `S` blocks, anybody can force the
creation of the deposit block implied by the L1 block at `C - S` where `C` is
the current L1 block number.

**TODO:** "anybody can force" — is this accurate or does this also happen
implicitly in the L2 block feed?

But the sequencing window serves another role, related to the notion of *epoch*.

The L2 block stream is divided into epochs, which each contain a deposit block,
followed by zero or more sequencer blocks. The deposit block for epoch `E` is
the one generated by the L1 block with number `E`.

Within the sequencing window `S`, the sequencer is allowed to add sequencer
blocks to the epoch. This means that blocks in the inclusive range `[E+1, E+S]`
are all susceptible to contain sequencer blocks for epoch `E`.

**TODO:** what do we pick as the actual sequencing window?

Clearly, the sequencing window for neighbouring epochs overlap. The sequencer is
free to skip to the next epoch `E` at any time during in its sequencing window
(`[E+1, E+S]`). It does so by either submitting a batch containing at least one
transaction whose block number is `>= X+2`, where `X` is the L2 block number of
the previous sequencer block. The deposit block for `E` will have number `X+1`.
If there is at least one transaction with block number `X+2`, it belongs to the
first sequencer block for the epoch. If such a transaction does not exist, the
block with number `X+2` is another deposit block (for epoch `E+1`), and the
epoch `E` does not have have any sequencer block.

An important consequence of this is that validators can see deposit blocks long
before their inclusion in the L2 chain, but do not know their block number and
cannot execute them, as yet unseen sequencer blocks may come first. Only when
the sequencer "skips over" the deposit block (as explained in the previous
paragraph) can the deposit block be considered part of the chain.

If the sequencer fails to skip over the deposit block within its sequencing
window (and not submitting any batch for the duration of the window is merely a
special case of this), then the deposit block will be forcefully included after
the last L2 block seen on chain. This block will necessarily appear within the
L1 block range `[E, E+S]` (for epoch `E` and sequencing window size `S`), as if
this range does not include any sequencer block, then it will at least contain
the forced deposit block for epoch `E-1`. Also note that the last L2 block can
still be this forced `E-1` deposit block, even if `[E, E+S]` is not empty.

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
