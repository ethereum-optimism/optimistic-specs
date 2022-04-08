import * as RLP from '@ethersproject/rlp'
import { BigNumber, BigNumberish } from '@ethersproject/bignumber'
import { getAddress } from '@ethersproject/address'
import {
  hexConcat,
  stripZeros,
  zeroPad,
  arrayify,
  BytesLike,
} from '@ethersproject/bytes'
import { keccak256 } from '@ethersproject/keccak256'
import { Zero } from '@ethersproject/constants'
import { ContractReceipt, Event } from '@ethersproject/contracts'

function formatNumber(value: BigNumberish, name: string): Uint8Array {
  const result = stripZeros(BigNumber.from(value).toHexString())
  if (result.length > 32) {
    throw new Error(`invalid length for ${name}`)
  }
  return result
}

function handleNumber(value: string): BigNumber {
  if (value === '0x') {
    return Zero
  }
  return BigNumber.from(value)
}

function handleAddress(value: string): string {
  if (value === '0x') {
    // @ts-ignore
    return null
  }
  return getAddress(value)
}

export enum SourceHashDomain {
  UserDeposit = 0,
  L1InfoDeposit = 1,
}

interface DepositTxOpts {
  sourceHash?: string
  from: string
  to: string | null
  mint: BigNumberish
  value: BigNumberish
  gas: BigNumberish
  data: string
  domain?: SourceHashDomain
  l1BlockHash?: string
  logIndex?: BigNumberish
  sequenceNumber?: BigNumberish
}

interface DepostTxExtraOpts {
  domain?: SourceHashDomain
  l1BlockHash?: string
  logIndex?: BigNumberish
  sequenceNumber?: BigNumberish
}

export class DepositTx {
  public type = '0x7E'
  private _sourceHash?: string
  public from: string
  public to: string | null
  public mint: BigNumberish
  public value: BigNumberish
  public gas: BigNumberish
  public data: BigNumberish

  public domain?: SourceHashDomain
  public l1BlockHash?: string
  public logIndex?: BigNumberish
  public sequenceNumber?: BigNumberish

  constructor(opts: Partial<DepositTxOpts> = {}) {
    this._sourceHash = opts.sourceHash
    this.from = opts.from!
    this.to = opts.to!
    this.mint = opts.mint!
    this.value = opts.value!
    this.gas = opts.gas!
    this.data = opts.data!
    this.domain = opts.domain
    this.l1BlockHash = opts.l1BlockHash
    this.logIndex = opts.logIndex
    this.sequenceNumber = opts.sequenceNumber
  }

  hash() {
    const encoded = this.encode()
    return keccak256(encoded)
  }

  sourceHash() {
    if (!this._sourceHash) {
      let marker: string
      switch (this.domain) {
        case SourceHashDomain.UserDeposit:
          marker = BigNumber.from(this.logIndex).toHexString()
          break
        case SourceHashDomain.L1InfoDeposit:
          marker = BigNumber.from(this.sequenceNumber).toHexString()
          break
        default:
          throw new Error(`Unknown domain: ${this.domain}`)
      }

      if (!this.l1BlockHash) {
        throw new Error('Need l1BlockHash to compute sourceHash')
      }

      const l1BlockHash = this.l1BlockHash
      const input = hexConcat([l1BlockHash, zeroPad(marker, 32)])
      const depositIDHash = keccak256(input)
      const domain = BigNumber.from(this.domain).toHexString()
      const domainInput = hexConcat([zeroPad(domain, 32), depositIDHash])
      this._sourceHash = keccak256(domainInput)
    }
    return this._sourceHash
  }

  encode() {
    const fields: any = [
      this.sourceHash() || '0x',
      getAddress(this.from) || '0x',
      this.to != null ? getAddress(this.to) : '0x',
      formatNumber(this.mint || 0, 'mint'),
      formatNumber(this.value || 0, 'value'),
      formatNumber(this.gas || 0, 'gas'),
      this.data || '0x',
    ]

    return hexConcat([this.type, RLP.encode(fields)])
  }

  decode(raw: BytesLike, extra: DepostTxExtraOpts = {}) {
    const payload = arrayify(raw)
    const transaction = RLP.decode(payload.slice(1))

    this._sourceHash = transaction[0]
    this.from = handleAddress(transaction[1])
    this.to = handleAddress(transaction[2])
    this.mint = handleNumber(transaction[3])
    this.value = handleNumber(transaction[4])
    this.gas = handleNumber(transaction[5])
    this.data = transaction[6]

    if ('l1BlockHash' in extra) {
      this.l1BlockHash = extra.l1BlockHash
    }
    if ('domain' in extra) {
      this.domain = extra.domain
    }
    if ('logIndex' in extra) {
      this.logIndex = extra.logIndex
    }
    if ('sequenceNumber' in extra) {
      this.sequenceNumber = extra.sequenceNumber
    }
    return this
  }

  static decode(raw: BytesLike, extra?: DepostTxExtraOpts): DepositTx {
    return new this().decode(raw, extra)
  }

  fromL1Receipt(receipt: ContractReceipt, index: number): DepositTx {
    if (!receipt.events) throw new Error('cannot parse receipt')
    const event = receipt.events[index]
    if (!event) {
      throw new Error(`event index ${index} does not exist`)
    }
    return this.fromL1Event(event)
  }

  static fromL1Receipt(receipt: ContractReceipt, index: number): DepositTx {
    return new this({}).fromL1Receipt(receipt, index)
  }

  fromL1Event(event: Event): DepositTx {
    if (event.event !== 'TransactionDeposited')
      throw new Error(`incorrect event type: ${event.event}`)
    if (typeof event.args === 'undefined') throw new Error('no event args')
    if (typeof event.args.from === 'undefined')
      throw new Error('"from" undefined')
    this.from = event.args.from
    if (typeof event.args.isCreation === 'undefined')
      throw new Error('"isCreation" undefined')
    if (typeof event.args.to === 'undefined') throw new Error('"to" undefined')
    this.to = event.args.isCreation ? null : event.args.to
    if (typeof event.args.mint === 'undefined')
      throw new Error('"mint" undefined')
    this.mint = event.args.mint
    if (typeof event.args.value === 'undefined')
      throw new Error('"value" undefined')
    this.value = event.args.value
    if (typeof event.args.gasLimit === 'undefined')
      throw new Error('"gasLimit" undefined')
    this.gas = event.args.gasLimit
    if (typeof event.args.data === 'undefined')
      throw new Error('"data" undefined')
    this.data = event.args.data
    this.domain = SourceHashDomain.UserDeposit
    this.l1BlockHash = event.blockHash
    this.logIndex = event.logIndex
    return this
  }

  static fromL1Event(event: Event): DepositTx {
    return new this({}).fromL1Event(event)
  }
}
