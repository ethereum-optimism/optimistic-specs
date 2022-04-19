// Named 000 in order to run first since the output submitter
// can fall behind.

/* Imports: External */
import { BigNumber, constants, Contract, ContractReceipt, utils, Wallet } from 'ethers'
import { awaitCondition } from '@eth-optimism/core-utils'
import * as rlp from 'rlp'
import { Block } from '@ethersproject/abstract-provider'

/* Imports: Internal */
import { WITHDRAWER_ADDR } from './shared/constants'
import env from './shared/env'
import { expect } from './shared/setup'
import winston from 'winston'

const withdrawerArtifact = require('../../contracts/artifacts/contracts/L2/Withdrawer.sol/Withdrawer.json')
const l2OOracleArtifact = require('../../contracts/artifacts/contracts/L1/L2OutputOracle.sol/L2OutputOracle.json')
const counterArtifact = require('../artifacts/Counter.sol/Counter.json')
const multiDepositorArtifact = require('../artifacts/MultiDepositor.sol/MultiDepositor.json')

describe('Withdrawals', () => {
  let logger: winston.Logger
  let portal: Contract
  let withdrawer: Contract

  let recipient: Wallet

  before(async () => {
    logger = env.logger
    portal = env.optimismPortal

    withdrawer = new Contract(
      WITHDRAWER_ADDR,
      withdrawerArtifact.abi,
    )
  })

  describe('simple withdrawals', () => {
    let nonce: BigNumber
    let burnBlock: Block
    let withdrawalHash: string
    const value = utils.parseEther('1')
    const gasLimit = 3000000

    before(async function () {
      this.timeout(60_000)
      recipient = Wallet.createRandom().connect(env.l2Provider)
      withdrawer = withdrawer.connect(recipient)

      logger.info('Generated new wallet', {
        recipient: recipient.address,
      })
      logger.info('Depositing to new address on L2')
      let tx = await portal.connect(env.l1Wallet).depositTransaction(
        recipient.address,
        utils.parseEther('1.337'),
        gasLimit,
        false,
        [],
        {
          value: utils.parseEther('1.337'),
        },
      )
      await tx.wait()

      await awaitCondition(async () => {
        const bal = await recipient.getBalance()
        return bal.eq(tx.value)
      })

      logger.info('Transferring funds on L1')
      tx = await env.l1Wallet.sendTransaction({
        to: recipient.address,
        value,
      })
      await tx.wait()
    })

    it('should create a withdrawal on L2', async () => {
      nonce = await withdrawer.nonce()
      const tx = await withdrawer.initiateWithdrawal(
        recipient.address,
        gasLimit,
        [],
        {
          value,
        },
      )
      const receipt: ContractReceipt = await tx.wait()
      expect(receipt.events!.length).to.eq(1)
      expect(receipt.events![0].args).to.deep.eq([
        nonce,
        recipient.address,
        recipient.address,
        value,
        BigNumber.from(gasLimit),
        '0x',
      ])

      burnBlock = await env.l2Provider.getBlock(receipt.blockHash)
      withdrawalHash = utils.keccak256(
        utils.defaultAbiCoder.encode(
          ['uint256', 'address', 'address', 'uint256', 'uint256', 'bytes'],
          [
            utils.hexZeroPad(nonce.toHexString(), 32),
            recipient.address,
            recipient.address,
            value,
            gasLimit,
            '0x',
          ],
        ),
      )
    })

    it('should verify the withdrawal on L1', async function () {
      recipient = recipient.connect(env.l1Provider)
      portal = portal.connect(recipient)
      const oracle = new Contract(
        await portal.L2_ORACLE(),
        l2OOracleArtifact.abi,
      ).connect(recipient)

      // Calculate the target output timestamp to make the withdrawal proof against. ie. the first
      // output with a timestamp greater than the burn block timestamp
      const submissionInterval = (await oracle.SUBMISSION_INTERVAL()).toNumber()
      const startingBlockTimestamp = (await oracle.STARTING_BLOCK_TIMESTAMP()).toNumber()
      let nextTimestamp = (await oracle.nextTimestamp()).toNumber()
      let targetOutputTimestamp
      if (burnBlock.timestamp < nextTimestamp) {
        // Just use the next timestamp
        targetOutputTimestamp = nextTimestamp
      } else {
        // Calculate the first timestamp greater than the burnBlock which will be appended.
        targetOutputTimestamp =
        Math.ceil(
          (burnBlock.timestamp - startingBlockTimestamp)
          / submissionInterval
        )
        * submissionInterval
        + startingBlockTimestamp
      }

      // Set the timeout based on the diff between latest output and target output timestamp
      let latestBlockTimestamp = (await oracle.latestBlockTimestamp()).toNumber()
      let difference = targetOutputTimestamp - latestBlockTimestamp
      this.timeout(difference * 5000)

      let output: string
      await awaitCondition(async () => {
        output = await oracle.getL2Output(targetOutputTimestamp)
        latestBlockTimestamp = (await oracle.latestBlockTimestamp()).toNumber()
        if(targetOutputTimestamp - latestBlockTimestamp < difference){
          // Only log when a new output has been appended
          difference = targetOutputTimestamp - latestBlockTimestamp
          logger.info('Waiting for output submission', {
            targetTimestamp: targetOutputTimestamp,
            latestOracleTS: latestBlockTimestamp,
            difference,
            output,
          })
        }
        return output != constants.HashZero
      }, 2000, 240)

      // suppress compilation errors since Typescript cannot detect
      // that awaitCondition above will throw if it times out.
      targetOutputTimestamp = targetOutputTimestamp!
      output = output!

      const blocksSinceBurn = Math.floor((targetOutputTimestamp - burnBlock.timestamp) / 2)
      const targetBlockNum = burnBlock.number + blocksSinceBurn + 1
      const targetBlockNumHex = utils.hexValue(targetBlockNum)
      const storageSlot = '00'.repeat(31) + '01' // i.e the second variable declared in the contract
      const proof = await env.l2Provider.send('eth_getProof', [
        WITHDRAWER_ADDR,
        [utils.keccak256(withdrawalHash + storageSlot)],
        targetBlockNumHex,
      ])

      const {stateRoot: targetStateRoot, hash: targetHash} = await env.l2Provider.send(
        'eth_getBlockByNumber',
        [
          targetBlockNumHex,
          false,
        ],
      )

      const finalizationPeriod = (await portal.FINALIZATION_PERIOD()).toNumber()
      logger.info('Waiting finalization period', {
        seconds: finalizationPeriod,
      })
      await new Promise((resolve) => setTimeout(resolve, finalizationPeriod * 1000))

      logger.info('Finalizing withdrawal')
      const initialBal = await recipient.getBalance()
      const tx = await portal.finalizeWithdrawalTransaction(
        nonce,
        recipient.address,
        recipient.address,
        value,
        gasLimit,
        '0x',
        targetOutputTimestamp,
        {
          version: constants.HashZero,
          stateRoot: targetStateRoot,
          withdrawerStorageRoot: proof.storageHash,
          latestBlockhash: targetHash,
        },
        rlp.encode(proof.storageProof[0].proof),
        {
          gasLimit,
        },
      )
      await tx.wait()
      const finalBal = await recipient.getBalance()
      expect(finalBal.gte(initialBal)).to.be.true
    }).timeout(180_000)
  })
})
