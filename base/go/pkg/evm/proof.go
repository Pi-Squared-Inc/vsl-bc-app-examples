package evm

import (
	"base/pkg/abstract_types"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/pkg/errors"
)

func GenerateProofDB(proofs [][]byte) (ethdb.KeyValueReader, error) {
	db := memorydb.New()
	for _, proof := range proofs {
		key := crypto.Keccak256(proof)
		err := db.Put(key, proof)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return db, nil
}

// DecodeProofs decodes the proofs with the state root and address
func DecodeProofs(stateRoot common.Hash, address common.Address, proof *abstract_types.AccountProof) ([]byte, error) {
	db, err := GenerateProofDB(proof.AccountProof)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	secureKey := crypto.Keccak256(address.Bytes())
	return trie.VerifyProof(stateRoot, secureKey, db)
}

// VerifyCode verifies the code with the code hash
func VerifyCode(code []byte, codeHash common.Hash) error {
	if codeHash == crypto.Keccak256Hash(code) {
		return nil
	}
	return errors.WithStack(errors.New("code verification failed"))
}

// VerifyProof verifies the account proofs with the state root and address
func VerifyProof(stateRoot common.Hash, address common.Address, proof *abstract_types.Account) error {
	accountRLP, err := DecodeProofs(stateRoot, address, &proof.Proof)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(accountRLP) == 0 &&
		proof.Proof.Balance.Cmp(big.NewInt(0)) == 0 &&
		proof.Proof.Nonce.Cmp(big.NewInt(0)) == 0 {
		return nil
	}
	var accountInformation [][]byte
	err = rlp.DecodeBytes(accountRLP, &accountInformation)
	if err != nil {
		return errors.WithStack(err)
	}

	// Verify code with code hash that decoded from proof
	err = VerifyCode(proof.Code, common.BytesToHash(accountInformation[3]))
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
