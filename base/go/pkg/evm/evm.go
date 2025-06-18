package evm

import (
	"base/pkg/abstract_types"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/pkg/errors"
)

func CreateConfig(chainId *big.Int) *params.ChainConfig {
	shanghaiTime := uint64(0)
	cancunTime := uint64(0)
	return &params.ChainConfig{
		ChainID:             chainId,
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        big.NewInt(0),
		DAOForkSupport:      true,
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
		ArrowGlacierBlock:   big.NewInt(0),
		GrayGlacierBlock:    big.NewInt(0),
		ShanghaiTime:        &shanghaiTime,
		CancunTime:          &cancunTime,
		BlobScheduleConfig:  params.DefaultBlobSchedule,
	}
}

// CreateEVM creates and initializes an EVM instance with the given parameters
//
//   - chainId: Chain ID
//   - stateRoot: State root hash of the EVM
//   - blockHeader: Initial block header
//   - accountProofs: Account proofs, including account proof, storage proof, and code
//   - getBlockHash: Function to get the block hash
func CreateEVM(chainId *big.Int, stateRoot common.Hash, blockHeader *types.Header, accountProofs []abstract_types.Account, getBlockHash func(u uint64) common.Hash) (*vm.EVM, *state.StateDB, error) {
	// Related variables
	hash := make([]byte, 32)
	hasher := crypto.NewKeccakState()
	memdb := rawdb.NewMemoryDatabase()

	// Header is contract with the pre-state block header state root and post-state block header other fields
	header := *blockHeader
	header.Root = stateRoot

	chainConfig := CreateConfig(chainId)

	// Write block header
	rawdb.WriteHeader(memdb, &header)
	// Set account and storage to Memory Database, this part is take https://github.com/ethereum/go-ethereum/blob/master/core/stateless/database.go#L29 as implementation reference
	for _, v := range accountProofs {
		// Verify proof
		err := VerifyProof(stateRoot, v.Proof.Addr, &v)
		if err != nil {
			return nil, nil, errors.WithStack(err)
		}

		// Set account
		for _, proof := range v.Proof.AccountProof {
			hasher.Reset()
			hasher.Write(proof)
			hasher.Read(hash)
			rawdb.WriteLegacyTrieNode(memdb, common.BytesToHash(hash), proof)
		}

		// Set code
		if len(v.Code) > 0 {
			hasher.Reset()
			hasher.Write(v.Code)
			hasher.Read(hash)
			rawdb.WriteCode(memdb, common.BytesToHash(hash), v.Code)
		}

		// Set storage proof
		for _, storageProof := range v.Proof.StorageProof {
			for _, proof := range storageProof.Proof {
				hasher.Reset()
				hasher.Write(proof)
				hasher.Read(hash)
				rawdb.WriteLegacyTrieNode(memdb, common.BytesToHash(hash), proof)
			}
		}
	}

	// Initialize StateDB and EVM
	stateDB, err := state.New(stateRoot, state.NewDatabase(triedb.NewDatabase(memdb, triedb.HashDefaults), nil))
	if err != nil {
		fmt.Printf("error initializing stateDB: %+v\n", err)
		return nil, nil, errors.WithStack(err)
	}
	// stateDBWrapper := NewStateDBWrapper(stateDB)
	evm := vm.NewEVM(vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Coinbase:    header.Coinbase,
		GasLimit:    header.GasLimit,
		BlockNumber: header.Number,
		Time:        header.Time,
		Difficulty:  header.Difficulty,
		BaseFee:     header.BaseFee,
		GetHash:     getBlockHash,
		Random:      &header.MixDigest,
		BlobBaseFee: big.NewInt(0),
	}, stateDB, chainConfig, vm.Config{
		StatelessSelfValidation: true,
		NoBaseFee:               false,
		ExtraEips:               []int{}, // Added more EIPs if needed
	})
	return evm, stateDB, nil
}
