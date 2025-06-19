package models

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/samber/lo"
)

// EVMProof is the model for the RPC response of eth_getProof
type EVMProof struct {
	Address      string            `json:"address"`
	AccountProof []string          `json:"accountProof"`
	Balance      string            `json:"balance"`
	CodeHash     string            `json:"codeHash"`
	Nonce        string            `json:"nonce"`
	StorageHash  string            `json:"storageHash"`
	StorageProof []EVMStorageProof `json:"storageProof"`
}

type EVMStorageProof struct {
	Key   string   `json:"key"`
	Value string   `json:"value"`
	Proof []string `json:"proof"`
}

// EVMAccessList is the model for the RPC response of eth_createAccessList
type EVMAccessList struct {
	Address     common.Address `json:"address"`
	StorageKeys []string       `json:"storageKeys"`
}

// EVMTransaction is the model for the RPC response of eth_getTransactionByHash
type EVMTransaction struct {
	Hash                 string          `json:"hash"`
	Nonce                string          `json:"nonce"`
	BlockHash            string          `json:"blockHash"`
	BlockNumber          string          `json:"blockNumber"`
	TransactionIndex     string          `json:"transactionIndex"`
	From                 string          `json:"from"`
	To                   string          `json:"to"`
	Value                string          `json:"value"`
	GasPrice             string          `json:"gasPrice"`
	Gas                  string          `json:"gas"`
	Input                string          `json:"input"`
	V                    string          `json:"v"`
	R                    string          `json:"r"`
	S                    string          `json:"s"`
	Type                 string          `json:"type"`
	MaxPriorityFeePerGas string          `json:"maxPriorityFeePerGas"`
	MaxFeePerGas         string          `json:"maxFeePerGas"`
	ChainId              string          `json:"chainId"`
	AccessList           []EVMAccessList `json:"accessList"`
}

func (t *EVMTransaction) GetChainId() (*big.Int, error) {
	chainId, err := hexutil.DecodeBig(t.ChainId)
	if err != nil {
		return nil, err
	}
	return chainId, nil
}

// ToMessage converts the Transaction to a core.Message for EVM usage
func (t *EVMTransaction) ToMessage() (*core.Message, error) {
	txTo := common.HexToAddress(t.To)
	txFrom := common.HexToAddress(t.From)
	txNonce := common.HexToHash(t.Nonce).Big()
	txGas := common.HexToHash(t.Gas).Big()
	txGasPrice := common.HexToHash(t.GasPrice).Big()
	txValue := common.HexToHash(t.Value).Big()
	txInput := common.FromHex(t.Input)
	txMaxPriorityFeePerGas := common.HexToHash(t.MaxPriorityFeePerGas).Big()
	txMaxFeePerGas := common.HexToHash(t.MaxFeePerGas).Big()

	switch t.Type {
	case "0x2":
		// EIP-1559 transaction
		convertedAccessList := lo.Map( // Convert AccessList to AccessTuple
			t.AccessList,
			func(item EVMAccessList, index int) types.AccessTuple {
				return types.AccessTuple{
					Address: item.Address,
					StorageKeys: lo.Map(item.StorageKeys, func(key string, index int) common.Hash {
						return common.HexToHash(key)
					}),
				}
			},
		)

		return &core.Message{
			From:       txFrom,
			To:         &txTo,
			Value:      txValue,
			Nonce:      txNonce.Uint64(),
			Data:       txInput,
			AccessList: convertedAccessList,
			GasTipCap:  txMaxPriorityFeePerGas,
			GasFeeCap:  txMaxFeePerGas,
			GasLimit:   txGas.Uint64(),
			GasPrice:   txGasPrice,
		}, nil
	default:
		// Legacy transaction
		return &core.Message{
			From:     txFrom,
			To:       &txTo,
			Value:    txValue,
			Nonce:    txNonce.Uint64(),
			Data:     txInput,
			GasLimit: txGas.Uint64(),
			GasPrice: txGasPrice,
		}, nil
	}
}

// TransactionReceipt is the model for the RPC response of eth_getTransactionReceipt, currently only used on the test files
type EVMTransactionReceipt struct {
	Hash                 string          `json:"hash"`
	Nonce                string          `json:"nonce"`
	BlockHash            string          `json:"blockHash"`
	BlockNumber          string          `json:"blockNumber"`
	TransactionIndex     string          `json:"transactionIndex"`
	From                 string          `json:"from"`
	To                   string          `json:"to"`
	Value                string          `json:"value"`
	GasPrice             string          `json:"gasPrice"`
	GasUsed              string          `json:"gasUsed"`
	MaxFeePerGas         string          `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string          `json:"maxPriorityFeePerGas"`
	Input                string          `json:"input"`
	R                    string          `json:"r"`
	S                    string          `json:"s"`
	V                    string          `json:"v"`
	YParity              string          `json:"yParity"`
	ChainId              string          `json:"chainId"`
	AccessList           []EVMAccessList `json:"accessList"`
	Type                 string          `json:"type"`
	ContractAddress      string          `json:"contractAddress"`
}

// EVMBlock is the model for the RPC response of eth_getBlockByHash and eth_getBlockByNumber
type EVMBlock struct {
	Hash                  string   `json:"hash"`
	ParentHash            string   `json:"parentHash"`
	Sha3Uncles            string   `json:"sha3Uncles"`
	Miner                 string   `json:"miner"`
	StateRoot             string   `json:"stateRoot"`
	TransactionsRoot      string   `json:"transactionsRoot"`
	ReceiptsRoot          string   `json:"receiptsRoot"`
	Number                string   `json:"number"`
	GasUsed               string   `json:"gasUsed"`
	GasLimit              string   `json:"gasLimit"`
	ExtraData             string   `json:"extraData"`
	LogsBloom             string   `json:"logsBloom"`
	Timestamp             string   `json:"timestamp"`
	Difficulty            string   `json:"difficulty"`
	TotalDifficulty       string   `json:"totalDifficulty"`
	SealField             []string `json:"sealField"`
	Uncles                []string `json:"uncles"`
	Transactions          []string `json:"transactions"`
	Size                  string   `json:"size"`
	MixHash               string   `json:"mixHash"`
	Nonce                 string   `json:"nonce"`
	BaseFeePerGas         string   `json:"baseFeePerGas"`
	WithdrawalsRoot       string   `json:"withdrawalsRoot"`
	ParentBeaconBlockRoot string   `json:"parentBeaconBlockRoot"`
	BlobGasUsed           string   `json:"blobGasUsed"`
	ExcessBlobGas         string   `json:"excessBlobGas"`
}
