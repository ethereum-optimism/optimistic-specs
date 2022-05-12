package surgery

import (
	"bytes"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"golang.org/x/crypto/sha3"
	"math/big"
	"path/filepath"
	"time"
)

var (
	OVMETHAddress = common.HexToAddress("0xDeadDeAddeAddEAddeadDEaDDEAdDeaDDeAD0000")
	emptyCodeHash = crypto.Keccak256(nil)
)

var zeroHash common.Hash

func Migrate(dataDir string, stateRoot common.Hash, genesis *core.Genesis, outDir string, expectedSupply *big.Int) error {
	// Instantiate the v0 LevelDB database.
	inDB, err := rawdb.NewLevelDBDatabase(
		filepath.Join(dataDir, "geth", "chaindata"),
		0,
		0,
		"",
		true,
	)
	if err != nil {
		log.Crit("error opening raw DB", "err", err)
	}

	// Instantiate the v0 state database.
	inUnderlyingDB := state.NewDatabase(inDB)
	inStateDB, err := state.New(stateRoot, inUnderlyingDB, nil)
	if err != nil {
		log.Crit("error opening state db", "err", err)
	}

	// Create and instantiate the Bedrock LevelDB database.
	outDB, err := rawdb.NewLevelDBDatabase(outDir, 0, 0, "", false)
	if err != nil {
		return err
	}

	if stateRoot == zeroHash {
		stateRoot = getStateRoot(inDB)
	}

	// Create and instantiate the Bedrock state database.
	outStateDB, err := state.New(common.Hash{}, state.NewDatabase(outDB), nil)
	if err != nil {
		log.Crit("error opening output state DB", "err", err)
	}

	// Dump the state and get the state root.
	log.Info("dumping state")
	newRoot := dumpState(inUnderlyingDB, inStateDB, outStateDB, stateRoot, genesis, expectedSupply)

	// Now that the state is dumped, insert the genesis block. We pass in a nil
	// database here because we don't want to update the state again with the
	// pre-allocs.
	//
	// Unlike regular Geth (which panics if you try to import a genesis state with a nonzero
	// block number), the block number can be anything.
	block := genesis.ToBlock(nil)

	// Geth block headers are immutable, so swap the root and make a new block with the
	// updated root.
	header := block.Header()
	header.Root = newRoot
	block = types.NewBlock(header, nil, nil, nil, trie.NewStackTrie(nil))
	blob, err := json.Marshal(genesis)
	if err != nil {
		log.Crit("error marshaling genesis state", "err", err)
	}

	// Write the genesis state to the database. This is taken verbatim from Geth's
	// core.Genesis struct.
	rawdb.WriteGenesisState(outDB, block.Hash(), blob)
	rawdb.WriteTd(outDB, block.Hash(), block.NumberU64(), block.Difficulty())
	rawdb.WriteBlock(outDB, block)
	rawdb.WriteReceipts(outDB, block.Hash(), block.NumberU64(), nil)
	rawdb.WriteCanonicalHash(outDB, block.Hash(), block.NumberU64())
	rawdb.WriteHeadBlockHash(outDB, block.Hash())
	rawdb.WriteHeadFastBlockHash(outDB, block.Hash())
	rawdb.WriteHeadHeaderHash(outDB, block.Hash())
	rawdb.WriteChainConfig(outDB, block.Hash(), genesis.Config)
	return nil
}

