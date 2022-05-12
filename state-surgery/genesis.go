package surgery

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/core"
	"io"
	"os"
)

func ReadGenesisFromFile(path string) (*core.Genesis, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadGenesis(f)
}

func ReadGenesis(r io.Reader) (*core.Genesis, error) {
	genesis := new(core.Genesis)
	if err := json.NewDecoder(r).Decode(genesis); err != nil {
		return nil, err
	}
	return genesis, nil
}
