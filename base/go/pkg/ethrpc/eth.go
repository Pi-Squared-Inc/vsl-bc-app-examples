package ethrpc

import (
	"base/pkg/abstract_types"
	"base/pkg/models"
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

func GetTransactionByHash(client *rpc.Client, ctx context.Context, txHash string) (*models.EVMTransaction, error) {
	resultJSON, err := CallContextWithJSONResponse(client, ctx, "eth_getTransactionByHash", txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction by hash: %v", err)
	}

	var tx models.EVMTransaction
	err = json.Unmarshal([]byte(*resultJSON), &tx)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %v", err)
	}
	return &tx, nil
}

func GetTransactionReceiptByHash(client *rpc.Client, ctx context.Context, txHash string) (*models.EVMTransactionReceipt, error) {
	resultJSON, err := CallContextWithJSONResponse(client, ctx, "eth_getTransactionReceipt", txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt by hash: %v", err)
	}

	var txReceipt models.EVMTransactionReceipt
	err = json.Unmarshal([]byte(*resultJSON), &txReceipt)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction receipt: %v", err)
	}
	return &txReceipt, nil
}

func CreateAccessList(client *rpc.Client, ctx context.Context, tx map[string]interface{}, blockNumber *big.Int) ([]models.EVMAccessList, *string, error) {
	resultJSON, err := CallContextWithJSONResponse(client, ctx, "eth_createAccessList", tx, hexutil.EncodeBig(blockNumber))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create access list: %v", err)
	}

	var accessList []models.EVMAccessList
	err = json.Unmarshal([]byte(gjson.Get(*resultJSON, "accessList").Raw), &accessList)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal access list: %v", err)
	}

	gasUnsigned := gjson.Get(*resultJSON, "gasUsed").String()
	return accessList, &gasUnsigned, nil
}

func GetBlockByNumber(client *rpc.Client, ctx context.Context, blockNumber string) (*models.EVMBlock, *types.Header, error) {
	resultJSON, err := CallContextWithJSONResponse(client, ctx, "eth_getBlockByNumber", blockNumber, false)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get block by number: %v", err)
	}
	var block models.EVMBlock
	err = json.Unmarshal([]byte(*resultJSON), &block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal block: %v", err)
	}
	var header types.Header
	err = json.Unmarshal([]byte(*resultJSON), &header)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal block: %v", err)
	}
	return &block, &header, nil
}

func GetProof(client *rpc.Client, ctx context.Context, address common.Address, storageKeys []string, blockNumber string) (*abstract_types.AccountProof, error) {
	resultJSON, err := CallContextWithJSONResponse(client, ctx, "eth_getProof", address, storageKeys, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get proof: %v", err)
	}

	var proof models.EVMProof
	err = json.Unmarshal([]byte(*resultJSON), &proof)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal proof: %v", err)
	}

	storageProof := make([]abstract_types.StorageProof, len(proof.StorageProof))
	for j, storage := range proof.StorageProof {
		proofBytes := make([][]byte, len(storage.Proof))
		for k, proof := range storage.Proof {
			proofBytes[k] = common.FromHex(proof)
		}
		storageProof[j] = abstract_types.StorageProof{
			Key:   common.HexToHash(storage.Key),
			Value: common.HexToHash(storage.Value),
			Proof: proofBytes,
		}
	}

	accountProof := make([][]byte, len(proof.AccountProof))
	for j, proof := range proof.AccountProof {
		accountProof[j] = common.FromHex(proof)
	}

	return &abstract_types.AccountProof{
		Addr:         common.HexToAddress(proof.Address),
		Balance:      hexutil.MustDecodeBig(proof.Balance),
		Nonce:        hexutil.MustDecodeBig(proof.Nonce),
		CodeHash:     common.HexToHash(proof.CodeHash),
		AccountProof: accountProof,
		StorageProof: storageProof,
		StorageHash:  common.HexToHash(proof.StorageHash),
	}, nil
}

func GetCode(client *rpc.Client, ctx context.Context, address common.Address, blockNumber string) (string, error) {
	resultJSON, err := CallContextWithJSONResponse(client, ctx, "eth_getCode", address, blockNumber)
	if err != nil {
		return "", fmt.Errorf("failed to get code: %v", err)
	}
	return gjson.Parse(*resultJSON).String(), nil
}

func GetChainId(client *rpc.Client, ctx context.Context) (string, error) {
	var chainId string
	err := client.CallContext(ctx, &chainId, "eth_chainId")
	if err != nil {
		return "", fmt.Errorf("failed to get chain ID: %v", err)
	}
	return chainId, nil
}

func SendTransaction(client *rpc.Client, ctx context.Context, tx map[string]interface{}) (string, error) {
	var txHash string
	err := client.CallContext(ctx, &txHash, "eth_sendTransaction", tx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %v", err)
	}
	return txHash, nil
}

func Call(client *rpc.Client, ctx context.Context, tx map[string]interface{}, blockNumber string) (string, error) {
	var result string
	err := client.CallContext(ctx, &result, "eth_call", tx, blockNumber)
	if err != nil {
		return "nil", fmt.Errorf("failed to call: %v", err)
	}
	return result, nil
}

// GetProofsByAccessList gets the proofs for a list of access lists
func GetProofsByAccessList(client *rpc.Client, ctx context.Context, accessList []models.EVMAccessList, blockNumber *big.Int) ([]abstract_types.Account, error) {
	proofs := make([]abstract_types.Account, len(accessList))
	for i, v := range accessList {
		blockNumberHex := hexutil.EncodeBig(blockNumber)
		proof, err := GetProof(client, ctx, v.Address, v.StorageKeys, blockNumberHex)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		code, err := GetCode(client, ctx, v.Address, blockNumberHex)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		proofs[i] = abstract_types.Account{
			Proof: *proof,
			Code:  common.FromHex(code),
		}
	}
	return proofs, nil
}
