import * as RLP from '@ethersproject/rlp'
import { BigNumber, BigNumberish } from '@ethersproject/bignumber'
import { getAddress } from '@ethersproject/address'
import { hexConcat, stripZeros } from '@ethersproject/bytes'
import { keccak256 } from '@ethersproject/keccak256'

function formatNumber(value: BigNumberish, name: string): Uint8Array {
  const result = stripZeros(BigNumber.from(value).toHexString())
  if (result.length > 32) {
    throw new Error(`invalid length for ${name}`)
  }
  return result
}

interface DepositTxOpts {
  blockHeight: BigNumberish
  transactionIndex: BigNumberish
  from: string
  to: string | null
  mint: BigNumberish
  value: BigNumberish
  gas: BigNumberish
  data: string
}

export class DepositTx {
  public type = '0x7E'
  public blockHeight: BigNumberish
  public transactionIndex: BigNumberish
  public from: string
  public to: string | null
  public mint: BigNumberish
  public value: BigNumberish
  public gas: BigNumberish
  public data: BigNumberish

  constructor(opts: DepositTxOpts) {
    this.blockHeight = opts.blockHeight
    this.transactionIndex = opts.transactionIndex
    this.from = opts.from
    this.to = opts.to
    this.mint = opts.mint
    this.value = opts.value
    this.gas = opts.gas
    this.data = opts.data
  }

  hash() {
    const encoded = this.encode()
    return keccak256(encoded)
  }

  encode() {
    const fields: any = [
      formatNumber(this.blockHeight || 0, 'blockHeight'),
      formatNumber(this.transactionIndex || 0, 'transactionIndex'),
      getAddress(this.from) || '0x',
      this.to != null ? getAddress(this.to) : '0x',
      formatNumber(this.mint || 0, 'mint'),
      formatNumber(this.value || 0, 'value'),
      formatNumber(this.gas || 0, 'gas'),
      this.data || '0x',
    ]

    return hexConcat([this.type, RLP.encode(fields)])
  }
}
