import { task, types } from 'hardhat/config'
import { Contract, providers, utils, Wallet } from 'ethers'
import dotenv from 'dotenv'
import { DepositTx } from '../helpers/index'

dotenv.config()

async function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

task('deposit', 'Deposits funds onto L2.')
  .addParam(
    'l1ProviderUrl',
    'L1 provider URL.',
    'http://localhost:8545',
    types.string
  )
  .addParam(
    'l2ProviderUrl',
    'L2 provider URL.',
    'http://localhost:9545',
    types.string
  )
  .addParam('to', 'Recipient address.', null, types.string)
  .addParam('amountEth', 'Amount in ETH to send.', null, types.string)
  .addOptionalParam(
    'privateKey',
    'Private key to send transaction',
    process.env.PRIVATE_KEY,
    types.string
  )
  .addOptionalParam(
    'depositContractAddr',
    'Address of deposit contract.',
    'deaddeaddeaddeaddeaddeaddeaddeaddead0001',
    types.string
  )
  .setAction(async (args) => {
    const {
      l1ProviderUrl,
      l2ProviderUrl,
      to,
      amountEth,
      depositContractAddr,
      privateKey,
    } = args
    const depositFeedArtifact = require('../artifacts/contracts/L1/DepositFeed.sol/DepositFeed.json')

    const l1Provider = new providers.JsonRpcProvider(l1ProviderUrl)
    const l2Provider = new providers.JsonRpcProvider(l2ProviderUrl)

    let l1Wallet: Wallet | providers.JsonRpcSigner
    if (privateKey) {
      l1Wallet = new Wallet(privateKey, l1Provider)
    } else {
      l1Wallet = l1Provider.getSigner()
    }

    const from = await l1Wallet.getAddress()
    console.log(`Sending from ${from}`)
    const balance = await l1Wallet.getBalance()
    if (balance.eq(0)) {
      throw new Error(`${from} has no balance`)
    }

    const depositFeed = new Contract(
      depositContractAddr,
      depositFeedArtifact.abi
    ).connect(l1Wallet)

    const amountWei = utils.parseEther(amountEth)
    const value = amountWei.add(utils.parseEther('0.01'))
    console.log(
      `Depositing ${amountEth} ETH to ${to} with ${value.toString()} value...`
    )
    // Below adds 0.01 ETH to account for gas.
    const tx = await depositFeed.depositTransaction(
      to,
      amountWei,
      '3000000',
      false,
      [],
      { value }
    )
    console.log(`Got TX hash ${tx.hash}. Waiting...`)
    const receipt = await tx.wait()

    const event = receipt.events[0]
    if (event?.event !== 'TransactionDeposited') {
      throw new Error('Transaction not deposited')
    }

    let found = false
    l2Provider.on('block', async (number) => {
      const block = await l2Provider.getBlock(number)
      for (const [i, hash] of block.transactions.entries()) {
        const l2tx = new DepositTx({
          blockHeight: number,
          transactionIndex: i,
          from: event.args.from,
          to: event.args.isCreation ? null : event.args.to,
          mint: event.args.mint,
          value: event.args.value,
          gas: event.args.gasLimit,
          data: event.args.data,
        })

        if (l2tx.hash() === hash) {
          console.log(`Got L2 TX hash ${hash}`)
          found = true
        }
      }
    })

    // eslint-disable-next-line
    while (!found) {
      await sleep(500)
    }
  })
