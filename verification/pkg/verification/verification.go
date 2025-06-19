package verification

import (
	"fmt"
	"log"

	types "base-tee/pkg/abstract_types"

	pb "github.com/google/go-tpm-tools/proto/attest"
	gotpm "github.com/google/go-tpm-tools/server"
	"google.golang.org/protobuf/proto"
)

func VerifyTEEComputationClaim(claim *types.TEEComputationClaim, verificationContext *types.TEEComputationClaimVerificationContext) error {
	log.Println("Verifying claim...")

	trustedAKs, err := getTrustedAKs()
	if err != nil {
		return fmt.Errorf("couldn't verify claim, failed to get trusted keys: %w", err)
	}

	attestation := &pb.Attestation{}
	err = proto.Unmarshal(verificationContext.Attestation, attestation)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	machineState, err := gotpm.VerifyAttestation(attestation, gotpm.VerifyOpts{
		Nonce:      claim.Nonce,
		TrustedAKs: trustedAKs,
	})
	if err != nil {
		return fmt.Errorf("failed to verify: %w", err)
	}

	if !machineState.GetSecureBoot().GetEnabled() {
		return fmt.Errorf("secure boot not enabled")
	}

	err = verifyQuotes(attestation.GetQuotes(), claim.DigestHistory)
	if err != nil {
		return err
	}

	log.Println("Verified!")
	return nil
}
