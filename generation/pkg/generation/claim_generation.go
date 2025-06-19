package generation

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log"

	types "base-tee/pkg/abstract_types"
)

const (
	PCRIndex uint32 = 23

	ResnetModelSHA1   string = "d24b18dbfffba75079c1f5dba667c97d6193a598"
	ResnetModelSHA256 string = "5bee01fc5ba0d1225f7860cb856c630c38d31f803d7e2fb1d82c4f9de078276c"
	ResnetModelSHA384 string = "ce8a2814c0707e0a0594a82ed3243f40a15c65b8ed8215e0ee08ff89abbffb1456730ea8d804c7d4b7938a88756df643"

	BlockProcessingKRethSHA1   string = "db66208f41fd32f7f2cec2c245e17ac2a64eaad5"
	BlockProcessingKRethSHA256 string = "bb28c5940f8cfa05e884103f10302ec58d23fc673ad92f1cd340bd85f2ab7b0c"
	BlockProcessingKRethSHA384 string = "faee8766664dc039607e5cd59a65e77e55b0646da2cbadb12a35e594ae5d2cfe0f3b99b2e8439e174de255251f9ad933"

	LLaMaModelSHA1   string = "5229bd4a103bf937e62c811218408ffa8aa62bad"
	LLaMaModelSHA256 string = "08a5566d61d7cb6b420c3e4387a39e0078e1f2fe5f055f3a03887385304d4bfa"
	LLaMaModelSHA384 string = "9e946a7e2d13afdd1bd08d6c5b5eee8840f10808e9bc5fe0da9eed8277e85644eb25fea58923e128c259d1c174968d6a"
)

func GenerateTEEComputationClaim(computation types.Computation, input []string, result string, byteAtt []byte, nonce []byte) (*types.TEEComputationClaim, *types.TEEComputationClaimVerificationContext, error) {
	// Claim generation is currently just a glorified constructor call.
	log.Println("Generating claim...")
	digest_history, err := computeHistory(computation, append(input, result))
	if err != nil {
		return nil, nil, fmt.Errorf("claim generation failed due to computeHistory: %w", err)
	}
	return &types.TEEComputationClaim{
			ClaimType:     types.ClaimTypeTEEComputation,
			Computation:   computation,
			Input:         input,
			Result:        result,
			DigestHistory: digest_history,
			Nonce:         nonce,
		}, &types.TEEComputationClaimVerificationContext{
			Attestation: byteAtt,
		}, nil
}

func computeHistory(computation types.Computation, events []string) (map[string][]string, error) {
	history := make(map[string][]string)
	history["sha1"] = make([]string, 0)
	history["sha256"] = make([]string, 0)
	history["sha384"] = make([]string, 0)

	switch computation {
	case types.InferImageClass:
		if len(events) != 2 {
			return nil, fmt.Errorf("unexpected event history for image classfication: %s", events)
		}
		history["sha1"] = append(history["sha1"], ResnetModelSHA1)
		history["sha256"] = append(history["sha256"], ResnetModelSHA256)
		history["sha384"] = append(history["sha384"], ResnetModelSHA384)
	case types.InferTextGen:
		if len(events) != 2 {
			return nil, fmt.Errorf("unexpected event history for image classfication: %s", events)
		}
		history["sha1"] = append(history["sha1"], LLaMaModelSHA1)
		history["sha256"] = append(history["sha256"], LLaMaModelSHA256)
		history["sha384"] = append(history["sha384"], LLaMaModelSHA384)
	case types.BlockProcessingKReth:
		if len(events) != 3 {
			return nil, fmt.Errorf("unexpected event history for block processing using kreth: %s", events)
		}
		// Removing the 1-th event from the `events` list
		// because it is the context which is not needed for the hash
		events = append(events[:1], events[2:]...)

		history["sha1"] = append(history["sha1"], BlockProcessingKRethSHA1)
		history["sha256"] = append(history["sha256"], BlockProcessingKRethSHA256)
		history["sha384"] = append(history["sha384"], BlockProcessingKRethSHA384)
	default:
	}

	for _, ev := range events {
		sha1 := sha1.Sum([]byte(ev))
		sha1sum := hex.EncodeToString(sha1[:])
		sha256 := sha256.Sum256([]byte(ev))
		sha256sum := hex.EncodeToString(sha256[:])
		sha384 := sha512.Sum384([]byte(ev))
		sha384sum := hex.EncodeToString(sha384[:])

		history["sha1"] = append(history["sha1"], sha1sum)
		history["sha256"] = append(history["sha256"], sha256sum)
		history["sha384"] = append(history["sha384"], sha384sum)
	}

	return history, nil
}
