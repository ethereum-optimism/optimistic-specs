import { expect } from 'chai'
import { DepositTx } from '../helpers'

describe('Helpers', () => {
  describe('DepositTx', () => {
    it('should serialize and hash', () => {
      // constants serialized using optimistic-geth
      const hash =
        '0xb041f1d6fe1fd9752760ed753b1386c751253d26909bb58a713d5fce31ffd504'
      const raw =
        '0x7ef8458203220194de3829a23df1479438622a08a116e8eb3f620bb594b7e390864a90b7b923c9f9310c6f98aafe43f707880de0b6b3a7640000880e043da617250000832dc6c080'

      const tx = new DepositTx({
        blockHeight: '0x322',
        transactionIndex: '0x1',
        from: '0xde3829a23df1479438622a08a116e8eb3f620bb5',
        to: '0xb7e390864a90b7b923c9f9310c6f98aafe43f707',
        mint: '0xde0b6b3a7640000',
        value: '0xe043da617250000',
        gas: '0x2dc6c0',
        data: '0x',
      })

      const encoded = tx.encode()
      expect(encoded).to.deep.eq(raw)
      const hashed = tx.hash()
      expect(hashed).to.deep.eq(hash)
    })
  })
})
