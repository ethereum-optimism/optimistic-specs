# Bisection game design

# Context

We assume that the sequencer committed to some transition in the roll-up's state $T = (R, transactions) \to R'$, where R' is the state root (aka `chainState.root`) after applying `transactions` to R, in order.

Suppose a challenger wants to challenge this state. Consider a `MachineState` type something like the following:

```tsx
type MachineState {
  root: Bytes32; // hash of (chainState.root, memory.root, instructions)
  chainState: ChainState;
  memory: Memory;
  instructions: Instruction;
  // etc
}
```

(MachineState is assumed to be a Merkle tree, thus the `root` property.)

We force the sequencer to post `S_0.root`, where `S_0 = {chainState: R, memory: emptyMemory, instructions: dissassemble(transactions)}` . The challenger contract must validate that `R` is equal to `R` posted for normal operation rollup. In addition, the challenger contract must validate that `dissassemble(transaction)` is correct. These validations are currently out of scope. We also force the sequencer to post `S'.root`, where `S' = {chainState: R', memory: emptyMemory, instructions: emptyInstructions}` . We might force the sequencer to post additional intermediate states.

We assume there is a function `progress: (state: MachineState) => MachineState` which takes in a machine state `s` and returns the next state. It works by popping the latest instruction, and modifying itself according to the instruction to produce the next state. When `instructions` is empty, `progress(state)` should indicate that it's reached a final state.

The following description is restricted in scope to _correctness_ requirements, and ignores _liveness_ requirements, **assuming that the sequencer & challenger each act in a timely manner.** The specific incentives & timeouts are not considered in this document.

# High level overview

The `progress` function _defines_ what it means for a transition $(R, transactions) \to R'$ to be valid, in the following sense:

1. From $(R, transactions)$, one can construct the "initial" machine state $M_0$.
2. Define $M_{i+1} = progress(M_i)$
3. Compute $M_1, M_2, ...$ until you arrive at a "terminal" state $M_t$
4. Check that $M_t.chainState.root == R'$. If yes, $(R, transactions) \to R'$ is valid; else, it is invalid.

In the event of a validator disputing a sequencer-posted state root, the "bisection game" plays out as follows.

1. [Off-chain] The sequencer computes _its_ belief about what the correct sequence of machine states is, generating a sequence $S_0, ..., S_n$. (A correctly behaving sequencer would find $M_i = S_i$.)
2. [Off-chain] The validator computes _their_ belief about what the correct sequence of machine states is, generating a sequence $V_0, ..., V_m$. (A correctly behaving validator would find $M_i = V_i$.)
3. [On-chain] The sequencer commits to the root of the machine state at some steps $step_0, step_1, ..., step_k$, where initially, $step_k = \lfloor k/n \rfloor$.
4. [On-chain] The validator compares $S_{step_i}.root$ to $V_{step_i}.root$ for each $i$, finding the first $i$ such that they differ. (This implies $S_{step_{i-1}} \to S_{step_i}$ is an invalid transition). They assert on-chain that $step_i$ is invalid, forcing the sequencer to split the transition from $step_{i-1}$ to $step_i$ into k smaller transitions.
5. Repeat 4, narrowing down the values of $step_i$ until the validator can directly prove an invalid transition on-chain.

**Comment:** This seems different from the game played in the Arbitrum contract. In an Arbitrum challenge, the proposer seems to initially commit to the root of a binary Merkle tree formed from the [entire execution trace](https://github.com/OffchainLabs/arbitrum/blob/41858dc49be854188ec821ed956e767db1e617a9/packages/arb-bridge-eth/contracts/challenge/Challenge.sol#L126). A path down that tree is [revealed by the proposer](https://github.com/OffchainLabs/arbitrum/blob/41858dc49be854188ec821ed956e767db1e617a9/packages/arb-bridge-eth/contracts/challenge/Challenge.sol#L189-L194). Thus, `Challenge.sol` seems to require the sequencer and challenger to hash every machine state up front, while this design allows the sequencer to hash machine states that are committed to on-chain.

Working prototype code can be found in [this repo](src/bisection.ts)), written in Typescript.

