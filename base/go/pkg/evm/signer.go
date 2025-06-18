package evm

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pkg/errors"
)

type SignedComponents struct {
	Hash string `json:"hash"`
	R    string `json:"r"`
	S    string `json:"s"`
	V    uint8  `json:"v"`
}

// EIP191Hash calculates the EIP-191 hash for a given bytes.
// This is equivalent to what `eth_sign` expects.
func EIP191Hash(data []byte) common.Hash {
	prefix := fmt.Appendf(nil, "\x19Ethereum Signed Message:\n%d", len(data))
	dataToHash := append(prefix, data...)
	return crypto.Keccak256Hash(dataToHash)
}

// PrivateKeyFromHex converts a hex string to an ECDSA private key.
func PrivateKeyFromHex(privateKeyHex string) (*ecdsa.PrivateKey, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return privateKey, nil
}

// AddressFromPrivateKey converts a private key to the address of the public key.
func AddressFromPrivateKey(privateKeyHex string) (*common.Address, error) {
	privateKey, err := PrivateKeyFromHex(privateKeyHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	return &address, nil
}

func SignMessage(privateKeyHex string, message any) (*SignedComponents, error) {
	privateKey, err := PrivateKeyFromHex(privateKeyHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Encode the message to bytes
	messageBytes, err := rlp.EncodeToBytes(message)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Calculate the EIP-191 hash
	messageHash := EIP191Hash(messageBytes)

	// Sign the message
	signature, err := crypto.Sign(messageHash.Bytes(), privateKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Extract the r, s, v
	r := signature[:32]
	s := signature[32:64]
	v := signature[64]

	// Adjust V for eth_sign compatibility (27 or 28)
	// https://github.com/ethereum/go-ethereum/blob/master/crypto/signature_nocgo.go#L92-L93 Ecrecover uses `v-27`
	// so the V value in the signature byte array is 0 or 1.
	// To match common signature representations (like in ethers.js, web3.js, or what many wallets expect for eth_sign),
	// V is typically 27 or 28.
	// So, if v_raw is 0, V becomes 27. If v_raw is 1, V becomes 28.
	adjustedV := v + 27

	return &SignedComponents{
		Hash: messageHash.Hex(),
		R:    hexutil.Encode(r),
		S:    hexutil.Encode(s),
		V:    adjustedV,
	}, nil
}
