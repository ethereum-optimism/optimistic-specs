import { HardhatUserConfig, task, subtask } from 'hardhat/config'
import '@nomiclabs/hardhat-waffle'
import '@typechain/hardhat'
import 'hardhat-gas-reporter'
import 'solidity-coverage'

task('accounts', 'Prints the list of accounts', async (taskArgs, hre) => {
  const accounts = await hre.ethers.getSigners()

  for (const account of accounts) {
    console.log(account.address)
  }
})

const { TASK_COMPILE_SOLIDITY_GET_SOLC_BUILD } = require("hardhat/builtin-tasks/task-names");

subtask(TASK_COMPILE_SOLIDITY_GET_SOLC_BUILD, async (args, hre, runSuper) => {
  return {
    compilerPath: "solc",
    isSolcJs: false,
    version: "0.8.11",
    // this is used as extra information in the build-info files, but other than
    // that is not important
    longVersion: "0.8.11-local"
  }
})

const config: HardhatUserConfig = {
  solidity: '0.8.11',
  networks: {},
  gasReporter: {
    enabled: process.env.REPORT_GAS !== undefined,
    currency: 'USD',
  },
}

export default config