## Initializing a challenge

Assume that a `Challenge.sol` contract is used to adjudicate a single challenge. There may be a `ChallengeFactory.sol` contract which deploys new `Challenge` contracts — this is how Arbitrum does it, for instance.

The challenger deploys a challenge contract instance to manage the challenge for the transition T. A commitment to R, R' and `transactions` should be recorded.

This logic is not currently specified.

## First commitment

The sequencer constructs some arbitrary sequence $S_0, S_1, ..., S_n$ and commits to the $k+1$ steps $[(S_0.root, 0), (S_{\lfloor n/k\rfloor}.root, \lfloor n/k \rfloor),..., (S_n.root, n)]$ on-chain:

```tsx
type StepCommitment = {root: Bytes32; step: number}

class ChallengeManager() {
  constructor(commitment: StepCommitment[]) {
    this.commitment = commitment;
  }
}

```

`k` is currently an unspecified parameter which dictates the branching factor. It should be at least 2.

- **Open question**: How large should `k` be? Why shouldn't it be as large as possible?

  > Vitalik suggests something like 128 parts or whatever the 4th root of the average number of steps is, to keep the number of rounds low.

  Why? Why not make `k` as large as possible?

  The goal might be to optimize for gas; in some cases, an extra round of bisection can drastically reduce the number of hashes to be revealed on-chain.

### Honest sequencer

In this step, the sequencer is committing to some initial machine state $S_0$, and a sequence of transitions $S_i → S_{i+1}$. The sequencer is claiming that this sequence of steps proves that $(R, transactions) \to R'$ is correct. The challenger is then tasked with proving one of 3 things:

1. $S_0$ is not correctly constructed from (R, transactions)
2. $S_n$ does not correctly correspond to R'
3. For some i, $S_{i+1} ≠ progress(S_i)$

An honest sequencer would therefore construct $S_i$ as follows. The sequencer constructs the initial machine state $S_0$ by:

- setting the state root to be R
- constructing the instructions from `transactions`
- correctly initializing other things (memory, flags, etc)

The sequencer then constructs $S_0, ..., S_n$ by the rule $S_{k+1} = progress(S_k)$, stopping when they arrive at a terminal state.

## Validating the starting state

The first step is a special case:

- An honest sequencer should (correctly) construct $S_0$ from $(R, transactions)$.
- A dishonest sequencer might not. In this case, the challenger needs to be able to prove that $S_0$ is incorrect.

```tsx
class ChallengeManager() {
  invalidateFirstStep(witness) {
    if (this.commitments[0].step !== 0) {
      throw 'can only invalidate step 0'
    }

    // TODO
  }
}
```

This code is left unspecified, as it depends on the `MachineState` type, which is not yet specified.

## Validating the end state

The last step is also a special case:

- An honest sequencer should commit to a terminal machine state $S_n$
- A dishonest sequencer might not. In this case, the challenger needs to prove things like:
  - $S_n$ is not terminal
  - $S_n$ does not include R' as the blockchain state root.

```tsx
class ChallengeManager() {
  invalidateLastStep(witness) {
    // TODO
  }
}
```

This code is left unspecified, as it depends on the `MachineState` type, which is not yet specified.

## Finding the first incorrect step

The challenger reacts to a commitment by specifying an index `i`

```tsx
class ChallengeManager() {
  assertIncorrectStep(idx) {
    this.splitIdx = idx
```

This forces the sequencer to later split `cm.commitment[splitIdx-1]` → `cm.commitment[splitIdx]` into k smaller transitions, each with roughly the same number of steps.

### Honest challenger

An honest challenger constructs the sequence $V_0, ..., V_m$ the same way that the honest sequencer constructs $S_0, S_1, ..., S_n$. (If one actor is dishonest, they may construct sequences of different length). For later reference, suppose $V_0, ..., V_m$ is stored in an array `validatedSteps`.

The challenger then proceeds to find the first i where $S_i  \neq V_i$.

