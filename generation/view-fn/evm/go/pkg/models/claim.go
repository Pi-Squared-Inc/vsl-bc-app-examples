package models

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"base/pkg/abstract_types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

const (
	EVMViewFnClaimEncodeAbiJSON = `[{"type":"function","name":"encode","inputs":[{"name":"","type":"tuple","internalType":"structEVMViewFnClaimVerifier.EVMViewFnClaim","components":[{"name":"claimType","type":"string","internalType":"string"},{"name":"trustBaseSpec","type":"string","internalType":"string"},{"name":"assumptions","type":"tuple","internalType":"structEVMViewFnClaimVerifier.Header","components":[{"name":"parentHash","type":"bytes32","internalType":"bytes32"},{"name":"uncleHash","type":"bytes32","internalType":"bytes32"},{"name":"coinbase","type":"address","internalType":"address"},{"name":"root","type":"bytes32","internalType":"bytes32"},{"name":"txHash","type":"bytes32","internalType":"bytes32"},{"name":"receiptHash","type":"bytes32","internalType":"bytes32"},{"name":"bloom","type":"bytes","internalType":"bytes"},{"name":"difficulty","type":"uint256","internalType":"uint256"},{"name":"number","type":"uint256","internalType":"uint256"},{"name":"gasLimit","type":"uint256","internalType":"uint256"},{"name":"gasUsed","type":"uint256","internalType":"uint256"},{"name":"time","type":"uint256","internalType":"uint256"},{"name":"extra","type":"bytes","internalType":"bytes"},{"name":"mixDigest","type":"bytes32","internalType":"bytes32"},{"name":"nonce","type":"bytes8","internalType":"bytes8"}]},{"name":"action","type":"tuple","internalType":"structEVMViewFnClaimVerifier.EVMCall","components":[{"name":"from","type":"address","internalType":"address"},{"name":"to","type":"address","internalType":"address"},{"name":"input","type":"bytes","internalType":"bytes"}]},{"name":"result","type":"bytes","internalType":"bytes"},{"name":"metadata","type":"tuple","internalType":"structEVMViewFnClaimVerifier.EVMMetadata","components":[{"name":"chainId","type":"uint256","internalType":"uint256"}]}]},{"name":"","type":"tuple","internalType":"structEVMViewFnClaimVerifier.EVMViewFnClaimVerificationData","components":[{"name":"accounts","type":"tuple[]","internalType":"structEVMViewFnClaimVerifier.Account[]","components":[{"name":"proof","type":"tuple","internalType":"structEVMViewFnClaimVerifier.AccountProof","components":[{"name":"addr","type":"address","internalType":"address"},{"name":"accountProof","type":"bytes[]","internalType":"bytes[]"},{"name":"balance","type":"uint256","internalType":"uint256"},{"name":"codeHash","type":"bytes32","internalType":"bytes32"},{"name":"nonce","type":"uint256","internalType":"uint256"},{"name":"storageHash","type":"bytes32","internalType":"bytes32"},{"name":"storageProof","type":"tuple[]","internalType":"structEVMViewFnClaimVerifier.StorageProof[]","components":[{"name":"key","type":"bytes32","internalType":"bytes32"},{"name":"value","type":"bytes32","internalType":"bytes32"},{"name":"proof","type":"bytes[]","internalType":"bytes[]"}]}]},{"name":"code","type":"bytes","internalType":"bytes"}]}]}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"pure"}]`
)

type EVMViewFnClaim struct {
	ClaimType     string                     `json:"type"`
	TrustBaseSpec string                     `json:"trustBaseSpec"`
	Assumptions   *abstract_types.Header     `json:"assumptions"`
	Action        *abstract_types.EVMCall    `json:"action"`
	Result        []byte                     `json:"result"`
	Metadata      abstract_types.EVMMetadata `json:"metadata"`
}

// TODO: Properly implement this
func (c *EVMViewFnClaim) GetId() (*string, error) {
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		fmt.Printf("Error encoding to JSON: %v\n", err)
		return nil, errors.New("error encoding to JSON")
	}

	hash := sha3.NewLegacyKeccak256()
	hash.Write(jsonBytes)
	hashString := hex.EncodeToString(hash.Sum(nil))
	return &hashString, nil
}

// VerificationContext for EVMViewFnClaim: the account proofs of the pre-state
type EVMViewFnClaimVerificationContext struct {
	Accounts []abstract_types.Account `json:"accounts"`
}

func GetAbi() abi.ABI {
	contractAbi, err := abi.JSON(strings.NewReader(EVMViewFnClaimEncodeAbiJSON))
	if err != nil {
		panic(err)
	}

	return contractAbi
}

