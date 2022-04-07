/* Imports: External */
import { task, types } from 'hardhat/config'
import dotenv from 'dotenv'

import { expectApprox, sleep } from '@eth-optimism/core-utils'
import { Wallet, BigNumber, Contract, ContractFactory, constants, utils } from 'ethers'
import { serialize } from '@ethersproject/transactions'
import { ethers } from 'hardhat'
import {
  TransactionReceipt,
  TransactionRequest,
} from '@ethersproject/providers'

/* Imports: Internal */
import { expect } from './shared/setup'
import { defaultTransactionFactory } from './shared/utils'
import { OptimismEnv } from './shared/env'

describe('Basic Deposit tests', () => {
  let env: OptimismEnv
  let wallet: Wallet

  before(async () => {
    env = await OptimismEnv.new()
  })

  describe('DepositFeed.depositTransaction', () => {
    it('should correctly process a valid deposit', async () => {
      const depositFeedArtifact = require('../../opnode/contracts/abis/DepositFeed.json')

      dotenv.config()

      const depositFeed = new Contract(
	"deaddeaddeaddeaddeaddeaddeaddeaddead0001",
	depositFeedArtifact.abi
      ).connect(env.l1Wallet)

      const tx = defaultTransactionFactory()
      tx.value = utils.parseEther("1.337")
      const gas = utils.parseEther('0.01')
      const result = await depositFeed.depositTransaction(
	tx.to,
	tx.value,
	'3000000',
	false,
	[],
	{
	  value: tx.value.add(gas),
	}
      )
      await result.wait()

      expect(result.value).to.equal(tx.value.add(gas))
    })
  })
})
