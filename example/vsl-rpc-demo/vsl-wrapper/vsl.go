package vsl_wrapper

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"

	types "base-tee/pkg/abstract_types"
	utils "vsl-rpc-demo/utils"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
)

type VSL struct {
	client *rpc.Client
	mtx    sync.Mutex
}

// Balances in VSL, not atto
const DEFAULT_BALANCE int64 = 1000
const MAX_FUNDABLE_BALANCE int64 = 10000

func DialVSL(host string, port string) (*VSL, error) {
	client, err := rpc.Dial(host + ":" + port)
	if err != nil {
		return nil, fmt.Errorf("failed dialing VSL: %w", err)
	}
	return &VSL{client: client, mtx: sync.Mutex{}}, nil
}

// Generates a new keypair+address, and preloads the account with 1000 VSL
// from the bank.
func (vsl *VSL) NewLoadedAccount(bank_priv *ecdsa.PrivateKey, bank_addr string) (string, *ecdsa.PrivateKey, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return "", nil, fmt.Errorf("failed generating private key: %s", err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", nil, fmt.Errorf("error casting public key to ECDSA: %s", err)
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	addr := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	fmt.Println("Public key: ", hexutil.Encode(publicKeyBytes))
	fmt.Println("Private key: ", hexutil.Encode(crypto.FromECDSA(privateKey)))
	fmt.Println("Address: ", addr)

	_, err = vsl.Pay(
		bank_priv,
		bank_addr,
		addr,
		new(big.Int).Mul(big.NewInt(DEFAULT_BALANCE), big.NewInt(1e18)),
	)
	if err != nil {
		return "", nil, fmt.Errorf("error on preloading account: %s", err)
	}
	return addr, privateKey, nil
}

// Sends VSL tokens from the bank to a given address
func (vsl *VSL) FundBalance(addr string, amount *big.Int, bank_priv *ecdsa.PrivateKey, bank_addr string) error {
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	var limit = new(big.Int).Mul(big.NewInt(MAX_FUNDABLE_BALANCE), big.NewInt(1e18))
	if amount.Cmp(limit) > 0 {
		return fmt.Errorf("amount must be less than 10000*10^18")
	}

	// Ensure that the address is not the bank address.
	if addr == bank_addr {
		return fmt.Errorf("cannot fund balance for the bank address itself")
	}
	// Check if the amount is less than 10000 VSL (fixed hard limit for now) and less than the bank balance.
	bankBalance, err := vsl.GetBalance(bank_addr)
	if err != nil {
		return fmt.Errorf("error retrieving bank balance: %s", err)
	}
	if amount.Cmp(bankBalance) > 0 {
		return fmt.Errorf("amount must be less than bank balance")
	}

	_, err = vsl.Pay(
		bank_priv,
		bank_addr,
		addr,
		amount,
	)
	if err != nil {
		return fmt.Errorf("error on calling Pay: %s", err)
	}

	return nil
}

// Amount is specified in atto-VSL (= 10^-18 VSL).
func (vsl *VSL) Pay(key *ecdsa.PrivateKey, from string, to string, amount *big.Int) (string, error) {
	var claimId *string
	// All endpoints that need a nonce must be within critical sections
	vsl.mtx.Lock()
	defer vsl.mtx.Unlock()
	u64Nonce, err := vsl.GetAccountNonce(from)
	if err != nil {
		return "", fmt.Errorf("failed retrieving nonce: %s", err)
	}
	nonce := strconv.FormatUint(u64Nonce, 10)
	strAmount := "0x" + amount.Text(16)
	msg := types.PayMessage{
		Amount: strAmount,
		From:   from,
		Nonce:  nonce,
		To:     to,
	}
	msgBytes, err := rlp.EncodeToBytes(msg)
	if err != nil {
		return "", fmt.Errorf("RLP encoding error: %s", err)
	}
	hash, r, s, v, err := utils.Sign(msgBytes, key)
	if err != nil {
		return "", fmt.Errorf("failed signing: %s", err)
	}
	err = vsl.client.Call(&claimId,
		"vsl_pay",
		types.SignedPayMessage{
			Amount: strAmount,
			From:   from,
			Nonce:  nonce,
			To:     to,
			Hash:   hash,
			R:      r,
			S:      s,
			V:      v,
		},
	)
	if err != nil {
		return "", fmt.Errorf("VSL error: %s", err)
	}
	return *claimId, nil
}

func (vsl *VSL) SubmitClaim(key *ecdsa.PrivateKey, claim string, claimType string, proof string, verifiers []string, client string, expiry_seconds uint64, fee *big.Int) (string, error) {
	var claimId *string
	// All endpoints that need a nonce must be within critical sections
	vsl.mtx.Lock()
	defer vsl.mtx.Unlock()
	exp_timestamp := uint64(time.Now().Unix()) + expiry_seconds
	u64Nonce, err := vsl.GetAccountNonce(client)
	if err != nil {
		return "", fmt.Errorf("failed retrieving nonce: %s", err)
	}
	nonce := strconv.FormatUint(u64Nonce, 10)
	strFee := "0x" + fee.Text(16)
	msg := types.SubmittedClaim{
		Claim:     claim,
		ClaimType: claimType,
		Client:    client,
		Expires: types.Timestamp{
			Seconds:     exp_timestamp,
			Nanoseconds: 0,
		},
		Fee:       strFee,
		Quorum:    1,
		Nonce:     nonce,
		Proof:     proof,
		Verifiers: verifiers,
	}
	msgBytes, err := rlp.EncodeToBytes(msg)
	if err != nil {
		return "", fmt.Errorf("RLP encoding error: %s", err)
	}
	hash, r, s, v, err := utils.Sign(msgBytes, key)
	if err != nil {
		return "", fmt.Errorf("failed signing: %s", err)
	}
	err = vsl.client.Call(&claimId,
		"vsl_submitClaim",
		types.SignedSubmittedClaim{
			Claim:     claim,
			ClaimType: claimType,
			Client:    client,
			Expires: types.Timestamp{
				Seconds:     exp_timestamp,
				Nanoseconds: 0,
			},
			Fee:       strFee,
			Quorum:    1,
			Nonce:     nonce,
			Proof:     proof,
			Verifiers: verifiers,
			Hash:      hash,
			R:         r,
			S:         s,
			V:         v,
		},
	)
	if err != nil {
		return "", fmt.Errorf("RPC error: %w", err)
	}
	return *claimId, nil
}

func (vsl *VSL) Settle(key *ecdsa.PrivateKey, verifierAddr string, claimID string) (string, error) {
	var validatedId string
	// All endpoints that need a nonce must be within critical sections
	vsl.mtx.Lock()
	defer vsl.mtx.Unlock()
	u64Nonce, err := vsl.GetAccountNonce(verifierAddr)
	if err != nil {
		return "", fmt.Errorf("failed retrieving nonce: %s", err)
	}
	nonce := strconv.FormatUint(u64Nonce, 10)
	msg := types.SettleClaimMessage{
		Verifier: verifierAddr,
		Nonce:    nonce,
		ClaimID:  claimID,
	}
	msgBytes, err := rlp.EncodeToBytes(msg)
	if err != nil {
		return "", fmt.Errorf("RLP encoding error: %s", err)
	}
	hash, r, s, v, err := utils.Sign(msgBytes, key)
	if err != nil {
		return "", fmt.Errorf("failed signing: %s", err)
	}
	err = vsl.client.Call(&validatedId,
		"vsl_settleClaim",
		types.SignedSettleClaimMessage{
			Verifier: verifierAddr,
			Nonce:    nonce,
			ClaimID:  claimID,
			Hash:     hash,
			R:        r,
			S:        s,
			V:        v,
		},
	)
	if err != nil {
		return "", fmt.Errorf("RPC error: %w", err)
	}
	return validatedId, nil
}

func (vsl *VSL) PollSettledByID(claimId string, submitTime time.Time, expiry_seconds uint64, loop_interval time.Duration) (types.SignedSettledVerifiedClaim, error) {
	var claim types.TimestampedSignedSettledVerifiedClaim
	err := vsl.client.Call(&claim,
		"vsl_getSettledClaimById",
		claimId,
	)
	for err != nil {
		// Question: Is this proper usage of expiry? Should relying party do this themselves?
		if time.Now().After(submitTime.Add(time.Duration(expiry_seconds) * time.Second)) {
			return types.SignedSettledVerifiedClaim{}, fmt.Errorf("claim not validated before expiry")
		}
		if err != nil {
			time.Sleep(loop_interval * time.Second)
		}
		err = vsl.client.Call(&claim,
			"vsl_getSettledClaimById",
			claimId,
		)
	}
	return claim.Data, nil
}

func (vsl *VSL) ListSubmittedByVerifier(verifier string, since types.Timestamp) ([]types.TimestampedSignedSubmittedClaim, error) {
	var newClaims []types.TimestampedSignedSubmittedClaim
	err := vsl.client.Call(&newClaims,
		"vsl_listSubmittedClaimsForReceiver",
		verifier,
		since,
	)
	if err != nil {
		return newClaims, err
	}
	return newClaims, nil
}

// Balance is returned in atto-VSL (= 10^-18 VSL).
func (vsl *VSL) GetBalance(addr string) (*big.Int, error) {
	var reply string
	err := vsl.client.Call(&reply,
		"vsl_getBalance",
		addr,
	)
	if err != nil {
		return big.NewInt(0), err
	}
	balance, ok := new(big.Int).SetString(reply, 0)
	if !ok {
		return big.NewInt(0), fmt.Errorf("failed converting VSL balance to big.Int")
	}
	return balance, nil
}

func (vsl *VSL) GetAccountNonce(addr string) (uint64, error) {
	var reply uint64
	err := vsl.client.Call(&reply,
		"vsl_getAccountNonce",
		addr,
	)
	if err != nil {
		return 0, err
	}
	return reply, nil
}

func (vsl *VSL) Close() {
	vsl.client.Close()
}
