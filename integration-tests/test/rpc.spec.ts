/* Imports: External */
import { expectApprox, sleep } from '@eth-optimism/core-utils'
import { Wallet, BigNumber, Contract, ContractFactory, constants } from 'ethers'
import { serialize } from '@ethersproject/transactions'
import { ethers } from 'hardhat'
import {
  TransactionReceipt,
  TransactionRequest,
} from '@ethersproject/providers'

/* Imports: Internal */
import { expect } from './shared/setup'
import {
  defaultTransactionFactory,
  L2_CHAINID,
  gasPriceForL2,
  envConfig,
} from './shared/utils'
import { OptimismEnv } from './shared/env'

describe('Basic RPC tests', () => {
  let env: OptimismEnv
  let wallet: Wallet

  before(async () => {
    env = await OptimismEnv.new()
    wallet = env.l2Wallet
  })

  describe('eth_sendRawTransaction', () => {
    it('should correctly process a valid transaction', async () => {
      const tx = defaultTransactionFactory()
      tx.gasPrice = await gasPriceForL2()
      const nonce = await wallet.getTransactionCount()
      const result = await wallet.sendTransaction(tx)

      expect(result.from).to.equal(wallet.address)
      expect(result.nonce).to.equal(nonce)
      expect(result.gasLimit.toNumber()).to.equal(tx.gasLimit)
      expect(result.data).to.equal(tx.data)
    })
  })
})
