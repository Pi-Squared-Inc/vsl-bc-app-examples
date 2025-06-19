package models

import (
	"log"
	"os"

	"base/pkg/vsl"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

type App struct {
	BackendEndpoint        string
	VSLRPC                 string
	VSLSubmitterAddress    string
	VSLSubmitterPrivateKey string
	VSLVerifierAddress     string
	VSLVerifierPrivateKey  string
	EthRPCClient           *ethclient.Client
	EthWSClient            *ethclient.Client
	VSLClient              *vsl.VSLRPCClient
}

func NewApp() (*App, error) {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) { // Ignore if .env file does not exist
		return nil, errors.WithStack(err)
	}

	backendEndpoint := os.Getenv("BACKEND_ENDPOINT")
	vslRPC := os.Getenv("VSL_RPC")
	vslSubmitterAddress := os.Getenv("VSL_SUBMITTER_ADDRESS")
	vslSubmitterPrivateKey := os.Getenv("VSL_SUBMITTER_PRIVATE_KEY")
	vslVerifierAddress := os.Getenv("VSL_VERIFIER_ADDRESS")
	vslVerifierPrivateKey := os.Getenv("VSL_VERIFIER_PRIVATE_KEY")
	rpcEndpoint := os.Getenv("SOURCE_RPC_ENDPOINT")
	wsEndpoint := os.Getenv("SOURCE_WEBSOCKET_ENDPOINT")

	rpcClient, err := rpc.Dial(rpcEndpoint)
	if err != nil {
		log.Fatalf("Failed to create RPC client: %+v", err)
	}

	ethRPCClient := ethclient.NewClient(rpcClient)
	ethWSClient, err := ethclient.Dial(wsEndpoint)
	if err != nil {
		log.Fatalf("Failed to create WS client: %+v", err)
	}

	vslClient := vsl.NewVSLRPCClient(vslRPC, vslSubmitterPrivateKey)

	return &App{
		BackendEndpoint:        backendEndpoint,
		VSLRPC:                 vslRPC,
		VSLClient:              vslClient,
		VSLSubmitterAddress:    vslSubmitterAddress,
		VSLSubmitterPrivateKey: vslSubmitterPrivateKey,
		VSLVerifierAddress:     vslVerifierAddress,
		VSLVerifierPrivateKey:  vslVerifierPrivateKey,
		EthRPCClient:           ethRPCClient,
		EthWSClient:            ethWSClient,
	}, nil
}
