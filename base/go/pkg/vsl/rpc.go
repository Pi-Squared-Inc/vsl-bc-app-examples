package vsl

import (
	"base/pkg/abstract_types"
	"base/pkg/evm"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/v3/client"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type VSLRPCClient struct {
	rpc        string
	privateKey string
	rpcClient  *client.Client
}

func NewVSLRPCClient(rpc string, privateKey string) *VSLRPCClient {
	return &VSLRPCClient{
		rpc:        rpc,
		privateKey: privateKey,
		rpcClient:  client.New(),
	}
}

func (c *VSLRPCClient) CallRaw(method string, params interface{}) (*gjson.Result, error) {
	response, err := c.rpcClient.Post(c.rpc, client.Config{
		Body: fiber.Map{
			"jsonrpc": "2.0",
			"method":  method,
			"params":  params,
			"id":      1,
		},
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	responseBody := gjson.ParseBytes(response.Body())

	if responseBody.Get("error").Exists() {
		return nil, errors.New(responseBody.Get("error").Raw)
	}

	return &responseBody, nil
}

type CreateAccountParams struct {
	OwnerAddress string `json:"owner_address"`
	Script       string `json:"script"`
	Label        string `json:"label"`
}

type SignedCreateAccountParams struct {
	CreateAccountParams
	evm.SignedComponents
}

func (c *VSLRPCClient) CreateAccount(params CreateAccountParams) (*string, error) {
	signedMessage, err := evm.SignMessage(c.privateKey, params)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	response, err := c.CallRaw("vsl_createAccount", fiber.Map{
		"account_data": SignedCreateAccountParams{
			CreateAccountParams: params,
			SignedComponents:    *signedMessage,
		},
	})
	if err != nil {
		return nil, err
	}

	result := response.Get("result").String()
	return &result, nil
}

type PayParams struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
	Nonce  string `json:"nonce"`
}

type SignedPayParams struct {
	PayParams
	evm.SignedComponents
}

func (c *VSLRPCClient) Pay(params PayParams) (*string, error) {
	signedMessage, err := evm.SignMessage(c.privateKey, params)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	response, err := c.CallRaw("vsl_pay", fiber.Map{
		"payment": SignedPayParams{
			PayParams:        params,
			SignedComponents: *signedMessage,
		},
	})
	if err != nil {
		return nil, err
	}

	result := response.Get("result").String()
	return &result, nil
}

type SubmitClaimParams struct {
	Claim     string                   `json:"claim"`
	ClaimType string                   `json:"claim_type"`
	Proof     string                   `json:"proof"`
	Nonce     string                   `json:"nonce"`
	To        []string                 `json:"to"`
	Quorum    uint16                   `json:"quorum"`
	From      string                   `json:"from"`
	Expires   abstract_types.Timestamp `json:"expires"`
	Fee       string                   `json:"fee"`
}

type SignedSubmitClaimParams struct {
	SubmitClaimParams
	evm.SignedComponents
}

func (c *VSLRPCClient) SubmitClaim(params SubmitClaimParams) (*string, error) {
	signedMessage, err := evm.SignMessage(c.privateKey, params)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	response, err := c.CallRaw("vsl_submitClaim", fiber.Map{
		"claim": SignedSubmitClaimParams{
			SubmitClaimParams: params,
			SignedComponents:  *signedMessage,
		},
	})
	if err != nil {
		return nil, err
	}

	claimId := response.Get("result").String()
	return &claimId, nil
}

type SettleClaimParams struct {
	From          string `json:"from"`
	Nonce         string `json:"nonce"`
	TargetClaimId string `json:"target_claim_id"`
}

type SignedSettleClaimParams struct {
	SettleClaimParams
	evm.SignedComponents
}

func (c *VSLRPCClient) SettleClaim(params SettleClaimParams) (*string, error) {
	signedMessage, err := evm.SignMessage(c.privateKey, params)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	response, err := c.CallRaw("vsl_settleClaim", fiber.Map{
		"settled_claim": SignedSettleClaimParams{
			SettleClaimParams: params,
			SignedComponents:  *signedMessage,
		},
	})
	if err != nil {
		return nil, err
	}

	claimId := response.Get("result").String()
	return &claimId, nil
}

type GetAccountNonceParams struct {
	AccountId string `json:"account_id"`
}

func (c *VSLRPCClient) GetAccountNonce(params GetAccountNonceParams) (*uint64, error) {
	response, err := c.CallRaw("vsl_getAccountNonce", params)
	if err != nil {
		return nil, err
	}
	nonce := response.Get("result").Uint()
	return &nonce, nil
}

type ListSubmittedClaimsForReceiverParams struct {
	Address string                   `json:"address"`
	Since   abstract_types.Timestamp `json:"since"`
}

func (c *VSLRPCClient) ListSubmittedClaimsForReceiver(params ListSubmittedClaimsForReceiverParams) ([]gjson.Result, error) {
	response, err := c.CallRaw("vsl_listSubmittedClaimsForReceiver", params)
	if err != nil {
		return nil, err
	}
	return response.Get("result").Array(), nil
}
