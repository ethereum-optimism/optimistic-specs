import _ from 'lodash';

type Bytes32 = number;

export type StepCommitment = {root: Bytes32; step: number};
type State = {root: Bytes32};
type Proof = {startingAt: number; witness: State};

export class ChallengeManager {
  public incorrectStepIndex = 0;
  constructor(
    public commitments: StepCommitment[],
    public progress: (state: State) => State,
    public fingerprint: (state: State) => Bytes32
  ) {
    if (this.commitments.length < 2) {
      throw 'invalid commitment length';
    }
  }

  assertInvalidStep(index: number) {
    // checks
    if (index == 0) {
      throw 'cannot assert first step invalid';
    }

    if (index > this.commitments.length - 1) {
      throw 'Invalid challenge';
    }

    // effects
    this.incorrectStepIndex = index;
  }

  split(commitments: StepCommitment[]): any {
    // checks
    const before = this.commitments[this.incorrectStepIndex - 1];
    const after = this.commitments[this.incorrectStepIndex];

    const numSplits = commitments.length - 1;
    if (numSplits < 2) {
      throw 'invalid commitment length';
    }

    const first = commitments[0];
    const last = commitments[numSplits];

    if (!_.isEqual(first, before)) {
      throw 'first commitment is invalid';
    }

    if (!_.isEqual(last, after)) {
      throw 'last commitment is invalid';
    }

    for (let i = 0; i < numSplits; i++) {
      const {step} = commitments[i];
      const expectedStep = first.step + Math.floor((i * (last.step - first.step)) / numSplits);
      if (step !== Math.floor(expectedStep)) {
        throw 'invalid indices';
      }
    }

    // effects
    this.commitments = commitments;
  }

  detectFraud({witness, startingAt}: Proof, gasLimit = 1): boolean {
    const before = this.commitments[startingAt];
    const after = this.commitments[startingAt + 1];

    if (before.root !== witness.root) {
      return false;
    }

    let gasUsed = 0;
    for (let i = 0; i < after.step - before.step; i++) {
      gasUsed += witness.root;
      witness = this.progress(witness);
    }

    // Simulate running out of gas
    // The assumption here is that a single step can _always_ be validated on-chain.
    if (after.step - before.step > 1 && gasUsed > gasLimit) {
      throw new Error('out of gas');
    }

    return after.root !== witness.root;
  }
}