func (c *EVMViewFnClaim) AbiEncode() ([]byte, error) {
	encodeAbi, err := abi.JSON(strings.NewReader(EVMViewFnClaimEncodeAbiJSON))
	if err != nil {
		panic(err)
	}

	method := encodeAbi.Methods["encode"]
	encoded, err := method.Inputs[:1].Pack(c)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func AbiDecodeEVMViewFnClaim(claimData []byte) (*EVMViewFnClaim, error) {
	abi := GetAbi()
	claim_abi := abi.Methods["encode"].Inputs[:1]
	values, err := claim_abi.UnpackValues(claimData)
	if err != nil {
		return nil, err
	}
	if len(values) != 1 {
		panic("Unexpected amount of values")
	}

	decodedClaim := reflect.ValueOf(values[0])

	// Assumptions
	assumptions := decodedClaim.FieldByName("Assumptions")
	parentHash := assumptions.FieldByName("ParentHash").Interface().([32]uint8)
	uncleHash := assumptions.FieldByName("UncleHash").Interface().([32]uint8)
	nonce := assumptions.FieldByName("Nonce").Interface().([8]uint8)
	root := assumptions.FieldByName("Root").Interface().([32]uint8)
	miner := assumptions.FieldByName("Coinbase").Interface().(common.Address)
	txHash := assumptions.FieldByName("TxHash").Interface().([32]uint8)
	receiptHash := assumptions.FieldByName("ReceiptHash").Interface().([32]uint8)

	// Action
	action := decodedClaim.FieldByName("Action")
	from := action.FieldByName("From").Interface().(common.Address)
	to := action.FieldByName("To").Interface().(common.Address)
	input := action.FieldByName("Input").Bytes()

	// Result
	resultBytes := decodedClaim.FieldByName("Result").Bytes()

	// Metadata
	metadata := decodedClaim.FieldByName("Metadata")
	chainId := metadata.FieldByName("ChainId").Interface().(*big.Int)

	// Construct claim
	claim := EVMViewFnClaim{
		ClaimType:     "EVMViewFn",
		TrustBaseSpec: decodedClaim.FieldByName("TrustBaseSpec").String(),
		Assumptions: &abstract_types.Header{
			ParentHash:  common.BytesToHash(parentHash[:]),
			UncleHash:   common.BytesToHash(uncleHash[:]),
			Coinbase:    miner,
			Root:        common.BytesToHash(root[:]),
			TxHash:      common.BytesToHash(txHash[:]),
			ReceiptHash: common.BytesToHash(receiptHash[:]),
			Bloom:       assumptions.FieldByName("Bloom").Bytes(),
			Difficulty:  assumptions.FieldByName("Difficulty").Interface().(*big.Int),
			Number:      assumptions.FieldByName("Number").Interface().(*big.Int),
			GasLimit:    assumptions.FieldByName("GasLimit").Interface().(*big.Int),
			GasUsed:     assumptions.FieldByName("GasUsed").Interface().(*big.Int),
			Time:        assumptions.FieldByName("Time").Interface().(*big.Int),
			Extra:       assumptions.FieldByName("Extra").Bytes(),
			MixDigest:   common.HexToHash(assumptions.FieldByName("MixDigest").String()),
			Nonce:       nonce,
		},
		Action: &abstract_types.EVMCall{
			From:  from,
			To:    to,
			Input: input,
		},
		Result: resultBytes,
		Metadata: abstract_types.EVMMetadata{
			ChainId: chainId,
		},
	}

	return &claim, nil
}
func (c *EVMViewFnClaimVerificationContext) AbiEncode() ([]byte, error) {
	encodeAbi, err := abi.JSON(strings.NewReader(EVMViewFnClaimEncodeAbiJSON))
	if err != nil {
		panic(err)
	}

	method := encodeAbi.Methods["encode"]
	encoded, err := method.Inputs[1:].Pack(c) // Slice the second input argument and pack it
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func AbiDecodeEVMViewFnClaimVerificationContext(verificationData []byte) (*EVMViewFnClaimVerificationContext, error) {
	abi := GetAbi()

	claim_abi := abi.Methods["encode"].Inputs[1:]
	values, err := claim_abi.UnpackValues(verificationData)
	if err != nil {
		return nil, err
	}
	if len(values) != 1 {
		panic("Unexpected amount of values")
	}

	decodedVerData := reflect.ValueOf(values[0])

	decodedAccounts := decodedVerData.FieldByName("Accounts")
	accounts := []abstract_types.Account{}
	for i := 0; i < decodedAccounts.Len(); i++ {
		decodedAccount := decodedAccounts.Index(i)

		// Unpack the proof
		decodedProof := decodedAccount.FieldByName("Proof")
		decodedAddr := decodedProof.FieldByName("Addr").Interface().(common.Address)
		decodedAccountProof := decodedProof.FieldByName("AccountProof").Interface().([][]byte)
		decodedBalance := decodedProof.FieldByName("Balance").Interface().(*big.Int)
		decodedCodeHash := decodedProof.FieldByName("CodeHash").Interface().([32]uint8)
		decodedNonce := decodedProof.FieldByName("Nonce").Interface().(*big.Int)
		decodedStorageHash := decodedProof.FieldByName("StorageHash").Interface().([32]uint8)
		decodedStorageProof := decodedProof.FieldByName("StorageProof")
		storageProof := []abstract_types.StorageProof{}
		for j := 0; j < decodedStorageProof.Len(); j++ {
			decodedStorage := decodedStorageProof.Index(j)
			decodedKey := decodedStorage.FieldByName("Key").Interface().([32]uint8)
			decodedValue := decodedStorage.FieldByName("Value").Interface().([32]uint8)
			decodedProof := decodedStorage.FieldByName("Proof").Interface().([][]byte)
			storageProof = append(storageProof, abstract_types.StorageProof{
				Key:   common.BytesToHash(decodedKey[:]),
				Value: common.BytesToHash(decodedValue[:]),
				Proof: decodedProof,
			})
		}

		// Append the account
		accounts = append(accounts, abstract_types.Account{
			Proof: abstract_types.AccountProof{
				Addr:         decodedAddr,
				AccountProof: decodedAccountProof,
				Balance:      decodedBalance,
				CodeHash:     decodedCodeHash,
				Nonce:        decodedNonce,
				StorageHash:  decodedStorageHash,
				StorageProof: storageProof,
			},
			Code: decodedAccount.FieldByName("Code").Bytes(),
		})
	}

	return &EVMViewFnClaimVerificationContext{Accounts: accounts}, nil
}
