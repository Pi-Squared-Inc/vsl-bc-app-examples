package generation

import (
	"context"
	"generation-block-processing-evm/pkg/models"
	"math/big"

	"base/pkg/abstract_types"
	basemodels "base/pkg/models"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pkg/errors"
)

// Generate generates a block claim with specified block number
//
// Parameters:
// - ethClient: The eth client instance
// - blockNumber: The block number
func Generate(ethClient *ethclient.Client, blockNumber *big.Int) (*models.EVMBlockProcessingClaim, *models.EVMBlockProcessingClaimVerificationContext, error) {
	ctx := context.Background()
	chainId, err := ethClient.ChainID(ctx)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Get previous and current block headers
	block, err := ethClient.BlockByNumber(ctx, blockNumber)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Reth witness (not used)
	// previousBlockHeader, err := app.EthRPCClient.HeaderByHash(ctx, block.ParentHash())
	// if err != nil {
	// 	return nil, nil, errors.WithStack(err)
	// }
	// var rethWitness *basemodels.RethWitness
	// err = app.EthRPCClient.Client().CallContext(ctx, &rethWitness, "debug_executionWitnessByBlockHash", block.Hash(), false)
	// if err != nil {
	// 	return nil, nil, errors.WithStack(err)
	// }
	// witness := rethWitness.ToStatelessWitness(previousBlockHeader, block.Header())

	// Geth witness
	var gethWitness *basemodels.GethWitness
	err = ethClient.Client().CallContext(ctx, &gethWitness, "debug_executionWitness", block.Hash())
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	witness := gethWitness.ToStatelessWitness(block.Header())

	// Serialize the witness into bytes with RLP
	witnessBytes, err := rlp.EncodeToBytes(witness)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Serialize the block into bytes with RLP
	blockBytes, err := rlp.EncodeToBytes(block)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return &models.EVMBlockProcessingClaim{
			ClaimType:   "MirroringGeth",
			Assumptions: gethWitness.Headers[0],
			Metadata: abstract_types.EVMMetadata{
				ChainId: chainId,
			},
			Result: blockBytes,
		}, &models.EVMBlockProcessingClaimVerificationContext{
			Witness: witnessBytes,
		}, nil
}
