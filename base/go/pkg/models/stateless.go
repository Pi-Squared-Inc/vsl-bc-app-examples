package models

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/types"
)

// RethWitness represents the witness that generate from Reth node
type RethWitness struct {
	Codes map[string]string `json:"codes"`
	Keys  map[string]string `json:"keys"`
	State map[string]string `json:"state"`
}

func (w *RethWitness) ToStatelessWitness(pastHeader *types.Header, header *types.Header) *stateless.Witness {
	witness, err := stateless.NewWitness(header, nil)
	if err != nil {
		panic(err)
	}

	witness.Headers = []*types.Header{pastHeader}

	for _, v := range w.Codes {
		decoded, err := hexutil.Decode(v)
		if err != nil {
			panic(err)
		}
		witness.AddCode(decoded)
	}

	for _, v := range w.State {
		decoded, err := hexutil.Decode(v)
		if err != nil {
			panic(err)
		}
		witness.AddState(map[string]struct{}{
			string(decoded): {},
		})
	}

	return witness
}

// GethWitness represents the witness that generate from Geth node
type GethWitness struct {
	Headers []*types.Header   `json:"headers"`
	Codes   map[string]string `json:"codes"`
	State   map[string]string `json:"state"`
}

func (w *GethWitness) ToStatelessWitness(header *types.Header) *stateless.Witness {
	witness, err := stateless.NewWitness(header, nil)
	if err != nil {
		panic(err)
	}
	witness.Headers = w.Headers

	for _, v := range w.Codes {
		decoded, err := hexutil.Decode(v)
		if err != nil {
			panic(err)
		}
		witness.AddCode(decoded)
	}

	for _, v := range w.State {
		decoded, err := hexutil.Decode(v)
		if err != nil {
			panic(err)
		}
		witness.AddState(map[string]struct{}{
			string(decoded): {},
		})
	}

	return witness
}
