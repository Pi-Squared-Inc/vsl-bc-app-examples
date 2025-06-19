package verification

import (
	"generation-block-processing-evm/pkg/models"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pkg/errors"
)

// Verify verifies a block processing claim
//
// Parameters:
// - claim: The block processing claim to verify
// - verificationContext: The verification context for the claim
func Verify(claim *models.EVMBlockProcessingClaim, verificationContext *models.EVMBlockProcessingClaimVerificationContext) error {
	// Deserialize the witness from bytes with RLP
	var witness *stateless.Witness
	err := rlp.DecodeBytes(verificationContext.Witness, &witness)
	if err != nil {
		return errors.WithStack(err)
	}

	// Deserialize the block from bytes with RLP
	var block *types.Block
	err = rlp.DecodeBytes(claim.Result, &block)
	if err != nil {
		return errors.WithStack(err)
	}

	// Check if the previous state root matches the claim's assumptions
	if witness.Root() != claim.Assumptions.Root {
		return errors.New("previous state root mismatch")
	}

	// Execute the block and get the post-state root and receipt root
	postStateRoot, postReceiptRoot, err := core.ExecuteStateless(params.MainnetChainConfig, vm.Config{}, block, witness)
	if err != nil {
		return errors.WithStack(err)
	}

	// Check if the post-state root matches the block's header root
	if postStateRoot != block.Header().Root {
		return errors.New("post-state root mismatch")
	}

	// Check if the post-state receipts root matches the block's header receipts root
	if postReceiptRoot != block.Header().ReceiptHash {
		return errors.New("post-state receipts root mismatch")
	}

	return nil
}