- If $S_0 \neq V_0$, the challenger should use `cm.invalidateFirstStep` to prove fraud.
- If $S_n$ doesn't include $R'$ as its state root, the challenger should use `cm.invalidateLastStep` to prove fraud.

Otherwise, they find the index of the last step committed to on-chain:

```tsx
// off-chain
function firstInvalidCommitment(cm: ChallengeManager) {
  const i = 0;
  while true {
    const committedStep = cm.commitment[i]
    const correctRoot = validatedSteps[commitededStep.step].root
    if (committedStep.root == correctRoot) {
      i += 1
    } else {
      return i
    }
  }
}

const splitIdx = lastCorrectCommitment(cm);
```

## Zooming in on the first discrepancy

Let `s = commitment[splitIdx]` and `s' = commitment[splitIdx+1]`.

Force the challenger to split the transition $s \to s'$ into k transitions of (approximately) equal size:

```tsx
class ChallengeManager() {
  split(commitment: CommitmentStep[]) {
    const before = this.commitment[this.splitIdx-1]
    const after = this.commitment[this.splitIdx]

    require that before == steps[0]
    require that after == steps[k-1]

    // the sequencer has to commit to jumps of roughly even length
    const total = after.step - before.step
    for i = 0,...,k:
      require that commitment[i].step == floor(before.step + i*total/k)
  }
}
```

## Detecting fraud

The challenger repeatedly splits until it is feasible to prove fraud on-chain:

```tsx
class ChallengeManager() {
  detectFraud({witness, startingAt}: Proof): boolean {
    const before = this.commitments[startingAt];
    const after = this.commitments[startingAt + 1];

    if (before.root !== witness.root) {
      return false
    }

    // Simulate running out of gas when the verifier tries to detect fraud
    // The assumption here is that a single step can _always_ be
    // validated on-chain.
    if (after.step - before.step > 1) {
      const gasUsed = BigNumber.from(before.root).add(after.root);
      if (gasUsed.gt(9000)) {
        throw 'out of gas';
      }
    }

    for (let i = 0; i < after.step - before.step; i++) {
      witness = this.progress(witness);
    }

		return after.root == witness.root
  }
}
```

If it returns true, the sequencer should be slashed.

(Optionally, the sequencer could short-circuit a griefing attack, by proving _not fraud_ on their turn. This is probably of limited utility; if the sequencer can prove innocence of `n` steps of execution, they only need to bisect $log_k(n)$ more times to get to a single step fraud proof. In practice, $\log_k(n)$ is less than 1; else, k calls to `progress(witness)` will probably use too much gas.)

# Improvements

## Challenger can "flip roles"

**\*Benefit:** The challenger cannot be forced to submit $256\log_k(2)$ transactions to successfully prove fraud.\*

Suppose the sequencer wanted to grief the challenger, and force many rounds of bisection. Specifically, let's say that $M_0, ..., M_{2m}$ is the correct sequence of machine states, and the sequencer commits to $S_0, S_n, S_{2n}$, where $n \gg m$.

The challenger should be given _one opportunity_ to "flip roles", at the start of the game, and commit to $M_0, M_n, M_{2m}$. This should only be allowed if $m < n$ (which the chain can plainly see). This prevents the sequencer from claiming that the correct sequence of machine state states is huge, leading to a very long game.

## Challenger also splits

**\*Benefit:** You can halve the number of transactions by allowing the challenger to split.\*

Suppose the sequencer has committed to

```tsx
[
  {root: r_1, step: i},
  {root: r_2, step: i + stepSize},
  {root: r_3, step: i + 2 * stepSize}
];
```

Suppose the challenger wants to assert that $r_1 \to r_2$ is invalid. This implies

The challenger could do so by _also_ committing to

```tsx
[
  {root: r_1, step: i},
  {root: r_mid, step: i + stepSize / 2},
  {root: r_2_corrected, step: i + stepSize}
];
```

This works because both the sequencer and challenger agreed that $r_1$ is the correct root for step $i$, but they disagree that $r_2$ is the correct root for step $i + stepSize$.

**Question:** Does this work fine?
