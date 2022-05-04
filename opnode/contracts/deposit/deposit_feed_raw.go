// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package deposit

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// WithdrawalVerifierOutputRootProof is an auto generated low-level Go binding around an user-defined struct.
type WithdrawalVerifierOutputRootProof struct {
	Version               [32]byte
	StateRoot             [32]byte
	WithdrawerStorageRoot [32]byte
	LatestBlockhash       [32]byte
}

// OptimismPortalMetaData contains all meta data concerning the OptimismPortal contract.
var OptimismPortalMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractL2OutputOracle\",\"name\":\"_l2Oracle\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_finalizationPeriod\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"InvalidOutputRootProof\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidWithdrawalInclusionProof\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NonZeroCreationTarget\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotYetFinal\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"WithdrawalAlreadyFinalized\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"mint\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"additionalGasPrice\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"additionalGasLimit\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"guaranteedGas\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"isCreation\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"TransactionDeposited\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"WithdrawalFinalized\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"FINALIZATION_PERIOD\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"L2_ORACLE\",\"outputs\":[{\"internalType\":\"contractL2OutputOracle\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_additionalGasPrice\",\"type\":\"uint256\"},{\"internalType\":\"uint64\",\"name\":\"_additionalGasLimit\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"_guaranteedGas\",\"type\":\"uint64\"},{\"internalType\":\"bool\",\"name\":\"_isCreation\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"depositTransaction\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_gasLimit\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_timestamp\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"version\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"withdrawerStorageRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"latestBlockhash\",\"type\":\"bytes32\"}],\"internalType\":\"structWithdrawalVerifier.OutputRootProof\",\"name\":\"_outputRootProof\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"_withdrawalProof\",\"type\":\"bytes\"}],\"name\":\"finalizeWithdrawalTransaction\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"finalizedWithdrawals\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"l2Sender\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x60c0604052600080546001600160a01b03191661dead17905534801561002457600080fd5b506040516200233e3803806200233e8339810160408190526100459161005b565b6001600160a01b0390911660a052608052610095565b6000806040838503121561006e57600080fd5b82516001600160a01b038116811461008557600080fd5b6020939093015192949293505050565b60805160a051612276620000c86000396000818160a801526103590152600081816101a601526102cd01526122766000f3fe6080604052600436106100685760003560e01c8063a14238e711610043578063a14238e714610134578063eecf1c3614610174578063ff61cc931461019457600080fd5b80621c2ff61461009657806354314103146100f45780639bf62d821461010757600080fd5b366100915761008f33346000806175306000604051806020016040528060008152506101d6565b005b600080fd5b3480156100a257600080fd5b506100ca7f000000000000000000000000000000000000000000000000000000000000000081565b60405173ffffffffffffffffffffffffffffffffffffffff90911681526020015b60405180910390f35b61008f610102366004611cba565b6101d6565b34801561011357600080fd5b506000546100ca9073ffffffffffffffffffffffffffffffffffffffff1681565b34801561014057600080fd5b5061016461014f366004611de6565b60016020526000908152604090205460ff1681565b60405190151581526020016100eb565b34801561018057600080fd5b5061008f61018f366004611e48565b6102cb565b3480156101a057600080fd5b506101c87f000000000000000000000000000000000000000000000000000000000000000081565b6040519081526020016100eb565b8180156101f8575073ffffffffffffffffffffffffffffffffffffffff871615155b1561022f576040517ff98844ef00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b33328114610250575033731111000000000000000000000000000000001111015b8773ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167fd9b5b1fc9eaff616a158a2011023083d44e5de459614dec8689632d8391c2312348a8a8a8a8a8a6040516102b99796959493929190611f48565b60405180910390a35050505050505050565b7f00000000000000000000000000000000000000000000000000000000000000008401421015610327576040517fe4750a3000000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6040517fa25ae557000000000000000000000000000000000000000000000000000000008152600481018590526000907f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff169063a25ae55790602401602060405180830381865afa1580156103b5573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906103d99190611ff9565b905061042d846040805182356020828101919091528301358183015290820135606082810191909152820135608082015260009060a001604051602081830303815290604052805190602001209050919050565b8114610465576040517f9cc00b5b00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b60006104768d8d8d8d8d8d8d610651565b90506104888186604001358686610693565b6104be576040517feb00eb2200000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b60008181526001602081905260409091205460ff161515141561050d576040517fae89945400000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b600180600083815260200190815260200160002060006101000a81548160ff0219169083151502179055508b6000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060008b73ffffffffffffffffffffffffffffffffffffffff168b8b908b8b6040516105a4929190612012565b600060405180830381858888f193505050503d80600081146105e2576040519150601f19603f3d011682016040523d82523d6000602084013e6105e7565b606091505b5050600080547fffffffffffffffffffffffff00000000000000000000000000000000000000001661dead17815560405191925083917f894485e328061b8d209b7dd043d2f613fc2892260497cadefac9a183962a990f9190a25050505050505050505050505050565b6000878787878787876040516020016106709796959493929190612022565b604051602081830303815290604052805190602001209050979650505050505050565b6000808560016040516020016106b3929190918252602082015260400190565b604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08184030181528282528051602091820120908301819052925061077991016040516020818303038152906040526040518060400160405280600181526020017f010000000000000000000000000000000000000000000000000000000000000081525086868080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152508b9250610783915050565b9695505050505050565b60008061078f8661079d565b9050610779818686866107cf565b606081805190602001206040516020016107b991815260200190565b6040516020818303038152906040529050919050565b60008060006107df878686610800565b915091508180156107f557506107f586826108fa565b979650505050505050565b60006060600061080f85610916565b90506000806000610821848a89610a11565b815192955090935091501580806108355750815b6108a0576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601a60248201527f50726f76696465642070726f6f6620697320696e76616c69642e00000000000060448201526064015b60405180910390fd5b6000816108bc57604051806020016040528060008152506108e8565b6108e8866108cb6001886120de565b815181106108db576108db6120f5565b6020026020010151610f2e565b919b919a509098505050505050505050565b6000818051906020012083805190602001201490505b92915050565b6060600061092383610f58565b90506000815167ffffffffffffffff81111561094157610941611c8b565b60405190808252806020026020018201604052801561098657816020015b604080518082019091526060808252602082015281526020019060019003908161095f5790505b50905060005b8251811015610a095760006109b98483815181106109ac576109ac6120f5565b6020026020010151610f8b565b905060405180604001604052808281526020016109d583610f58565b8152508383815181106109ea576109ea6120f5565b6020026020010181905250508080610a0190612124565b91505061098c565b509392505050565b60006060818080610a2187611035565b90506000869050600080610a48604051806040016040528060608152602001606081525090565b60005b8c51811015610eea578c8181518110610a6657610a666120f5565b602002602001015191508284610a7c919061215d565b9350610a8960018861215d565b965083610b0757815180516020909101208514610b02576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601160248201527f496e76616c696420726f6f7420686173680000000000000000000000000000006044820152606401610897565b610bf8565b815151602011610b8357815180516020909101208514610b02576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601b60248201527f496e76616c6964206c6172676520696e7465726e616c206861736800000000006044820152606401610897565b84610b9183600001516111b8565b14610bf8576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601a60248201527f496e76616c696420696e7465726e616c206e6f646520686173680000000000006044820152606401610897565b610c046010600161215d565b8260200151511415610c7d578551841415610c1e57610eea565b6000868581518110610c3257610c326120f5565b602001015160f81c60f81b60f81c9050600083602001518260ff1681518110610c5d57610c5d6120f5565b60200260200101519050610c70816111e0565b9650600194505050610ed8565b60028260200151511415610e76576000610c968361121d565b9050600081600081518110610cad57610cad6120f5565b016020015160f81c90506000610cc46002836121a4565b610ccf9060026121c6565b90506000610ce0848360ff16611241565b90506000610cee8b8a611241565b90506000610cfc8383611277565b905060ff851660021480610d13575060ff85166003145b15610d6957808351148015610d285750808251145b15610d3a57610d37818b61215d565b99505b507f80000000000000000000000000000000000000000000000000000000000000009950610eea945050505050565b60ff85161580610d7c575060ff85166001145b15610dee5782518114610db857507f80000000000000000000000000000000000000000000000000000000000000009950610eea945050505050565b610ddf8860200151600181518110610dd257610dd26120f5565b60200260200101516111e0565b9a509750610ed8945050505050565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f52656365697665642061206e6f6465207769746820616e20756e6b6e6f776e2060448201527f70726566697800000000000000000000000000000000000000000000000000006064820152608401610897565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f526563656976656420616e20756e706172736561626c65206e6f64652e0000006044820152606401610897565b80610ee281612124565b915050610a4b565b507f8000000000000000000000000000000000000000000000000000000000000000841486610f198786611241565b909e909d50909b509950505050505050505050565b6020810151805160609161091091610f48906001906120de565b815181106109ac576109ac6120f5565b60408051808201825260008082526020918201528151808301909252825182528083019082015260609061091090611323565b60606000806000610f9b85611556565b919450925090506000816001811115610fb657610fb66121e9565b1461101d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601860248201527f496e76616c696420524c502062797465732076616c75652e00000000000000006044820152606401610897565b61102c8560200151848461195d565b95945050505050565b60606000825160026110479190612218565b67ffffffffffffffff81111561105f5761105f611c8b565b6040519080825280601f01601f191660200182016040528015611089576020820181803683370190505b50905060005b83518110156111b15760048482815181106110ac576110ac6120f5565b01602001517fff0000000000000000000000000000000000000000000000000000000000000016901c826110e1836002612218565b815181106110f1576110f16120f5565b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a9053506010848281518110611134576111346120f5565b0160200151611146919060f81c6121a4565b60f81b82611155836002612218565b61116090600161215d565b81518110611170576111706120f5565b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a905350806111a981612124565b91505061108f565b5092915050565b60006020825110156111cc57506020015190565b818060200190518101906109109190611ff9565b60006060602083600001511015611201576111fa83611a3c565b905061120d565b61120a83610f8b565b90505b611216816111b8565b9392505050565b606061091061123c83602001516000815181106109ac576109ac6120f5565b611035565b6060825182106112605750604080516020810190915260008152610910565b611216838384865161127291906120de565b611a47565b6000805b80845111801561128b5750808351115b801561130c57508281815181106112a4576112a46120f5565b602001015160f81c60f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168482815181106112e3576112e36120f5565b01602001517fff0000000000000000000000000000000000000000000000000000000000000016145b15611216578061131b81612124565b91505061127b565b606060008061133184611556565b9193509091506001905081600181111561134d5761134d6121e9565b146113b4576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601760248201527f496e76616c696420524c50206c6973742076616c75652e0000000000000000006044820152606401610897565b6040805160208082526104208201909252600091816020015b60408051808201909152600080825260208201528152602001906001900390816113cd5790505090506000835b865181101561154b5760208210611493576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602a60248201527f50726f766964656420524c50206c6973742065786365656473206d6178206c6960448201527f7374206c656e6774682e000000000000000000000000000000000000000000006064820152608401610897565b6000806114d06040518060400160405280858c600001516114b491906120de565b8152602001858c602001516114c9919061215d565b9052611556565b5091509150604051806040016040528083836114ec919061215d565b8152602001848b60200151611501919061215d565b815250858581518110611516576115166120f5565b602090810291909101015261152c60018561215d565b9350611538818361215d565b611542908461215d565b925050506113fa565b508152949350505050565b6000806000808460000151116115c8576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601860248201527f524c50206974656d2063616e6e6f74206265206e756c6c2e00000000000000006044820152606401610897565b6020840151805160001a607f81116115ed576000600160009450945094505050611956565b60b781116116835760006116026080836120de565b905080876000015111611671576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601960248201527f496e76616c696420524c502073686f727420737472696e672e000000000000006044820152606401610897565b60019550935060009250611956915050565b60bf81116117a657600061169860b7836120de565b905080876000015111611707576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f496e76616c696420524c50206c6f6e6720737472696e67206c656e6774682e006044820152606401610897565b600183015160208290036101000a9004611721818361215d565b88511161178a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601860248201527f496e76616c696420524c50206c6f6e6720737472696e672e00000000000000006044820152606401610897565b61179582600161215d565b965094506000935061195692505050565b60f7811161183b5760006117bb60c0836120de565b90508087600001511161182a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601760248201527f496e76616c696420524c502073686f7274206c6973742e0000000000000000006044820152606401610897565b600195509350849250611956915050565b600061184860f7836120de565b9050808760000151116118b7576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f496e76616c696420524c50206c6f6e67206c697374206c656e6774682e0000006044820152606401610897565b600183015160208290036101000a90046118d1818361215d565b88511161193a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601660248201527f496e76616c696420524c50206c6f6e67206c6973742e000000000000000000006044820152606401610897565b61194582600161215d565b965094506001935061195692505050565b9193909250565b606060008267ffffffffffffffff81111561197a5761197a611c8b565b6040519080825280601f01601f1916602001820160405280156119a4576020820181803683370190505b5090508051600014156119b8579050611216565b60006119c4858761215d565b90506020820160005b6119d8602087612255565b811015611a0f57825182526119ee60208461215d565b92506119fb60208361215d565b915080611a0781612124565b9150506119cd565b5060006001602087066020036101000a039050808251168119845116178252839450505050509392505050565b606061091082611c34565b606081611a5581601f61215d565b1015611abd576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152600e60248201527f736c6963655f6f766572666c6f770000000000000000000000000000000000006044820152606401610897565b82611ac8838261215d565b1015611b30576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152600e60248201527f736c6963655f6f766572666c6f770000000000000000000000000000000000006044820152606401610897565b611b3a828461215d565b84511015611ba4576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601160248201527f736c6963655f6f75744f66426f756e64730000000000000000000000000000006044820152606401610897565b606082158015611bc35760405191506000825260208201604052611c2b565b6040519150601f8416801560200281840101858101878315602002848b0101015b81831015611bfc578051835260209283019201611be4565b5050858452601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016604052505b50949350505050565b606061091082602001516000846000015161195d565b803573ffffffffffffffffffffffffffffffffffffffff81168114611c6e57600080fd5b919050565b803567ffffffffffffffff81168114611c6e57600080fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b600080600080600080600060e0888a031215611cd557600080fd5b611cde88611c4a565b96506020880135955060408801359450611cfa60608901611c73565b9350611d0860808901611c73565b925060a08801358015158114611d1d57600080fd5b915060c088013567ffffffffffffffff80821115611d3a57600080fd5b818a0191508a601f830112611d4e57600080fd5b813581811115611d6057611d60611c8b565b604051601f82017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0908116603f01168101908382118183101715611da657611da6611c8b565b816040528281528d6020848701011115611dbf57600080fd5b82602086016020830137600060208483010152809550505050505092959891949750929550565b600060208284031215611df857600080fd5b5035919050565b60008083601f840112611e1157600080fd5b50813567ffffffffffffffff811115611e2957600080fd5b602083019150836020828501011115611e4157600080fd5b9250929050565b60008060008060008060008060008060006101808c8e031215611e6a57600080fd5b8b359a50611e7a60208d01611c4a565b9950611e8860408d01611c4a565b985060608c0135975060808c0135965067ffffffffffffffff60a08d01351115611eb157600080fd5b611ec18d60a08e01358e01611dff565b909650945060c08c0135935060808c8e037fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff20011215611eff57600080fd5b60e08c01925067ffffffffffffffff6101608d01351115611f1f57600080fd5b611f308d6101608e01358e01611dff565b81935080925050509295989b509295989b9093969950565b87815260006020888184015287604084015267ffffffffffffffff808816606085015280871660808501525084151560a084015260e060c084015283518060e085015260005b81811015611fab5785810183015185820161010001528201611f8e565b81811115611fbe57600061010083870101525b50601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe01692909201610100019998505050505050505050565b60006020828403121561200b57600080fd5b5051919050565b8183823760009101908152919050565b878152600073ffffffffffffffffffffffffffffffffffffffff808916602084015280881660408401525085606083015284608083015260c060a08301528260c0830152828460e0840137600060e0848401015260e07fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0601f850116830101905098975050505050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000828210156120f0576120f06120af565b500390565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff821415612156576121566120af565b5060010190565b60008219821115612170576121706120af565b500190565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b600060ff8316806121b7576121b7612175565b8060ff84160691505092915050565b600060ff821660ff8416808210156121e0576121e06120af565b90039392505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b6000817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0483118215151615612250576122506120af565b500290565b60008261226457612264612175565b50049056fea164736f6c634300080a000a",
}