func dumpState(inDB state.Database, inStateDB *state.StateDB, outStateDB *state.StateDB, root common.Hash, genesis *core.Genesis, expectedSupply *big.Int) common.Hash {
	// Open a trie based on the currently-configured state root.
	tr, err := inDB.OpenTrie(root)
	if err != nil {
		panic(err)
	}

	// Iterate over the Genesis allocation accounts. These will override
	// and accounts found in the state.
	log.Info("importing allocated accounts")
	for addr, account := range genesis.Alloc {
		outStateDB.AddBalance(addr, account.Balance)
		outStateDB.SetCode(addr, account.Code)
		outStateDB.SetNonce(addr, account.Nonce)
		for key, value := range account.Storage {
			outStateDB.SetState(addr, key, value)
		}
		log.Info("allocated account", "addr", addr)
	}

	var (
		accounts     uint64
		lastAccounts uint64
		start        = time.Now()
		logged       = time.Now()
	)
	log.Info("Trie dumping started", "root", tr.Hash())

	// Keep track of total OVM ETH migrated to verify that the script worked.
	totalOVM := new(big.Int)

	// Iterate over each account in the state.
	it := trie.NewIterator(tr.NodeIterator(nil))
	for it.Next() {
		// It's up to use to decode trie data.
		var data types.StateAccount
		if err := rlp.DecodeBytes(it.Value, &data); err != nil {
			panic(err)
		}

		addrBytes := tr.GetKey(it.Key)
		addr := common.BytesToAddress(addrBytes)

		// Skip genesis addresses.
		if _, ok := genesis.Alloc[addr]; ok {
			log.Info("skipping preallocated account", "addr", addr)
			continue
		}

		addrHash := crypto.Keccak256Hash(addr[:])
		code := getCode(addrHash, data, inDB)

		// Get the OVM ETH balance based on the address's storage key.
		ovmBalance := getOVMETHBalance(inStateDB, addr)

		// No accounts should have a balance in state. If they do, bail.
		if data.Balance.Sign() > 0 {
			log.Crit("account has non-zero OVM eth balance", "addr", addr)
		}

		// Actually perform the migration by setting the appropriate values in state.
		outStateDB.AddBalance(addr, ovmBalance)
		outStateDB.SetCode(addr, code)
		outStateDB.SetNonce(addr, data.Nonce)

		// Bump the total OVM balance.
		totalOVM = totalOVM.Add(totalOVM, ovmBalance)

		// Grab the storage trie.
		storageTrie, err := inDB.OpenStorageTrie(addrHash, data.Root)
		if err != nil {
			panic(err)
		}
		storageIt := trie.NewIterator(storageTrie.NodeIterator(nil))
		var storageSlots uint64
		storageLogged := time.Now()
		for storageIt.Next() {
			storageSlots++
			_, content, _, err := rlp.Split(storageIt.Value)
			if err != nil {
				panic(err)
			}

			// Update each storage slot for this account in state.
			outStateDB.SetState(
				addr,
				common.BytesToHash(storageTrie.GetKey(storageIt.Key)),
				common.BytesToHash(content),
			)

			// Log status every 8 seconds.
			if time.Since(storageLogged) > 8*time.Second {
				since := time.Since(start)
				log.Info("Storage dumping in progress", "addr", addr, "storage_slots", storageSlots,
					"elapsed", common.PrettyDuration(since))
				storageLogged = time.Now()
			}
		}

		accounts++

		// Log status every 8 seconds.
		sinceLogged := time.Since(logged)
		if sinceLogged > 8*time.Second {
			sinceStart := time.Since(start)
			rate := float64(accounts-lastAccounts) / float64(sinceLogged/time.Second)

			log.Info("Trie dumping in progress", "at", it.Key, "accounts", accounts,
				"elapsed", common.PrettyDuration(sinceStart), "accs_per_s", rate)
			logged = time.Now()
			lastAccounts = accounts
		}
	}
	log.Info("Trie dumping complete", "accounts", accounts,
		"elapsed", common.PrettyDuration(time.Since(start)), "total_ovm_eth", totalOVM)

	if totalOVM.Cmp(expectedSupply) != 0 {
		log.Crit("total eth supply doesn't match!", "got", totalOVM, "expected", expectedSupply)
	}

	// Commit the state DB. First call flushes changes in memory. Second persists them to disk.
	log.Info("committing state DB")
	newRoot, err := outStateDB.Commit(false)
	if err != nil {
		log.Crit("error writing output state DB", "err", err)
	}
	log.Info("committed state DB", "root", newRoot)
	log.Info("committing trie DB")
	if err := outStateDB.Database().TrieDB().Commit(newRoot, true, nil); err != nil {
		log.Crit("error writing output trie DB", "err", err)
	}
	log.Info("committed trie DB")

	return newRoot
}

// getCode returns a contract's code. Taken verbatim from Geth.
func getCode(addrHash common.Hash, data types.StateAccount, db state.Database) []byte {
	if bytes.Equal(data.CodeHash, emptyCodeHash) {
		return nil
	}

	code, err := db.ContractCode(
		addrHash,
		common.BytesToHash(data.CodeHash),
	)
	if err != nil {
		panic(err)
	}
	return code
}

// getStateRoot returns the head block's state root.
func getStateRoot(db ethdb.Reader) common.Hash {
	block := rawdb.ReadHeadBlock(db)
	return block.Root()
}

// getOVMETHBalance gets a user's OVM ETH balance from state by querying the
// appropriate storage slot directly.
func getOVMETHBalance(inStateDB *state.StateDB, addr common.Address) *big.Int {
	position := common.Big0
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(common.LeftPadBytes(addr.Bytes(), 32))
	hasher.Write(common.LeftPadBytes(position.Bytes(), 32))
	digest := hasher.Sum(nil)
	balKey := common.BytesToHash(digest)
	return inStateDB.GetState(OVMETHAddress, balKey).Big()
}
