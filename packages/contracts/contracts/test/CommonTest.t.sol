//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

/* Testing utilities */
import { Test } from "forge-std/Test.sol";
import { L2OutputOracle } from "../L1/L2OutputOracle.sol";
import { L2ToL1MessagePasser } from "../L2/L2ToL1MessagePasser.sol";
import { L1StandardBridge } from "../L1/L1StandardBridge.sol";
import { L2StandardBridge } from "../L2/L2StandardBridge.sol";
import { OptimismMintableTokenFactory } from "../universal/OptimismMintableTokenFactory.sol";
import { OptimismMintableERC20 } from "../universal/OptimismMintableERC20.sol";
import { OptimismPortal } from "../L1/OptimismPortal.sol";
import { L2ToL1MessagePasser } from "../L2/L2ToL1MessagePasser.sol";
import { L1CrossDomainMessenger } from "../L1/L1CrossDomainMessenger.sol";
import { L2CrossDomainMessenger } from "../L2/L2CrossDomainMessenger.sol";
import { AddressAliasHelper } from "@eth-optimism/contracts/standards/AddressAliasHelper.sol";

import {
    Lib_PredeployAddresses
} from "@eth-optimism/contracts/libraries/constants/Lib_PredeployAddresses.sol";

import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { console } from "forge-std/console.sol";

contract CommonTest is Test {
    address alice = address(128);
    address bob = address(256);

    address immutable ZERO_ADDRESS = address(0);
    address immutable NON_ZERO_ADDRESS = address(1);
    uint256 immutable NON_ZERO_VALUE = 100;
    uint256 immutable ZERO_VALUE = 0;
    uint64 immutable NON_ZERO_GASLIMIT = 50000;
    bytes32 nonZeroHash = keccak256(abi.encode("NON_ZERO"));
    bytes NON_ZERO_DATA = hex"0000111122223333444455556666777788889999aaaabbbbccccddddeeeeffff0000";

    function _setUp() public {
        // Give alice and bob some ETH
        vm.deal(alice, 1 << 16);
        vm.deal(bob, 1 << 16);

        vm.label(alice, "alice");
        vm.label(bob, "bob");
    }
}
contract L2OutputOracle_Initializer is CommonTest {
    // Test target
    L2OutputOracle oracle;

    // Constructor arguments
    address sequencer = 0x000000000000000000000000000000000000AbBa;
    uint256 submissionInterval = 1800;
    uint256 l2BlockTime = 2;
    bytes32 genesisL2Output = keccak256(abi.encode(0));
    uint256 historicalTotalBlocks = 100;

    // Cache of the initial L2 timestamp
    uint256 startingBlockTimestamp;

    // By default the first block has timestamp zero, which will cause underflows in the tests
    uint256 initTime = 1000;

    function setUp() public virtual {
        _setUp();

        // Move time forward so we have a non-zero starting timestamp
        vm.warp(initTime);
        // Deploy the L2OutputOracle and transfer owernship to the sequencer
        oracle = new L2OutputOracle(
            submissionInterval,
            l2BlockTime,
            genesisL2Output,
            historicalTotalBlocks,
            initTime,
            sequencer
        );
        startingBlockTimestamp = block.timestamp;
    }
}

contract Messenger_Initializer is L2OutputOracle_Initializer {
    OptimismPortal op;
    L1CrossDomainMessenger L1Messenger;
    L2CrossDomainMessenger L2Messenger;
    L2ToL1MessagePasser messagePasser;

    event SentMessage(
        address indexed target,
        address sender,
        bytes message,
        uint256 messageNonce,
        uint256 gasLimit
    );

    event WithdrawalInitiated(
        uint256 indexed nonce,
        address indexed sender,
        address indexed target,
        uint256 value,
        uint256 gasLimit,
        bytes data
    );

    event RelayedMessage(bytes32 indexed msgHash);

    event TransactionDeposited(
        address indexed from,
        address indexed to,
        uint256 mint,
        uint256 value,
        uint64 gasLimit,
        bool isCreation,
        bytes data
    );

    event WithdrawalFinalized(bytes32 indexed, bool success);

    function setUp() public virtual override {
        super.setUp();

        // Deploy the OptimismPortal
        op = new OptimismPortal(oracle, 100);
        vm.label(address(op), "OptimismPortal");

        L1Messenger = new L1CrossDomainMessenger();
        L1Messenger.initialize(op);

        L2CrossDomainMessenger l2m = new L2CrossDomainMessenger();
        vm.etch(Lib_PredeployAddresses.L2_CROSS_DOMAIN_MESSENGER, address(l2m).code);
        L2Messenger = L2CrossDomainMessenger(Lib_PredeployAddresses.L2_CROSS_DOMAIN_MESSENGER);

        L2Messenger.initialize(address(L1Messenger));

        // Set the L2ToL1MessagePasser at the correct address
        L2ToL1MessagePasser mp = new L2ToL1MessagePasser();
        vm.etch(Lib_PredeployAddresses.L2_TO_L1_MESSAGE_PASSER, address(mp).code);
        messagePasser = L2ToL1MessagePasser(payable(Lib_PredeployAddresses.L2_TO_L1_MESSAGE_PASSER));

        vm.label(
            Lib_PredeployAddresses.L2_TO_L1_MESSAGE_PASSER,
            "L2ToL1MessagePasser"
        );

        vm.label(
            Lib_PredeployAddresses.L2_CROSS_DOMAIN_MESSENGER,
            "L2CrossDomainMessenger"
        );

        vm.label(
            AddressAliasHelper.applyL1ToL2Alias(address(L1Messenger)),
            "L1CrossDomainMessenger_aliased"
        );
    }
}