// OptimismPortalABI is the input ABI used to generate the binding from.
// Deprecated: Use OptimismPortalMetaData.ABI instead.
var OptimismPortalABI = OptimismPortalMetaData.ABI

// OptimismPortalBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use OptimismPortalMetaData.Bin instead.
var OptimismPortalBin = OptimismPortalMetaData.Bin

// DeployOptimismPortal deploys a new Ethereum contract, binding an instance of OptimismPortal to it.
func DeployOptimismPortal(auth *bind.TransactOpts, backend bind.ContractBackend, _l2Oracle common.Address, _finalizationPeriod *big.Int) (common.Address, *types.Transaction, *OptimismPortal, error) {
	parsed, err := OptimismPortalMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(OptimismPortalBin), backend, _l2Oracle, _finalizationPeriod)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &OptimismPortal{OptimismPortalCaller: OptimismPortalCaller{contract: contract}, OptimismPortalTransactor: OptimismPortalTransactor{contract: contract}, OptimismPortalFilterer: OptimismPortalFilterer{contract: contract}}, nil
}

// OptimismPortal is an auto generated Go binding around an Ethereum contract.
type OptimismPortal struct {
	OptimismPortalCaller     // Read-only binding to the contract
	OptimismPortalTransactor // Write-only binding to the contract
	OptimismPortalFilterer   // Log filterer for contract events
}

