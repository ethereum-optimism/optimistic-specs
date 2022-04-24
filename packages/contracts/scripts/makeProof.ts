// Script for generating an inclusion proof for use in testing
import { generateMockWithdrawalProof } from '../helpers'

const args = process.argv.slice(2)

const [nonce, sender, target, value, gasLimit, data] = args

const main = async () => {
  const proof = await generateMockWithdrawalProof({
    nonce: +nonce,
    sender,
    target,
    value: +value,
    gasLimit: +gasLimit,
    data,
  })
  console.log(proof)
}
main()
