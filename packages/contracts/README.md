# Optimism: Bedrock Edition - Contracts

## Usage

Install node modules with yarn (v1), and Node.js (14+).

```shell
yarn
```

## Running Tests

The repo currently uses a mix of typescript tests (run with HardHat) and solidity tests (run with forge). The project
uses the default hardhat directory structure.

See installation instructions for forge [here](https://github.com/gakonst/foundry).

The full test suite can be executed via `yarn`:

```shell
yarn test
```

To run only typescript tests:

```shell
yarn test:hh
```

To run only solidity tests:

```shell
yarn test:forge
```
