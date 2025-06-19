package verification

import (
	"base/pkg/abstract_types"
	"base/pkg/evm"
	"bytes"
	"generation-view-fn-evm/pkg/models"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/pkg/errors"
)

// Verify validates a view function claim
func Verify(claim *models.EVMViewFnClaim, verificationContext *models.EVMViewFnClaimVerificationContext) error {
	header := claim.Assumptions
	evmCall := claim.Action
	result := claim.Result

	evm, _, err := evm.CreateEVM(claim.Metadata.ChainId, header.Root, header.ToGethHeader(), verificationContext.Accounts, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	// Apply query
	localOutput, err := callQuery(evm, evmCall, header.GasLimit.Uint64())
	if err != nil {
		return errors.WithStack(err)
	}

	// Compare output
	err = compareOutput(localOutput, result)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// callQuery calls the query on the local EVM
func callQuery(evm *vm.EVM, evmCall *abstract_types.EVMCall, gasLimit uint64) ([]byte, error) {
	result, _, err := evm.StaticCall(
		evmCall.From,
		evmCall.To,
		evmCall.Input,
		gasLimit,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return result, nil
}

// compareOutput compares the output with the expected output
func compareOutput(output, expectedOutput []byte) error {
	if bytes.Equal(output, expectedOutput) {
		return nil
	}
	return errors.New("output does not match expected output")
}
