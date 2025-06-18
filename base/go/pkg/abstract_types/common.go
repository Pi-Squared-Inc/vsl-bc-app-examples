package abstract_types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type EVMCall struct {
	// From is the address of the caller
	From common.Address `json:"from"`
	// To is the address of the contract
	To common.Address `json:"to"`
	// Input is the input data of the query
	Input []byte `json:"input"`
}

type EVMMetadata struct {
	ChainId *big.Int `json:"chainId"`
}

type Account struct {
	Proof AccountProof `json:"proof"`
	Code  []byte       `json:"code"`
}

type AccountProof struct {
	Addr         common.Address
	AccountProof [][]byte
	Balance      *big.Int
	CodeHash     common.Hash
	Nonce        *big.Int
	StorageHash  common.Hash
	StorageProof []StorageProof
}

type StorageProof struct {
	Key   [32]byte
	Value [32]byte
	Proof [][]byte
}

type Header struct {
	ParentHash  common.Hash    `json:"parentHash"       gencodec:"required"`
	UncleHash   common.Hash    `json:"sha3Uncles"       gencodec:"required"`
	Coinbase    common.Address `json:"miner"`
	Root        common.Hash    `json:"stateRoot"        gencodec:"required"`
	TxHash      common.Hash    `json:"transactionsRoot" gencodec:"required"`
	ReceiptHash common.Hash    `json:"receiptsRoot"     gencodec:"required"`
	Bloom       []byte         `json:"logsBloom"        gencodec:"required"`
	Difficulty  *big.Int       `json:"difficulty"       gencodec:"required"`
	Number      *big.Int       `json:"number"           gencodec:"required"`
	GasLimit    *big.Int       `json:"gasLimit"         gencodec:"required"`
	GasUsed     *big.Int       `json:"gasUsed"          gencodec:"required"`
	Time        *big.Int       `json:"timestamp"        gencodec:"required"`
	Extra       []byte         `json:"extraData"        gencodec:"required"`
	MixDigest   common.Hash    `json:"mixHash"`
	Nonce       [8]byte        `json:"nonce"`
}

func (h *Header) ToGethHeader() *types.Header {
	return &types.Header{
		ParentHash:  h.ParentHash,
		UncleHash:   h.UncleHash,
		Coinbase:    h.Coinbase,
		Root:        h.Root,
		TxHash:      h.TxHash,
		ReceiptHash: h.ReceiptHash,
		Bloom:       types.Bloom(h.Bloom),
		Difficulty:  h.Difficulty,
		Number:      h.Number,
		GasLimit:    h.GasLimit.Uint64(),
		GasUsed:     h.GasUsed.Uint64(),
		Time:        h.Time.Uint64(),
		Extra:       h.Extra,
		MixDigest:   h.MixDigest,
		Nonce:       h.Nonce,
	}
}
