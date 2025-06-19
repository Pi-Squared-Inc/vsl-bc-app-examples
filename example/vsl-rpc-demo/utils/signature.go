package utils

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func Sign(dataBytes []byte, key *ecdsa.PrivateKey) (string, string, string, string, error) {
	hashData := accounts.TextHash(dataBytes)
	signatureBytes, err := crypto.Sign(hashData, key)
	if err != nil {
		return "", "", "", "", err
	}
	r := signatureBytes[:32]
	s := signatureBytes[32:64]
	v := signatureBytes[64:]
	return hexutil.Encode(hashData[:]), hexutil.Encode(r), hexutil.Encode(s), hexutil.Encode(v), nil
}

func VerifySign(dataBytes []byte, addr string, hash string, rStr string, sStr string, vStr string) error {
	// 1. Parse signature components from hexadecimal strings
	r, err := hexutil.Decode(rStr)
	if err != nil || len(r) != 32 {
		return fmt.Errorf("could not decode r")
	}
	s, err := hexutil.Decode(sStr)
	if err != nil || len(s) != 32 {
		return fmt.Errorf("could not decode s")
	}
	v, err := hexutil.Decode(vStr)
	if err != nil || len(v) != 1 {
		return fmt.Errorf("could not decode v")
	}

	// 2. Compare message hash with provided hash
	messageHash := accounts.TextHash(dataBytes)
	providedHashBytes, err := hexutil.Decode(hash)
	if err != nil {
		return fmt.Errorf("could not decode hash")
	}
	if !bytes.Equal(messageHash, providedHashBytes) {
		return fmt.Errorf("hash mismatch")
	}

	// 3. Construct the 65-byte raw signature array (R || S || V)
	signature := make([]byte, 65)
	copy(signature[0:32], r)
	copy(signature[32:64], s)
	vByte := v[0]
	if vByte == 0 || vByte == 1 {
		signature[64] = vByte
	} else if vByte == 27 || vByte == 28 {
		signature[64] = vByte - 27
	} else {
		return fmt.Errorf("invalid v value")
	}

	// 4. Recover the public key & address from the message hash and signature
	recoveredPubKeyBytes, err := crypto.Ecrecover(messageHash, signature)
	if err != nil {
		return fmt.Errorf("could not recover public key bytes from signature")
	}
	publicKey, err := crypto.UnmarshalPubkey(recoveredPubKeyBytes)
	if err != nil {
		return fmt.Errorf("could not unmarshal public key from bytes")
	}
	recoveredAddress := crypto.PubkeyToAddress(*publicKey)

	// 5. Compare the recovered address with the expected address
	expectedAddress := common.HexToAddress(addr)

	if recoveredAddress != expectedAddress {
		return fmt.Errorf("address mismatch")
	}
	return nil
}
