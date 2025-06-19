package generation

import (
	"base/pkg/abstract_types"
	basemodels "base/pkg/models"
	"context"
	"generation-view-fn-evm/pkg/models"
	"log"
	"math/big"
	"strings"
	"time"

	"base/pkg/ethrpc"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

// Generate generates a view function claim for a bridge transaction
//
// Parameters:
// - ethClient: The eth client instance
// - event: The event of the bridge transaction
func Generate(ethClient *ethclient.Client, event types.Log, sourceUslContractAddress common.Address, sourceUslContractABIJSON string) (*models.EVMViewFnClaim, *models.EVMViewFnClaimVerificationContext, error) {
	ctx := context.Background()

	blockNumberBigInt := new(big.Int).SetUint64(event.BlockNumber)

	chainId, err := ethClient.ChainID(ctx)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Get transaction information
	eventTx, _, err := ethClient.TransactionByHash(ctx, event.TxHash)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	eventTxFrom, err := types.Sender(types.LatestSignerForChainID(chainId), eventTx)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Construct the bridge contract abi
	sourceUslContractABI, err := abi.JSON(strings.NewReader(sourceUslContractABIJSON))
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Check if the event is from the USL contract
	if event.Address.Hex() != sourceUslContractAddress.Hex() {
		return nil, nil, errors.New("event address does not match source USL contract address")
	}

	// Unpack the event data
	eventData, err := sourceUslContractABI.Unpack("genStateQueryClaim", event.Data)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Get the relay message input and payload
	getRelayMessageInputBytes := eventData[4].([]byte)
	// eventMessagePayload := eventData[5].([]byte)

	// Get the relay message through the USL contract
	getRelayMessageOutput, err := ethClient.CallContract(ctx, ethereum.CallMsg{
		From: eventTxFrom,
		To:   &sourceUslContractAddress,
		Data: getRelayMessageInputBytes,
	}, blockNumberBigInt)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Unpack the relay message and check if the message payload matches, debug purpose
	// relayMessage, err := sourceUslContractABI.Unpack("relays", getRelayMessageOutput)
	// if err != nil {
	// 	return nil, errors.WithStack(err)
	// }
	// if hexutil.Encode(relayMessage[0].([]byte)) != hexutil.Encode(eventMessagePayload) {
	// 	return nil, errors.New("message payload does not match")
	// }

	// Get the block and block header
	block, err := ethClient.BlockByNumber(ctx, blockNumberBigInt)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Create access list
	accessList, _, err := ethrpc.CreateAccessList(ethClient.Client(), ctx, map[string]interface{}{
		// "from": eventTxFrom,
		"from":     "0x0000000000000000000000000000000000000000", // Use zero address to avoid the insufficient fund error, only for testnet
		"to":       sourceUslContractAddress.Hex(),
		"input":    hexutil.Encode(getRelayMessageInputBytes),
		"gasPrice": hexutil.EncodeBig(block.BaseFee()),
	}, block.Number())
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	accessList = append(accessList, basemodels.EVMAccessList{Address: eventTxFrom, StorageKeys: []string{}})

	// Get the account proofs with retry logic
	var account []abstract_types.Account
	waitTime := time.Duration(1) * time.Second
	maxRetries := 10
	for i := range maxRetries {
		account, err = ethrpc.GetProofsByAccessList(ethClient.Client(), ctx, accessList, blockNumberBigInt)
		if err == nil {
			break
		}
		log.Printf("GetProofsByAccessList failed, retrying... (attempt %d/%d)", i+1, maxRetries)
		time.Sleep(waitTime)
	}

	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	blockHeader := block.Header()

	return &models.EVMViewFnClaim{
			ClaimType: "EVMViewFn",
			Assumptions: &abstract_types.Header{
				ParentHash:  blockHeader.ParentHash,
				UncleHash:   blockHeader.UncleHash,
				Coinbase:    blockHeader.Coinbase,
				Root:        blockHeader.Root,
				TxHash:      blockHeader.TxHash,
				ReceiptHash: blockHeader.ReceiptHash,
				Bloom:       blockHeader.Bloom[:],
				Difficulty:  blockHeader.Difficulty,
				Number:      blockHeader.Number,
				GasLimit:    big.NewInt(int64(blockHeader.GasLimit)),
				GasUsed:     big.NewInt(int64(blockHeader.GasUsed)),
				Time:        big.NewInt(int64(blockHeader.Time)),
				Extra:       blockHeader.Extra,
				MixDigest:   blockHeader.MixDigest,
				Nonce:       blockHeader.Nonce,
			},
			Action: &abstract_types.EVMCall{
				From:  eventTxFrom,
				To:    sourceUslContractAddress,
				Input: getRelayMessageInputBytes,
			},
			Result: getRelayMessageOutput,
			Metadata: abstract_types.EVMMetadata{
				ChainId: chainId,
			},
		}, &models.EVMViewFnClaimVerificationContext{
			Accounts: account,
		}, nil
}
