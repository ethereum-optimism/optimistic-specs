/* Imports: External */
import { Wallet, providers, BigNumber, utils } from 'ethers'
import { remove0x } from '@eth-optimism/core-utils'
import {
  asL2Provider,
} from '@eth-optimism/sdk'
import { cleanEnv, str, num, bool, makeValidator } from 'envalid'
import dotenv from 'dotenv'
dotenv.config()

export const isLiveNetwork = () => {
  return process.env.IS_LIVE_NETWORK === 'true'
}

export const DEFAULT_TEST_GAS_L1 = 330_000
export const DEFAULT_TEST_GAS_L2 = 1_300_000
export const ON_CHAIN_GAS_PRICE = 'onchain'

const gasPriceValidator = makeValidator((gasPrice) => {
  if (gasPrice === 'onchain') {
    return gasPrice
  }

  return num()._parse(gasPrice).toString()
})

const procEnv = cleanEnv(process.env, {
  L1_GAS_PRICE: gasPriceValidator({
    default: '0',
  }),
  L1_URL: str({ default: 'http://localhost:8545' }),
  L1_POLLING_INTERVAL: num({ default: 10 }),

  L2_CHAINID: num({ default: 17 }),
  L2_GAS_PRICE: gasPriceValidator({
    default: 'onchain',
  }),
  L2_URL: str({ default: 'http://localhost:9545' }),
  L2_POLLING_INTERVAL: num({ default: 10 }),

  PRIVATE_KEY: str({
    default:
      '0xbf7604d9d3a1c7748642b1b7b05c2bd219c9faa91458b370f85e5a40f3b03af7',
  }),

  MOCHA_TIMEOUT: num({
    default: 120_000,
  }),
  MOCHA_BAIL: bool({
    default: false,
  }),
})

export const envConfig = procEnv

// The hardhat instance
export const l1Provider = new providers.JsonRpcProvider(procEnv.L1_URL)
l1Provider.pollingInterval = procEnv.L1_POLLING_INTERVAL

export const l2Provider = asL2Provider(
  new providers.JsonRpcProvider(procEnv.L2_URL)
)
l2Provider.pollingInterval = procEnv.L2_POLLING_INTERVAL

// The sequencer private key which is funded on L1
export const l1Wallet = new Wallet(procEnv.PRIVATE_KEY, l1Provider)

// A random private key which should always be funded with deposits from L1 -> L2
// if it's using non-0 gas price
export const l2Wallet = l1Wallet.connect(l2Provider)

// Predeploys
export const L2_CHAINID = procEnv.L2_CHAINID

const abiCoder = new utils.AbiCoder()
export const encodeSolidityRevertMessage = (_reason: string): string => {
  return '0x08c379a0' + remove0x(abiCoder.encode(['string'], [_reason]))
}

export const defaultTransactionFactory = () => {
  return {
    to: '0x' + '1234'.repeat(10),
    gasLimit: 8_000_000,
    gasPrice: BigNumber.from(0),
    data: '0x',
    value: 0,
  }
}

export const gasPriceForL2 = async () => {
  if (procEnv.L2_GAS_PRICE === ON_CHAIN_GAS_PRICE) {
    return l2Wallet.getGasPrice()
  }

  return utils.parseUnits(procEnv.L2_GAS_PRICE, 'wei')
}

export const gasPriceForL1 = async () => {
  if (procEnv.L1_GAS_PRICE === ON_CHAIN_GAS_PRICE) {
    return l1Wallet.getGasPrice()
  }

  return utils.parseUnits(procEnv.L1_GAS_PRICE, 'wei')
}

export const die = (...args) => {
  console.log(...args)
  process.exit(1)
}

export const logStderr = (msg: string) => {
  process.stderr.write(`${msg}\n`)
}
