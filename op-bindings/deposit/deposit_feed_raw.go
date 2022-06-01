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
	ABI: "[{\"inputs\":[{\"internalType\":\"contractL2OutputOracle\",\"name\":\"_l2Oracle\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_finalizationPeriod\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"InvalidOutputRootProof\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidWithdrawalInclusionProof\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NonZeroCreationTarget\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"WithdrawalAlreadyFinalized\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"mint\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"gasLimit\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"isCreation\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"TransactionDeposited\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"name\":\"WithdrawalFinalized\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"FINALIZATION_PERIOD\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"L2_ORACLE\",\"outputs\":[{\"internalType\":\"contractL2OutputOracle\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"},{\"internalType\":\"uint64\",\"name\":\"_gasLimit\",\"type\":\"uint64\"},{\"internalType\":\"bool\",\"name\":\"_isCreation\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"depositTransaction\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_gasLimit\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"_l2Timestamp\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"version\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"withdrawerStorageRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"latestBlockhash\",\"type\":\"bytes32\"}],\"internalType\":\"structWithdrawalVerifier.OutputRootProof\",\"name\":\"_outputRootProof\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"_withdrawalProof\",\"type\":\"bytes\"}],\"name\":\"finalizeWithdrawalTransaction\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"finalizedWithdrawals\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"l2Sender\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x60c0604052600080546001600160a01b03191661dead1790553480156200002557600080fd5b50604051620026aa380380620026aa83398101604081905262000048916200005f565b6001600160a01b0390911660a0526080526200009b565b600080604083850312156200007357600080fd5b82516001600160a01b03811681146200008b57600080fd5b6020939093015192949293505050565b60805160a0516125dc620000ce6000396000818160a6015261036d01526000818161019701526103f001526125dc6000f3fe6080604052600436106100685760003560e01c8063e9e05c4211610043578063e9e05c421461015f578063eecf1c3614610172578063ff61cc931461018557600080fd5b80621c2ff6146100945780639bf62d82146100f2578063a14238e71461011f57600080fd5b3661008f5761008d3334620186a06000604051806020016040528060008152506101c7565b005b600080fd5b3480156100a057600080fd5b506100c87f000000000000000000000000000000000000000000000000000000000000000081565b60405173ffffffffffffffffffffffffffffffffffffffff90911681526020015b60405180910390f35b3480156100fe57600080fd5b506000546100c89073ffffffffffffffffffffffffffffffffffffffff1681565b34801561012b57600080fd5b5061014f61013a366004611f44565b60016020526000908152604090205460ff1681565b60405190151581526020016100e9565b61008d61016d366004612004565b6101c7565b61008d61018036600461214e565b6102b6565b34801561019157600080fd5b506101b97f000000000000000000000000000000000000000000000000000000000000000081565b6040519081526020016100e9565b8180156101e9575073ffffffffffffffffffffffffffffffffffffffff851615155b15610220576040517ff98844ef00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b33328114610241575033731111000000000000000000000000000000001111015b8573ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f78231ae6eb73366f912bb1d64351601fb76344c537bbab635ce14d0f376f019534888888886040516102a69594939291906122b9565b60405180910390a3505050505050565b73ffffffffffffffffffffffffffffffffffffffff891630141561033b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601c60248201527f43616e6e6f742073656e64206d65737361676520746f2073656c662e0000000060448201526064015b60405180910390fd5b6040517fa25ae557000000000000000000000000000000000000000000000000000000008152600481018590526000907f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff169063a25ae557906024016040805180830381865afa1580156103c8573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906103ec91906122f0565b90507f0000000000000000000000000000000000000000000000000000000000000000816020015161041e919061236e565b4211610486576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601e60248201527f50726f706f73616c206973206e6f74207965742066696e616c697a65642e00006044820152606401610332565b61049d61049836869003860186612386565b610890565b8151146104d6576040517f9cc00b5b00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b600061051c8d8d8d8d8d8d8d8080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152506108ec92505050565b905061056381866040013586868080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061092b92505050565b610599576040517feb00eb2200000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b60008181526001602081905260409091205460ff16151514156105e8576040517fae89945400000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b600081815260016020819052604090912080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0016909117905561062e614e208a61236e565b5a10156106bd576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602860248201527f496e73756666696369656e742067617320746f2066696e616c697a652077697460448201527f6864726177616c2e0000000000000000000000000000000000000000000000006064820152608401610332565b60005473ffffffffffffffffffffffffffffffffffffffff1661dead1461078d576040517f08c379a0000000000000000000000000000000000000000000000000000000008152602060048201526044602482018190527f43616e6e6f742077697468647261772061207472616e73616374696f6e207769908201527f7468696e20616e6f74686572207769746864726177616c207472616e7361637460648201527f696f6e2e00000000000000000000000000000000000000000000000000000000608482015260a401610332565b600080547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff8e16178155604080516020601f8b01819004810282018101909252898152610812918e918d918f9186918f908f90819084018382808284376000920191909152506109f492505050565b50600080547fffffffffffffffffffffffff00000000000000000000000000000000000000001661dead17905560405190915082907fdb5c7652857aa163daadd670e116628fb42e869d8ac4251ef8971d9e5727df1b9061087890841515815260200190565b60405180910390a25050505050505050505050505050565b600081600001518260200151836040015184606001516040516020016108cf949392919093845260208401929092526040830152606082015260800190565b604051602081830303815290604052805190602001209050919050565b6000868686868686604051602001610909969594939291906123ec565b6040516020818303038152906040528051906020012090509695505050505050565b604080516020810185905260009181018290528190606001604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0818403018152828252805160209182012090830181905292506109e99101604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0818403018152828201909152600182527f01000000000000000000000000000000000000000000000000000000000000006020830152908587610a7f565b9150505b9392505050565b6000606060008060008661ffff1667ffffffffffffffff811115610a1a57610a1a611f86565b6040519080825280601f01601f191660200182016040528015610a44576020820181803683370190505b5090506000808751602089018b8e8ef191503d925086831115610a65578692505b828152826000602083013e90999098509650505050505050565b600080610a8b86610aa3565b9050610a9981868686610ad5565b9695505050505050565b60608180519060200120604051602001610abf91815260200190565b6040516020818303038152906040529050919050565b6000806000610ae5878686610b06565b91509150818015610afb5750610afb8682610bfb565b979650505050505050565b600060606000610b1585610c17565b90506000806000610b27848a89610d12565b81519295509093509150158080610b3b5750815b610ba1576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601a60248201527f50726f76696465642070726f6f6620697320696e76616c69642e0000000000006044820152606401610332565b600081610bbd5760405180602001604052806000815250610be9565b610be986610bcc600188612443565b81518110610bdc57610bdc61245a565b602002602001015161122f565b919b919a509098505050505050505050565b6000818051906020012083805190602001201490505b92915050565b60606000610c2483611259565b90506000815167ffffffffffffffff811115610c4257610c42611f86565b604051908082528060200260200182016040528015610c8757816020015b6040805180820190915260608082526020820152815260200190600190039081610c605790505b50905060005b8251811015610d0a576000610cba848381518110610cad57610cad61245a565b602002602001015161128c565b90506040518060400160405280828152602001610cd683611259565b815250838381518110610ceb57610ceb61245a565b6020026020010181905250508080610d0290612489565b915050610c8d565b509392505050565b60006060818080610d2287611336565b90506000869050600080610d49604051806040016040528060608152602001606081525090565b60005b8c518110156111eb578c8181518110610d6757610d6761245a565b602002602001015191508284610d7d919061236e565b9350610d8a60018861236e565b965083610e0857815180516020909101208514610e03576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601160248201527f496e76616c696420726f6f7420686173680000000000000000000000000000006044820152606401610332565b610ef9565b815151602011610e8457815180516020909101208514610e03576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601b60248201527f496e76616c6964206c6172676520696e7465726e616c206861736800000000006044820152606401610332565b84610e9283600001516114b9565b14610ef9576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601a60248201527f496e76616c696420696e7465726e616c206e6f646520686173680000000000006044820152606401610332565b610f056010600161236e565b8260200151511415610f7e578551841415610f1f576111eb565b6000868581518110610f3357610f3361245a565b602001015160f81c60f81b60f81c9050600083602001518260ff1681518110610f5e57610f5e61245a565b60200260200101519050610f71816114e1565b96506001945050506111d9565b60028260200151511415611177576000610f9783611517565b9050600081600081518110610fae57610fae61245a565b016020015160f81c90506000610fc56002836124f1565b610fd0906002612513565b90506000610fe1848360ff1661153b565b90506000610fef8b8a61153b565b90506000610ffd8383611571565b905060ff851660021480611014575060ff85166003145b1561106a578083511480156110295750808251145b1561103b57611038818b61236e565b99505b507f800000000000000000000000000000000000000000000000000000000000000099506111eb945050505050565b60ff8516158061107d575060ff85166001145b156110ef57825181146110b957507f800000000000000000000000000000000000000000000000000000000000000099506111eb945050505050565b6110e088602001516001815181106110d3576110d361245a565b60200260200101516114e1565b9a5097506111d9945050505050565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f52656365697665642061206e6f6465207769746820616e20756e6b6e6f776e2060448201527f70726566697800000000000000000000000000000000000000000000000000006064820152608401610332565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f526563656976656420616e20756e706172736561626c65206e6f64652e0000006044820152606401610332565b806111e381612489565b915050610d4c565b507f800000000000000000000000000000000000000000000000000000000000000084148661121a878661153b565b909e909d50909b509950505050505050505050565b60208101518051606091610c119161124990600190612443565b81518110610cad57610cad61245a565b604080518082018252600080825260209182015281518083019092528251825280830190820152606090610c119061161d565b6060600080600061129c85611850565b9194509250905060008160018111156112b7576112b7612536565b1461131e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601860248201527f496e76616c696420524c502062797465732076616c75652e00000000000000006044820152606401610332565b61132d85602001518484611c57565b95945050505050565b60606000825160026113489190612565565b67ffffffffffffffff81111561136057611360611f86565b6040519080825280601f01601f19166020018201604052801561138a576020820181803683370190505b50905060005b83518110156114b25760048482815181106113ad576113ad61245a565b01602001517fff0000000000000000000000000000000000000000000000000000000000000016901c826113e2836002612565565b815181106113f2576113f261245a565b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a90535060108482815181106114355761143561245a565b0160200151611447919060f81c6124f1565b60f81b82611456836002612565565b61146190600161236e565b815181106114715761147161245a565b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a905350806114aa81612489565b915050611390565b5092915050565b60006020825110156114cd57506020015190565b81806020019051810190610c1191906125a2565b60006060602083600001511015611502576114fb83611d36565b905061150e565b61150b8361128c565b90505b6109ed816114b9565b6060610c116115368360200151600081518110610cad57610cad61245a565b611336565b60608251821061155a5750604080516020810190915260008152610c11565b6109ed838384865161156c9190612443565b611d41565b6000805b8084511180156115855750808351115b8015611606575082818151811061159e5761159e61245a565b602001015160f81c60f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168482815181106115dd576115dd61245a565b01602001517fff0000000000000000000000000000000000000000000000000000000000000016145b156109ed578061161581612489565b915050611575565b606060008061162b84611850565b9193509091506001905081600181111561164757611647612536565b146116ae576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601760248201527f496e76616c696420524c50206c6973742076616c75652e0000000000000000006044820152606401610332565b6040805160208082526104208201909252600091816020015b60408051808201909152600080825260208201528152602001906001900390816116c75790505090506000835b8651811015611845576020821061178d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602a60248201527f50726f766964656420524c50206c6973742065786365656473206d6178206c6960448201527f7374206c656e6774682e000000000000000000000000000000000000000000006064820152608401610332565b6000806117ca6040518060400160405280858c600001516117ae9190612443565b8152602001858c602001516117c3919061236e565b9052611850565b5091509150604051806040016040528083836117e6919061236e565b8152602001848b602001516117fb919061236e565b8152508585815181106118105761181061245a565b602090810291909101015261182660018561236e565b9350611832818361236e565b61183c908461236e565b925050506116f4565b508152949350505050565b6000806000808460000151116118c2576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601860248201527f524c50206974656d2063616e6e6f74206265206e756c6c2e00000000000000006044820152606401610332565b6020840151805160001a607f81116118e7576000600160009450945094505050611c50565b60b7811161197d5760006118fc608083612443565b90508087600001511161196b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601960248201527f496e76616c696420524c502073686f727420737472696e672e000000000000006044820152606401610332565b60019550935060009250611c50915050565b60bf8111611aa057600061199260b783612443565b905080876000015111611a01576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f496e76616c696420524c50206c6f6e6720737472696e67206c656e6774682e006044820152606401610332565b600183015160208290036101000a9004611a1b818361236e565b885111611a84576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601860248201527f496e76616c696420524c50206c6f6e6720737472696e672e00000000000000006044820152606401610332565b611a8f82600161236e565b9650945060009350611c5092505050565b60f78111611b35576000611ab560c083612443565b905080876000015111611b24576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601760248201527f496e76616c696420524c502073686f7274206c6973742e0000000000000000006044820152606401610332565b600195509350849250611c50915050565b6000611b4260f783612443565b905080876000015111611bb1576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f496e76616c696420524c50206c6f6e67206c697374206c656e6774682e0000006044820152606401610332565b600183015160208290036101000a9004611bcb818361236e565b885111611c34576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601660248201527f496e76616c696420524c50206c6f6e67206c6973742e000000000000000000006044820152606401610332565b611c3f82600161236e565b9650945060019350611c5092505050565b9193909250565b606060008267ffffffffffffffff811115611c7457611c74611f86565b6040519080825280601f01601f191660200182016040528015611c9e576020820181803683370190505b509050805160001415611cb25790506109ed565b6000611cbe858761236e565b90506020820160005b611cd26020876125bb565b811015611d095782518252611ce860208461236e565b9250611cf560208361236e565b915080611d0181612489565b915050611cc7565b5060006001602087066020036101000a039050808251168119845116178252839450505050509392505050565b6060610c1182611f2e565b606081611d4f81601f61236e565b1015611db7576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152600e60248201527f736c6963655f6f766572666c6f770000000000000000000000000000000000006044820152606401610332565b82611dc2838261236e565b1015611e2a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152600e60248201527f736c6963655f6f766572666c6f770000000000000000000000000000000000006044820152606401610332565b611e34828461236e565b84511015611e9e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601160248201527f736c6963655f6f75744f66426f756e64730000000000000000000000000000006044820152606401610332565b606082158015611ebd5760405191506000825260208201604052611f25565b6040519150601f8416801560200281840101858101878315602002848b0101015b81831015611ef6578051835260209283019201611ede565b5050858452601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016604052505b50949350505050565b6060610c11826020015160008460000151611c57565b600060208284031215611f5657600080fd5b5035919050565b803573ffffffffffffffffffffffffffffffffffffffff81168114611f8157600080fd5b919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b604051601f82017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016810167ffffffffffffffff81118282101715611ffc57611ffc611f86565b604052919050565b600080600080600060a0868803121561201c57600080fd5b61202586611f5d565b94506020808701359450604087013567ffffffffffffffff808216821461204b57600080fd5b909450606088013590811515821461206257600080fd5b9093506080880135908082111561207857600080fd5b818901915089601f83011261208c57600080fd5b81358181111561209e5761209e611f86565b6120ce847fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0601f84011601611fb5565b91508082528a848285010111156120e457600080fd5b80848401858401376000848284010152508093505050509295509295909350565b60008083601f84011261211757600080fd5b50813567ffffffffffffffff81111561212f57600080fd5b60208301915083602082850101111561214757600080fd5b9250929050565b60008060008060008060008060008060006101808c8e03121561217057600080fd5b8b359a5061218060208d01611f5d565b995061218e60408d01611f5d565b985060608c0135975060808c0135965067ffffffffffffffff60a08d013511156121b757600080fd5b6121c78d60a08e01358e01612105565b909650945060c08c0135935060808c8e037fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff2001121561220557600080fd5b60e08c01925067ffffffffffffffff6101608d0135111561222557600080fd5b6122368d6101608e01358e01612105565b81935080925050509295989b509295989b9093969950565b6000815180845260005b8181101561227457602081850181015186830182015201612258565b81811115612286576000602083870101525b50601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169290920160200192915050565b85815284602082015267ffffffffffffffff84166040820152821515606082015260a060808201526000610afb60a083018461224e565b60006040828403121561230257600080fd5b6040516040810181811067ffffffffffffffff8211171561232557612325611f86565b604052825181526020928301519281019290925250919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600082198211156123815761238161233f565b500190565b60006080828403121561239857600080fd5b6040516080810181811067ffffffffffffffff821117156123bb576123bb611f86565b8060405250823581526020830135602082015260408301356040820152606083013560608201528091505092915050565b868152600073ffffffffffffffffffffffffffffffffffffffff808816602084015280871660408401525084606083015283608083015260c060a083015261243760c083018461224e565b98975050505050505050565b6000828210156124555761245561233f565b500390565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8214156124bb576124bb61233f565b5060010190565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b600060ff831680612504576125046124c2565b8060ff84160691505092915050565b600060ff821660ff84168082101561252d5761252d61233f565b90039392505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b6000817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff048311821515161561259d5761259d61233f565b500290565b6000602082840312156125b457600080fd5b5051919050565b6000826125ca576125ca6124c2565b50049056fea164736f6c634300080a000a",
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

// DepositTransaction is a paid mutator transaction binding the contract method 0xe9e05c42.
//
// Solidity: function depositTransaction(address _to, uint256 _value, uint64 _gasLimit, bool _isCreation, bytes _data) payable returns()
func (_OptimismPortal *OptimismPortalTransactor) DepositTransaction(opts *bind.TransactOpts, _to common.Address, _value *big.Int, _gasLimit uint64, _isCreation bool, _data []byte) (*types.Transaction, error) {
	return _OptimismPortal.contract.Transact(opts, "depositTransaction", _to, _value, _gasLimit, _isCreation, _data)
}

// DepositTransaction is a paid mutator transaction binding the contract method 0xe9e05c42.
//
// Solidity: function depositTransaction(address _to, uint256 _value, uint64 _gasLimit, bool _isCreation, bytes _data) payable returns()
func (_OptimismPortal *OptimismPortalSession) DepositTransaction(_to common.Address, _value *big.Int, _gasLimit uint64, _isCreation bool, _data []byte) (*types.Transaction, error) {
	return _OptimismPortal.Contract.DepositTransaction(&_OptimismPortal.TransactOpts, _to, _value, _gasLimit, _isCreation, _data)
}

// DepositTransaction is a paid mutator transaction binding the contract method 0xe9e05c42.
//
// Solidity: function depositTransaction(address _to, uint256 _value, uint64 _gasLimit, bool _isCreation, bytes _data) payable returns()
func (_OptimismPortal *OptimismPortalTransactorSession) DepositTransaction(_to common.Address, _value *big.Int, _gasLimit uint64, _isCreation bool, _data []byte) (*types.Transaction, error) {
	return _OptimismPortal.Contract.DepositTransaction(&_OptimismPortal.TransactOpts, _to, _value, _gasLimit, _isCreation, _data)
}

// FinalizeWithdrawalTransaction is a paid mutator transaction binding the contract method 0xeecf1c36.
//
// Solidity: function finalizeWithdrawalTransaction(uint256 _nonce, address _sender, address _target, uint256 _value, uint256 _gasLimit, bytes _data, uint256 _l2Timestamp, (bytes32,bytes32,bytes32,bytes32) _outputRootProof, bytes _withdrawalProof) payable returns()
func (_OptimismPortal *OptimismPortalTransactor) FinalizeWithdrawalTransaction(opts *bind.TransactOpts, _nonce *big.Int, _sender common.Address, _target common.Address, _value *big.Int, _gasLimit *big.Int, _data []byte, _l2Timestamp *big.Int, _outputRootProof WithdrawalVerifierOutputRootProof, _withdrawalProof []byte) (*types.Transaction, error) {
	return _OptimismPortal.contract.Transact(opts, "finalizeWithdrawalTransaction", _nonce, _sender, _target, _value, _gasLimit, _data, _l2Timestamp, _outputRootProof, _withdrawalProof)
}

// FinalizeWithdrawalTransaction is a paid mutator transaction binding the contract method 0xeecf1c36.
//
// Solidity: function finalizeWithdrawalTransaction(uint256 _nonce, address _sender, address _target, uint256 _value, uint256 _gasLimit, bytes _data, uint256 _l2Timestamp, (bytes32,bytes32,bytes32,bytes32) _outputRootProof, bytes _withdrawalProof) payable returns()
func (_OptimismPortal *OptimismPortalSession) FinalizeWithdrawalTransaction(_nonce *big.Int, _sender common.Address, _target common.Address, _value *big.Int, _gasLimit *big.Int, _data []byte, _l2Timestamp *big.Int, _outputRootProof WithdrawalVerifierOutputRootProof, _withdrawalProof []byte) (*types.Transaction, error) {
	return _OptimismPortal.Contract.FinalizeWithdrawalTransaction(&_OptimismPortal.TransactOpts, _nonce, _sender, _target, _value, _gasLimit, _data, _l2Timestamp, _outputRootProof, _withdrawalProof)
}

// FinalizeWithdrawalTransaction is a paid mutator transaction binding the contract method 0xeecf1c36.
//
// Solidity: function finalizeWithdrawalTransaction(uint256 _nonce, address _sender, address _target, uint256 _value, uint256 _gasLimit, bytes _data, uint256 _l2Timestamp, (bytes32,bytes32,bytes32,bytes32) _outputRootProof, bytes _withdrawalProof) payable returns()
func (_OptimismPortal *OptimismPortalTransactorSession) FinalizeWithdrawalTransaction(_nonce *big.Int, _sender common.Address, _target common.Address, _value *big.Int, _gasLimit *big.Int, _data []byte, _l2Timestamp *big.Int, _outputRootProof WithdrawalVerifierOutputRootProof, _withdrawalProof []byte) (*types.Transaction, error) {
	return _OptimismPortal.Contract.FinalizeWithdrawalTransaction(&_OptimismPortal.TransactOpts, _nonce, _sender, _target, _value, _gasLimit, _data, _l2Timestamp, _outputRootProof, _withdrawalProof)
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
	From       common.Address
	To         common.Address
	Mint       *big.Int
	Value      *big.Int
	GasLimit   uint64
	IsCreation bool
	Data       []byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterTransactionDeposited is a free log retrieval operation binding the contract event 0x78231ae6eb73366f912bb1d64351601fb76344c537bbab635ce14d0f376f0195.
//
// Solidity: event TransactionDeposited(address indexed from, address indexed to, uint256 mint, uint256 value, uint64 gasLimit, bool isCreation, bytes data)
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

// WatchTransactionDeposited is a free log subscription operation binding the contract event 0x78231ae6eb73366f912bb1d64351601fb76344c537bbab635ce14d0f376f0195.
//
// Solidity: event TransactionDeposited(address indexed from, address indexed to, uint256 mint, uint256 value, uint64 gasLimit, bool isCreation, bytes data)
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

// ParseTransactionDeposited is a log parse operation binding the contract event 0x78231ae6eb73366f912bb1d64351601fb76344c537bbab635ce14d0f376f0195.
//
// Solidity: event TransactionDeposited(address indexed from, address indexed to, uint256 mint, uint256 value, uint64 gasLimit, bool isCreation, bytes data)
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
	Arg0    [32]byte
	Success bool
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterWithdrawalFinalized is a free log retrieval operation binding the contract event 0xdb5c7652857aa163daadd670e116628fb42e869d8ac4251ef8971d9e5727df1b.
//
// Solidity: event WithdrawalFinalized(bytes32 indexed arg0, bool success)
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

// WatchWithdrawalFinalized is a free log subscription operation binding the contract event 0xdb5c7652857aa163daadd670e116628fb42e869d8ac4251ef8971d9e5727df1b.
//
// Solidity: event WithdrawalFinalized(bytes32 indexed arg0, bool success)
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

// ParseWithdrawalFinalized is a log parse operation binding the contract event 0xdb5c7652857aa163daadd670e116628fb42e869d8ac4251ef8971d9e5727df1b.
//
// Solidity: event WithdrawalFinalized(bytes32 indexed arg0, bool success)
func (_OptimismPortal *OptimismPortalFilterer) ParseWithdrawalFinalized(log types.Log) (*OptimismPortalWithdrawalFinalized, error) {
	event := new(OptimismPortalWithdrawalFinalized)
	if err := _OptimismPortal.contract.UnpackLog(event, "WithdrawalFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
