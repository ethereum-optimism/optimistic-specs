/* Imports: External */
import { Wallet, providers } from 'ethers'

/* Imports: Internal */
import {
  l1Provider,
  l2Provider,
  l1Wallet,
  l2Wallet,
} from './utils'

/// Helper class for instantiating a test environment with a funded account
export class OptimismEnv {
  // The wallets
  l1Wallet: Wallet
  l2Wallet: Wallet

  // The providers
  l1Provider: providers.JsonRpcProvider
  l2Provider: providers.JsonRpcProvider

  constructor(args: any) {
    this.l1Wallet = args.l1Wallet
    this.l2Wallet = args.l2Wallet
    this.l1Provider = args.l1Provider
    this.l2Provider = args.l2Provider
  }

  static async new(): Promise<OptimismEnv> {
    return new OptimismEnv({
      l1Wallet,
      l2Wallet,
      l1Provider,
      l2Provider,
    })
  }
}
