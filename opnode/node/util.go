package node

import (
	"bytes"

	"github.com/ethereum-optimism/optimistic-specs/opnode/l2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/trie"
)

func ComputeL2OutputRoot(l2OutputRootVersion l2.Bytes32, blockHash common.Hash, blockRoot common.Hash, storageRoot common.Hash) []byte {
	var buf bytes.Buffer
	buf.Write(l2OutputRootVersion[:])
	buf.Write(blockRoot.Bytes())
	buf.Write(storageRoot[:])
	buf.Write(blockHash.Bytes())
	return crypto.Keccak256(buf.Bytes())
}

func VerifyAccountProof(stateRoot common.Hash, accountRoot common.Hash, proof []string) error {
	p := newProofDB(proof)
	_, err := trie.VerifyProof(stateRoot, accountRoot[:], p)
	return err
}

type proofDB struct {
	m map[string][]byte
}

func newProofDB(proof []string) *proofDB {
	db := &proofDB{make(map[string][]byte)}

	for _, p := range proof {
		buf := common.FromHex(p)
		hash := crypto.Keccak256(buf)
		db.m[string(hash)] = buf
	}
	return db
}

func (d *proofDB) Get(key []byte) ([]byte, error) {
	return d.m[string(key)], nil
}

func (d *proofDB) Has(key []byte) (bool, error) {
	val, _ := d.Get(key)
	return val != nil, nil
}