contract Bridge_Initializer is Messenger_Initializer {
    L1StandardBridge L1Bridge;
    L2StandardBridge L2Bridge;
    OptimismMintableTokenFactory L2TokenFactory;
    OptimismMintableTokenFactory L1TokenFactory;
    ERC20 L1Token;
    OptimismMintableERC20 L2Token;
    ERC20 NativeL2Token;
    OptimismMintableERC20 RemoteL1Token;

    event ETHDepositInitiated(
        address indexed _from,
        address indexed _to,
        uint256 _amount,
        bytes _data
    );

    event ETHWithdrawalFinalized(
        address indexed _from,
        address indexed _to,
        uint256 _amount,
        bytes _data
    );

    event ERC20DepositInitiated(
        address indexed _l1Token,
        address indexed _l2Token,
        address indexed _from,
        address _to,
        uint256 _amount,
        bytes _data
    );

    event ERC20WithdrawalFinalized(
        address indexed _l1Token,
        address indexed _l2Token,
        address indexed _from,
        address _to,
        uint256 _amount,
        bytes _data
    );

    event WithdrawalInitiated(
        address indexed _l1Token,
        address indexed _l2Token,
        address indexed _from,
        address _to,
        uint256 _amount,
        bytes _data
    );

    event DepositFinalized(
        address indexed _l1Token,
        address indexed _l2Token,
        address indexed _from,
        address _to,
        uint256 _amount,
        bytes _data
    );

    event DepositFailed(
        address indexed _l1Token,
        address indexed _l2Token,
        address indexed _from,
        address _to,
        uint256 _amount,
        bytes _data
    );

    event ETHBridgeInitiated(
        address indexed _from,
        address indexed _to,
        uint256 _amount,
        bytes _data
    );

    event ETHBridgeFinalized(
        address indexed _from,
        address indexed _to,
        uint256 _amount,
        bytes _data
    );

    event ERC20BridgeInitiated(
        address indexed _localToken,
        address indexed _remoteToken,
        address indexed _from,
        address _to,
        uint256 _amount,
        bytes _data
    );

    event ERC20BridgeFinalized(
        address indexed _localToken,
        address indexed _remoteToken,
        address indexed _from,
        address _to,
        uint256 _amount,
        bytes _data
    );

    event ERC20BridgeFailed(
        address indexed _localToken,
        address indexed _remoteToken,
        address indexed _from,
        address _to,
        uint256 _amount,
        bytes _data
    );

    function setUp() public virtual override {
        super.setUp();

        vm.label(
            Lib_PredeployAddresses.L2_STANDARD_BRIDGE,
            "L2StandardBridge"
        );
        vm.label(
            Lib_PredeployAddresses.L2_STANDARD_TOKEN_FACTORY,
            "L2StandardTokenFactory"
        );

        // Deploy the L1 bridge and initialize it with the address of the
        // L1CrossDomainMessenger
        L1Bridge = new L1StandardBridge();
        L1Bridge.initialize(payable(address(L1Messenger)));
        vm.label(address(L1Bridge), "L1StandardBridge");

        // Deploy the L2StandardBridge, move it to the correct predeploy
        // address and then initialize it
        L2StandardBridge l2B = new L2StandardBridge();
        vm.etch(Lib_PredeployAddresses.L2_STANDARD_BRIDGE, address(l2B).code);
        L2Bridge = L2StandardBridge(payable(Lib_PredeployAddresses.L2_STANDARD_BRIDGE));
        L2Bridge.initialize(payable(address(L1Bridge)));

        // Set up the L2 mintable token factory
        OptimismMintableTokenFactory factory = new OptimismMintableTokenFactory();
        vm.etch(Lib_PredeployAddresses.L2_STANDARD_TOKEN_FACTORY, address(factory).code);
        L2TokenFactory = OptimismMintableTokenFactory(Lib_PredeployAddresses.L2_STANDARD_TOKEN_FACTORY);
        L2TokenFactory.initialize(Lib_PredeployAddresses.L2_STANDARD_BRIDGE);

        L1Token = new ERC20("Native L1 Token", "L1T");

        // Deploy the L2 ERC20 now
        L2Token = OptimismMintableERC20(L2TokenFactory.createStandardL2Token(
            address(L1Token),
            string(abi.encodePacked("L2-", L1Token.name())),
            string(abi.encodePacked("L2-", L1Token.symbol()))
        ));

        NativeL2Token = new ERC20("Native L2 Token", "L2T");
        L1TokenFactory = new OptimismMintableTokenFactory();
        L1TokenFactory.initialize(address(L1Bridge));

        RemoteL1Token = OptimismMintableERC20(L1TokenFactory.createStandardL2Token(
            address(NativeL2Token),
            string(abi.encodePacked("L1-", NativeL2Token.name())),
            string(abi.encodePacked("L1-", NativeL2Token.symbol()))
        ));
    }
}