// OptimismPortalCaller is an auto generated read-only Go binding around an Ethereum contract.
type OptimismPortalCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OptimismPortalTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OptimismPortalTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OptimismPortalFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OptimismPortalFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OptimismPortalSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OptimismPortalSession struct {
	Contract     *OptimismPortal   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OptimismPortalCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OptimismPortalCallerSession struct {
	Contract *OptimismPortalCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// OptimismPortalTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OptimismPortalTransactorSession struct {
	Contract     *OptimismPortalTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// OptimismPortalRaw is an auto generated low-level Go binding around an Ethereum contract.
type OptimismPortalRaw struct {
	Contract *OptimismPortal // Generic contract binding to access the raw methods on
}

// OptimismPortalCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OptimismPortalCallerRaw struct {
	Contract *OptimismPortalCaller // Generic read-only contract binding to access the raw methods on
}

// OptimismPortalTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OptimismPortalTransactorRaw struct {
	Contract *OptimismPortalTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOptimismPortal creates a new instance of OptimismPortal, bound to a specific deployed contract.
func NewOptimismPortal(address common.Address, backend bind.ContractBackend) (*OptimismPortal, error) {
	contract, err := bindOptimismPortal(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OptimismPortal{OptimismPortalCaller: OptimismPortalCaller{contract: contract}, OptimismPortalTransactor: OptimismPortalTransactor{contract: contract}, OptimismPortalFilterer: OptimismPortalFilterer{contract: contract}}, nil
}

// NewOptimismPortalCaller creates a new read-only instance of OptimismPortal, bound to a specific deployed contract.
func NewOptimismPortalCaller(address common.Address, caller bind.ContractCaller) (*OptimismPortalCaller, error) {
	contract, err := bindOptimismPortal(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OptimismPortalCaller{contract: contract}, nil
}

// NewOptimismPortalTransactor creates a new write-only instance of OptimismPortal, bound to a specific deployed contract.
func NewOptimismPortalTransactor(address common.Address, transactor bind.ContractTransactor) (*OptimismPortalTransactor, error) {
	contract, err := bindOptimismPortal(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OptimismPortalTransactor{contract: contract}, nil
}

// NewOptimismPortalFilterer creates a new log filterer instance of OptimismPortal, bound to a specific deployed contract.
func NewOptimismPortalFilterer(address common.Address, filterer bind.ContractFilterer) (*OptimismPortalFilterer, error) {
	contract, err := bindOptimismPortal(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OptimismPortalFilterer{contract: contract}, nil
}

// bindOptimismPortal binds a generic wrapper to an already deployed contract.
func bindOptimismPortal(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(OptimismPortalABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OptimismPortal *OptimismPortalRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OptimismPortal.Contract.OptimismPortalCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OptimismPortal *OptimismPortalRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OptimismPortal.Contract.OptimismPortalTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OptimismPortal *OptimismPortalRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OptimismPortal.Contract.OptimismPortalTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OptimismPortal *OptimismPortalCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OptimismPortal.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OptimismPortal *OptimismPortalTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OptimismPortal.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OptimismPortal *OptimismPortalTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OptimismPortal.Contract.contract.Transact(opts, method, params...)
}

// FINALIZATIONPERIOD is a free data retrieval call binding the contract method 0xff61cc93.
//
// Solidity: function FINALIZATION_PERIOD() view returns(uint256)
func (_OptimismPortal *OptimismPortalCaller) FINALIZATIONPERIOD(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _OptimismPortal.contract.Call(opts, &out, "FINALIZATION_PERIOD")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FINALIZATIONPERIOD is a free data retrieval call binding the contract method 0xff61cc93.
//
// Solidity: function FINALIZATION_PERIOD() view returns(uint256)
func (_OptimismPortal *OptimismPortalSession) FINALIZATIONPERIOD() (*big.Int, error) {
	return _OptimismPortal.Contract.FINALIZATIONPERIOD(&_OptimismPortal.CallOpts)
}

// FINALIZATIONPERIOD is a free data retrieval call binding the contract method 0xff61cc93.
//
// Solidity: function FINALIZATION_PERIOD() view returns(uint256)
func (_OptimismPortal *OptimismPortalCallerSession) FINALIZATIONPERIOD() (*big.Int, error) {
	return _OptimismPortal.Contract.FINALIZATIONPERIOD(&_OptimismPortal.CallOpts)
}

// L2ORACLE is a free data retrieval call binding the contract method 0x001c2ff6.
//
// Solidity: function L2_ORACLE() view returns(address)
func (_OptimismPortal *OptimismPortalCaller) L2ORACLE(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OptimismPortal.contract.Call(opts, &out, "L2_ORACLE")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L2ORACLE is a free data retrieval call binding the contract method 0x001c2ff6.
//
// Solidity: function L2_ORACLE() view returns(address)
func (_OptimismPortal *OptimismPortalSession) L2ORACLE() (common.Address, error) {
	return _OptimismPortal.Contract.L2ORACLE(&_OptimismPortal.CallOpts)
}

// L2ORACLE is a free data retrieval call binding the contract method 0x001c2ff6.
//
// Solidity: function L2_ORACLE() view returns(address)
func (_OptimismPortal *OptimismPortalCallerSession) L2ORACLE() (common.Address, error) {
	return _OptimismPortal.Contract.L2ORACLE(&_OptimismPortal.CallOpts)
}

// FinalizedWithdrawals is a free data retrieval call binding the contract method 0xa14238e7.
//
// Solidity: function finalizedWithdrawals(bytes32 ) view returns(bool)
func (_OptimismPortal *OptimismPortalCaller) FinalizedWithdrawals(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _OptimismPortal.contract.Call(opts, &out, "finalizedWithdrawals", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// FinalizedWithdrawals is a free data retrieval call binding the contract method 0xa14238e7.
//
// Solidity: function finalizedWithdrawals(bytes32 ) view returns(bool)
func (_OptimismPortal *OptimismPortalSession) FinalizedWithdrawals(arg0 [32]byte) (bool, error) {
	return _OptimismPortal.Contract.FinalizedWithdrawals(&_OptimismPortal.CallOpts, arg0)
}

// FinalizedWithdrawals is a free data retrieval call binding the contract method 0xa14238e7.
//
// Solidity: function finalizedWithdrawals(bytes32 ) view returns(bool)
func (_OptimismPortal *OptimismPortalCallerSession) FinalizedWithdrawals(arg0 [32]byte) (bool, error) {
	return _OptimismPortal.Contract.FinalizedWithdrawals(&_OptimismPortal.CallOpts, arg0)
}

// L2Sender is a free data retrieval call binding the contract method 0x9bf62d82.
//
// Solidity: function l2Sender() view returns(address)
func (_OptimismPortal *OptimismPortalCaller) L2Sender(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OptimismPortal.contract.Call(opts, &out, "l2Sender")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L2Sender is a free data retrieval call binding the contract method 0x9bf62d82.
//
// Solidity: function l2Sender() view returns(address)
func (_OptimismPortal *OptimismPortalSession) L2Sender() (common.Address, error) {
	return _OptimismPortal.Contract.L2Sender(&_OptimismPortal.CallOpts)
}

// L2Sender is a free data retrieval call binding the contract method 0x9bf62d82.
//
// Solidity: function l2Sender() view returns(address)
func (_OptimismPortal *OptimismPortalCallerSession) L2Sender() (common.Address, error) {
	return _OptimismPortal.Contract.L2Sender(&_OptimismPortal.CallOpts)
}

// DepositTransaction is a paid mutator transaction binding the contract method 0x54314103.
//
// Solidity: function depositTransaction(address _to, uint256 _value, uint256 _additionalGasPrice, uint64 _additionalGasLimit, uint64 _guaranteedGas, bool _isCreation, bytes _data) payable returns()
func (_OptimismPortal *OptimismPortalTransactor) DepositTransaction(opts *bind.TransactOpts, _to common.Address, _value *big.Int, _additionalGasPrice *big.Int, _additionalGasLimit uint64, _guaranteedGas uint64, _isCreation bool, _data []byte) (*types.Transaction, error) {
	return _OptimismPortal.contract.Transact(opts, "depositTransaction", _to, _value, _additionalGasPrice, _additionalGasLimit, _guaranteedGas, _isCreation, _data)
}

// DepositTransaction is a paid mutator transaction binding the contract method 0x54314103.
//
// Solidity: function depositTransaction(address _to, uint256 _value, uint256 _additionalGasPrice, uint64 _additionalGasLimit, uint64 _guaranteedGas, bool _isCreation, bytes _data) payable returns()
func (_OptimismPortal *OptimismPortalSession) DepositTransaction(_to common.Address, _value *big.Int, _additionalGasPrice *big.Int, _additionalGasLimit uint64, _guaranteedGas uint64, _isCreation bool, _data []byte) (*types.Transaction, error) {
	return _OptimismPortal.Contract.DepositTransaction(&_OptimismPortal.TransactOpts, _to, _value, _additionalGasPrice, _additionalGasLimit, _guaranteedGas, _isCreation, _data)
}

// DepositTransaction is a paid mutator transaction binding the contract method 0x54314103.
//
// Solidity: function depositTransaction(address _to, uint256 _value, uint256 _additionalGasPrice, uint64 _additionalGasLimit, uint64 _guaranteedGas, bool _isCreation, bytes _data) payable returns()
func (_OptimismPortal *OptimismPortalTransactorSession) DepositTransaction(_to common.Address, _value *big.Int, _additionalGasPrice *big.Int, _additionalGasLimit uint64, _guaranteedGas uint64, _isCreation bool, _data []byte) (*types.Transaction, error) {
	return _OptimismPortal.Contract.DepositTransaction(&_OptimismPortal.TransactOpts, _to, _value, _additionalGasPrice, _additionalGasLimit, _guaranteedGas, _isCreation, _data)
}

// FinalizeWithdrawalTransaction is a paid mutator transaction binding the contract method 0xeecf1c36.
//
// Solidity: function finalizeWithdrawalTransaction(uint256 _nonce, address _sender, address _target, uint256 _value, uint256 _gasLimit, bytes _data, uint256 _timestamp, (bytes32,bytes32,bytes32,bytes32) _outputRootProof, bytes _withdrawalProof) returns()
func (_OptimismPortal *OptimismPortalTransactor) FinalizeWithdrawalTransaction(opts *bind.TransactOpts, _nonce *big.Int, _sender common.Address, _target common.Address, _value *big.Int, _gasLimit *big.Int, _data []byte, _timestamp *big.Int, _outputRootProof WithdrawalVerifierOutputRootProof, _withdrawalProof []byte) (*types.Transaction, error) {
	return _OptimismPortal.contract.Transact(opts, "finalizeWithdrawalTransaction", _nonce, _sender, _target, _value, _gasLimit, _data, _timestamp, _outputRootProof, _withdrawalProof)
}

// FinalizeWithdrawalTransaction is a paid mutator transaction binding the contract method 0xeecf1c36.
//
// Solidity: function finalizeWithdrawalTransaction(uint256 _nonce, address _sender, address _target, uint256 _value, uint256 _gasLimit, bytes _data, uint256 _timestamp, (bytes32,bytes32,bytes32,bytes32) _outputRootProof, bytes _withdrawalProof) returns()
func (_OptimismPortal *OptimismPortalSession) FinalizeWithdrawalTransaction(_nonce *big.Int, _sender common.Address, _target common.Address, _value *big.Int, _gasLimit *big.Int, _data []byte, _timestamp *big.Int, _outputRootProof WithdrawalVerifierOutputRootProof, _withdrawalProof []byte) (*types.Transaction, error) {
	return _OptimismPortal.Contract.FinalizeWithdrawalTransaction(&_OptimismPortal.TransactOpts, _nonce, _sender, _target, _value, _gasLimit, _data, _timestamp, _outputRootProof, _withdrawalProof)
}

// FinalizeWithdrawalTransaction is a paid mutator transaction binding the contract method 0xeecf1c36.
//
// Solidity: function finalizeWithdrawalTransaction(uint256 _nonce, address _sender, address _target, uint256 _value, uint256 _gasLimit, bytes _data, uint256 _timestamp, (bytes32,bytes32,bytes32,bytes32) _outputRootProof, bytes _withdrawalProof) returns()
func (_OptimismPortal *OptimismPortalTransactorSession) FinalizeWithdrawalTransaction(_nonce *big.Int, _sender common.Address, _target common.Address, _value *big.Int, _gasLimit *big.Int, _data []byte, _timestamp *big.Int, _outputRootProof WithdrawalVerifierOutputRootProof, _withdrawalProof []byte) (*types.Transaction, error) {
	return _OptimismPortal.Contract.FinalizeWithdrawalTransaction(&_OptimismPortal.TransactOpts, _nonce, _sender, _target, _value, _gasLimit, _data, _timestamp, _outputRootProof, _withdrawalProof)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_OptimismPortal *OptimismPortalTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OptimismPortal.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_OptimismPortal *OptimismPortalSession) Receive() (*types.Transaction, error) {
	return _OptimismPortal.Contract.Receive(&_OptimismPortal.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_OptimismPortal *OptimismPortalTransactorSession) Receive() (*types.Transaction, error) {
	return _OptimismPortal.Contract.Receive(&_OptimismPortal.TransactOpts)
}

// OptimismPortalTransactionDepositedIterator is returned from FilterTransactionDeposited and is used to iterate over the raw logs and unpacked data for TransactionDeposited events raised by the OptimismPortal contract.
type OptimismPortalTransactionDepositedIterator struct {
	Event *OptimismPortalTransactionDeposited // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OptimismPortalTransactionDepositedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OptimismPortalTransactionDeposited)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OptimismPortalTransactionDeposited)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OptimismPortalTransactionDepositedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OptimismPortalTransactionDepositedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OptimismPortalTransactionDeposited represents a TransactionDeposited event raised by the OptimismPortal contract.
type OptimismPortalTransactionDeposited struct {
	From               common.Address
	To                 common.Address
	Mint               *big.Int
	Value              *big.Int
	AdditionalGasPrice *big.Int
	AdditionalGasLimit uint64
	GuaranteedGas      uint64
	IsCreation         bool
	Data               []byte
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterTransactionDeposited is a free log retrieval operation binding the contract event 0xd9b5b1fc9eaff616a158a2011023083d44e5de459614dec8689632d8391c2312.
//
// Solidity: event TransactionDeposited(address indexed from, address indexed to, uint256 mint, uint256 value, uint256 additionalGasPrice, uint64 additionalGasLimit, uint64 guaranteedGas, bool isCreation, bytes data)
func (_OptimismPortal *OptimismPortalFilterer) FilterTransactionDeposited(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*OptimismPortalTransactionDepositedIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _OptimismPortal.contract.FilterLogs(opts, "TransactionDeposited", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &OptimismPortalTransactionDepositedIterator{contract: _OptimismPortal.contract, event: "TransactionDeposited", logs: logs, sub: sub}, nil
}

// WatchTransactionDeposited is a free log subscription operation binding the contract event 0xd9b5b1fc9eaff616a158a2011023083d44e5de459614dec8689632d8391c2312.
//
// Solidity: event TransactionDeposited(address indexed from, address indexed to, uint256 mint, uint256 value, uint256 additionalGasPrice, uint64 additionalGasLimit, uint64 guaranteedGas, bool isCreation, bytes data)
func (_OptimismPortal *OptimismPortalFilterer) WatchTransactionDeposited(opts *bind.WatchOpts, sink chan<- *OptimismPortalTransactionDeposited, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _OptimismPortal.contract.WatchLogs(opts, "TransactionDeposited", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OptimismPortalTransactionDeposited)
				if err := _OptimismPortal.contract.UnpackLog(event, "TransactionDeposited", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransactionDeposited is a log parse operation binding the contract event 0xd9b5b1fc9eaff616a158a2011023083d44e5de459614dec8689632d8391c2312.
//
// Solidity: event TransactionDeposited(address indexed from, address indexed to, uint256 mint, uint256 value, uint256 additionalGasPrice, uint64 additionalGasLimit, uint64 guaranteedGas, bool isCreation, bytes data)
func (_OptimismPortal *OptimismPortalFilterer) ParseTransactionDeposited(log types.Log) (*OptimismPortalTransactionDeposited, error) {
	event := new(OptimismPortalTransactionDeposited)
	if err := _OptimismPortal.contract.UnpackLog(event, "TransactionDeposited", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OptimismPortalWithdrawalFinalizedIterator is returned from FilterWithdrawalFinalized and is used to iterate over the raw logs and unpacked data for WithdrawalFinalized events raised by the OptimismPortal contract.
type OptimismPortalWithdrawalFinalizedIterator struct {
	Event *OptimismPortalWithdrawalFinalized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OptimismPortalWithdrawalFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OptimismPortalWithdrawalFinalized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OptimismPortalWithdrawalFinalized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OptimismPortalWithdrawalFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OptimismPortalWithdrawalFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OptimismPortalWithdrawalFinalized represents a WithdrawalFinalized event raised by the OptimismPortal contract.
type OptimismPortalWithdrawalFinalized struct {
	Arg0 [32]byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterWithdrawalFinalized is a free log retrieval operation binding the contract event 0x894485e328061b8d209b7dd043d2f613fc2892260497cadefac9a183962a990f.
//
// Solidity: event WithdrawalFinalized(bytes32 indexed arg0)
func (_OptimismPortal *OptimismPortalFilterer) FilterWithdrawalFinalized(opts *bind.FilterOpts, arg0 [][32]byte) (*OptimismPortalWithdrawalFinalizedIterator, error) {

	var arg0Rule []interface{}
	for _, arg0Item := range arg0 {
		arg0Rule = append(arg0Rule, arg0Item)
	}

	logs, sub, err := _OptimismPortal.contract.FilterLogs(opts, "WithdrawalFinalized", arg0Rule)
	if err != nil {
		return nil, err
	}
	return &OptimismPortalWithdrawalFinalizedIterator{contract: _OptimismPortal.contract, event: "WithdrawalFinalized", logs: logs, sub: sub}, nil
}

// WatchWithdrawalFinalized is a free log subscription operation binding the contract event 0x894485e328061b8d209b7dd043d2f613fc2892260497cadefac9a183962a990f.
//
// Solidity: event WithdrawalFinalized(bytes32 indexed arg0)
func (_OptimismPortal *OptimismPortalFilterer) WatchWithdrawalFinalized(opts *bind.WatchOpts, sink chan<- *OptimismPortalWithdrawalFinalized, arg0 [][32]byte) (event.Subscription, error) {

	var arg0Rule []interface{}
	for _, arg0Item := range arg0 {
		arg0Rule = append(arg0Rule, arg0Item)
	}

	logs, sub, err := _OptimismPortal.contract.WatchLogs(opts, "WithdrawalFinalized", arg0Rule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OptimismPortalWithdrawalFinalized)
				if err := _OptimismPortal.contract.UnpackLog(event, "WithdrawalFinalized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWithdrawalFinalized is a log parse operation binding the contract event 0x894485e328061b8d209b7dd043d2f613fc2892260497cadefac9a183962a990f.
//
// Solidity: event WithdrawalFinalized(bytes32 indexed arg0)
func (_OptimismPortal *OptimismPortalFilterer) ParseWithdrawalFinalized(log types.Log) (*OptimismPortalWithdrawalFinalized, error) {
	event := new(OptimismPortalWithdrawalFinalized)
	if err := _OptimismPortal.contract.UnpackLog(event, "WithdrawalFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
