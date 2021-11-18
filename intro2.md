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
can write once and deploy anywhere — just change the chain ID to 42 and you're
good to go.

Because L2 is implemented a little bit differently, there are a few differences
to be aware of. These differences are very small — for instance, transactions in
the same L2 block could have different timestamp, unlike on L1 (which is
actually an improvement!). You can find an exhaustive list of these differences
on [this page][L1L2Diff].

**TODO:** is that example real?

[L1L2Diff]: ./l1-l2-diff.md

### 🎶 All together now 🎶

#### Optimistic Ethereum is an _EVM equivalent_, _optimistic rollup_ protocol designed to _scale Ethereum_.

## Network Participants

There are three actors in Optimistic Ethereum: users, sequencers, and verifiers.

![Network Overview](./assets/network-participants-overview.svg)

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

1. Verifying rollup integrity and disputing invalid assertions.
2. Serving rollup data to users; and
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

When you send a transaction, it is sent to the sequencer. The sequencer will
verify your transaction, execute it if valid, and confirm it. This happens much
faster than on Ethereum mainnet (around ~1s).

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

## L2 Blocks, Batches & Epochs

Optimistic Ethereum has blocks, however these work slightly differently from L1
blocks, which are mined by proof-of-work and will soon be decided by
proof-of-stake.

Optimistic Ethereum has two kind of blocks:
1. L2 deposit blocks
2. L2 sequencer blocks

As users submit transactions, the sequencer confirms them and accumulates them
in **L2 sequencer blocks**. The sequencer includes these blocks in *batches*
which it then submits to L1 (as calldata of a contract call to the *sequencer
block feed contract*).

**TODO:** contract name is speculative

A batch may contain multiple block or even partial blocks: the start of a block
can be at the end of a batch while the rest of the block can be at the start of
the next batch.

**_L2 deposit blocks_**, on the other hand, arise from L1 blocks. Each L1 block
implicitly defines one L2 deposit block, which comprises the following data:

1. L1 block properties, namely
   - block hash
   - block number
   - timestamp
   - base fee
2. L2 transactions submitted on L1

These L2 transactions that have been submitted L1 are called *deposits*.

These transactions have two main uses:
- Ensure that the rollup remains live even if the sequencer goes down or starts
  censoring L2 transactions. Consequently, funds can never remain stuck on the
  rollup.
- Deposit ETH and ERC-20 tokens onto the rollup ("bridging"). This is achieved
  by sending the token to a contract which locks it, then sending a L2
  transaction (on L1) instructing a L2 contract to mint an equivalent L2 token.
  The transaction records the address of the L1 contract that submitted it, and
  the L2 contract only allows minting if the transaction was submitted by an
  authorized L1 contract.

  Anybody can operate such a pair of contracts, but Optimism PBC operates the
  bridge for ETH and some blue chip tokens ([the gateway]). See also other
  [trusted bridge providers].

[the gateway]: https://gateway.optimism.io/
[trusted bridge providers]: https://www.optimism.io/apps/bridges

Finally, we have **epochs**. There is one epoch per L1 block. Each epoch starts
with the L2 deposit block generated by the L1 block, followed by zero or more L2
sequencer blocks. The epoch number is equal to the L1 block number.

**TODO:** are we agreed on epoch number == L1 block number?

For every L2 sequencer block, the sequencer chooses the epoch in which the block
will appear. Successive blocks must belong to epochs with monotonically
increasing numbers, and the chosen epochs must respecting sequencing window
constraints (see next section).

## L2 Chain Derivation

The canonical L2 blockchain is derived from the L1 blockchain. For this purpose,
the sequencer and verifiers read:

- the L2 sequencer blocks posted on L1 by the sequencer, to the *L2 sequencer
  block feed contract*
- L1 block properties (number, timestamp, basefee, ...)
- L2 transactions submitted on L1 (**TODO:** where?)

We derive an L2 deposit block from each L1 block. The L2 deposit block contains
the L1 block properties and the L2 transactions submitted on L1. This deposit
block is the first block of the epoch with the same number as the L1 block.

The L2 deposit block is followed by all L2 sequencer blocks for the same epoch
(as determined by the sequencer), however these L2 sequencer blocks must land on
L1 within the epoch's *sequencing window*.

Each epoch derived from L1 block number `B` has a sequencing window of constant
size `S`, such that its sequencing window is the L1 block range `[B+1,B+S]`.

It is illegal for the sequencer to submit a sequencer block for epoch `B` after
L1 block `B+S`. If he does anyway, that block (or the block chunk, if the block
is split between multiple batches) is simply ignored during the L2 blockchain
derivation process.

An important consequence of the sequencing window is that the blocks in epoch
`B` can be fully determined by looking at at the L1 block range `[B,B+S]` (the
block `B` + its sequencing window).

---
---

**TODO:** explain more stuff :)
