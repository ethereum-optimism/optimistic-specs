# state-surgery

This package performs state surgery. It takes a v0 database and a partial `genesis.json` as input, and creates an initialized Bedrock database as output. It performs the following steps to do this:

1. Iterates over the old state.
2. For each account in the old state, add that account and its storage to the new state after copying its balance from the OVM_ETH contract.
3. Iterates over the pre-allocated accounts in the genesis file and adds them to the new state.
4. Configures a genesis block in the new state using `genesis.json`.

This process takes about two hours on mainnet.

Unlike previous iterations of our state surgery scripts, this one does not write results to a `genesis.json` file. This is for the following reasons:

1. **Performance**. It's much faster to write binary to LevelDB than it is to write strings to a JSON file.
2. **State Size**. There are nearly 1MM accounts on mainnet, which would create a genesis file several gigabytes in size. This is impossible for Geth to import without a large amount of memory, since the entire JSON gets buffered into memory. Importing the entire state database will be much faster, and can be done with fewer resources.

## Compilation

Run `make surgery`.

## Usage

```
NAME:
   surgery - migrates data from v0 to Bedrock

USAGE:
   surgery [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --data-dir value, -d value               data directory to read
   --state-root value, -r value             state root to dump
   --genesis-file value, -g value           path to a genesis file
   --out-dir value, -o value                path to output directory
   --expected-total-supply value, -e value  expected total ETH supply
   --help, -h                               show help (default: false)
```