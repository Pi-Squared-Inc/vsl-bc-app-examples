package verification

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	tpmpb "github.com/google/go-tpm-tools/proto/tpm"
	"github.com/google/go-tpm/legacy/tpm2"
)

const (
	PCRIndex uint32 = 23
)

func verifyQuotes(quotes []*tpmpb.Quote, history map[string][]string) error {
	for _, quote := range quotes {
		pcrs := quote.GetPcrs()
		goldenPcr := getReferenceHash(history, tpm2.Algorithm(pcrs.GetHash()))
		if goldenPcr == nil {
			continue
		}
		if !bytes.Equal(goldenPcr, pcrs.Pcrs[PCRIndex]) {
			return fmt.Errorf("integrity checking of PCR quotes for program and result failed")
		}
	}
	return nil
}

func getReferenceHash(history map[string][]string, algo tpm2.Algorithm) []byte {
	var refHash []byte
	switch algo {
	case tpm2.AlgSHA1:
		refHash = make([]byte, 20)[:]
		for _, digest := range history["sha1"] {
			bDigest, err := hex.DecodeString(digest)
			if err != nil {
				return nil
			}
			extend := sha1.Sum(append(refHash, bDigest...))
			refHash = extend[:]
		}
	case tpm2.AlgSHA256:
		refHash = make([]byte, 32)[:]
		for _, digest := range history["sha256"] {
			bDigest, err := hex.DecodeString(digest)
			if err != nil {
				return nil
			}
			extend := sha256.Sum256(append(refHash, bDigest...))
			refHash = extend[:]
		}
	case tpm2.AlgSHA384:
		refHash = make([]byte, 48)[:]
		for _, digest := range history["sha384"] {
			bDigest, err := hex.DecodeString(digest)
			if err != nil {
				return nil
			}
			extend := sha512.Sum384(append(refHash, bDigest...))
			refHash = extend[:]
		}
	}
	return refHash
}
