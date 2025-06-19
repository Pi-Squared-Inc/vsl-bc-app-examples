package utils

import (
	"crypto/ecdsa"
	"log"
	"testing" // Import the testing package

	"github.com/ethereum/go-ethereum/crypto"
)

func TestSignVerify_Ok(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("cannot cast public key to ECDSA public key")
	}
	signerAddress := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	messageData := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce ac malesuada dolor. Nam fermentum in.")
	signedMessageHash, signatureR, signatureS, signatureV, err := Sign(messageData, privateKey)
	if err != nil {
		t.Fatal("cannot sign")
	}

	err = VerifySign(messageData, signerAddress, signedMessageHash, signatureR, signatureS, signatureV)
	if err != nil {
		t.Errorf("fail: %s", err.Error())
		t.Logf("signer address: %s", signerAddress)
		t.Logf("original message: \"%s\"", string(messageData))
		t.Logf("signature hash: %s", signedMessageHash)
		t.Logf("signature r: %s", signatureR)
		t.Logf("signature s: %s", signatureS)
		t.Logf("signature v: %s", signatureV)
	} else {
		t.Logf("pass")
	}
}

func TestSignVerify_Nok(t *testing.T) {
	// Generate a private key and sign a message to get valid signature components
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("failed to generate private key: %v", err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
	signerAddress := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	messageData := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Ut aliquet viverra finibus. Aliquam lectus.")
	signedMessageHash, signatureR, signatureS, signatureV, err := Sign(messageData, privateKey)
	if err != nil {
		t.Fatal("cannot sign")
	}

	tests := []struct {
		name     string
		msgBytes []byte
		addr     string
		hash     string
		r        string
		s        string
		v        string
	}{
		{
			name:     "Mismatched mesage",
			msgBytes: []byte("Gorem ipsum dolor sit amet, consectetur adipiscing elit. Ut aliquet viverra finibus. Aliquam lectus."),
			addr:     signerAddress,
			hash:     signedMessageHash,
			r:        signatureR,
			s:        signatureS,
			v:        signatureV,
		},
		{
			name:     "Mismatched signer address",
			msgBytes: messageData,
			addr:     "0x0000000000000000000000000000000000000001",
			hash:     signedMessageHash,
			r:        signatureR,
			s:        signatureS,
		},
		{
			name:     "Corrupted hash",
			msgBytes: messageData,
			addr:     signerAddress,
			hash:     "abcd",
			r:        signatureR,
			s:        signatureS,
			v:        signatureV,
		},
		{
			name:     "Corrputed R",
			msgBytes: messageData,
			addr:     signerAddress,
			hash:     signedMessageHash,
			r:        signatureR + "00",
			s:        signatureS,
			v:        signatureV,
		},
		{
			name:     "Corrputed S",
			msgBytes: messageData,
			addr:     signerAddress,
			hash:     signedMessageHash,
			r:        signatureR,
			s:        signatureS + "00",
			v:        signatureV,
		},
		{
			name:     "Corrputed V",
			msgBytes: messageData,
			addr:     signerAddress,
			hash:     signedMessageHash,
			r:        signatureR,
			s:        signatureS,
			v:        signatureV + "00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifySign(tt.msgBytes, tt.addr, tt.hash, tt.r, tt.s, tt.v)
			if err == nil {
				t.Errorf("fail on %s.", tt.name)
			}
		})
	}
}
