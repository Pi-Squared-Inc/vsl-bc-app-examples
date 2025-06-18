package models

import (
	"context"
	"log"
	"math/big"
	"os"

	"base/pkg/vsl"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

type App struct {
	API                          *fiber.App
	Port                         string
	BackendAPIEndpoint           string
	SourceChainRPCEndpoint       string
	SourceChainWebsocketEndpoint string
	SourceVSLContractFunction    string
	SourceVSLContractAddress     *common.Address
	SourceVSLContractABIJSON     string
	RPCClient                    *rpc.Client
	EthRPCClient                 *ethclient.Client
	EthWSClient                  *ethclient.Client
	ChainId                      *big.Int
	GethClient                   *gethclient.Client
	VSLRPC                       string
	VSLRPCClient                 *vsl.VSLRPCClient
	VSLClientAddress             string
	VSLClientPrivateKey          string
	VSLVerifierAddress           string
	VSLVerifierPrivateKey        string
}

func NewApp() *App {
	port := os.Getenv("PORT")
	sourceChainRPCEndpoint := os.Getenv("SOURCE_RPC_ENDPOINT")
	sourceChainWebsocketEndpoint := os.Getenv("SOURCE_WEBSOCKET_ENDPOINT")
	backendAPIEndpoint := os.Getenv("BACKEND_API_ENDPOINT")

	log.Printf("Start observing chain for state query claims")
	sourceVSLContractFunction := os.Getenv("SOURCE_VSL_CONTRACT_FUNCTION")
	sourceVSLContractAddressHex := os.Getenv("SOURCE_VSL_CONTRACT_ADDRESS")
	sourceVSLContractAddress := common.HexToAddress(sourceVSLContractAddressHex)
	sourceVSLContractABIJSON := os.Getenv("SOURCE_VSL_CONTRACT_ABI_JSON")

	// VSL
	vslRPC := os.Getenv("VSL_RPC")
	vslClientAddress := os.Getenv("VSL_CLIENT_ADDRESS")
	vslClientPrivateKey := os.Getenv("VSL_CLIENT_PRIVATE_KEY")
	vslVerifierAddress := os.Getenv("VSL_VERIFIER_ADDRESS")
	vslVerifierPrivateKey := os.Getenv("VSL_VERIFIER_PRIVATE_KEY")

	rpcClient, err := rpc.Dial(sourceChainRPCEndpoint)
	if err != nil {
		log.Fatalf("Failed to create RPC client: %+v", err)
	}

	ethRPCClient := ethclient.NewClient(rpcClient)
	ethWSClient, err := ethclient.Dial(sourceChainWebsocketEndpoint)
	if err != nil {
		log.Fatalf("Failed to create WS client: %+v", err)
	}

	gethClient := gethclient.New(rpcClient)

	ctx := context.Background()
	chainId, err := ethRPCClient.ChainID(ctx)
	if err != nil {
		log.Fatalf("Failed to get chain ID: %+v", err)
	}

	fiberApp := fiber.New()
	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
	}))

	app := &App{
		API:                          fiberApp,
		Port:                         port,
		BackendAPIEndpoint:           backendAPIEndpoint,
		SourceChainRPCEndpoint:       sourceChainRPCEndpoint,
		SourceChainWebsocketEndpoint: sourceChainWebsocketEndpoint,
		SourceVSLContractFunction:    sourceVSLContractFunction,
		SourceVSLContractAddress:     &sourceVSLContractAddress,
		SourceVSLContractABIJSON:     sourceVSLContractABIJSON,
		RPCClient:                    rpcClient,
		EthRPCClient:                 ethRPCClient,
		EthWSClient:                  ethWSClient,
		GethClient:                   gethClient,
		ChainId:                      chainId,
		VSLRPC:                       vslRPC,
		VSLRPCClient:                 vsl.NewVSLRPCClient(vslRPC, vslClientPrivateKey),
		VSLClientAddress:             vslClientAddress,
		VSLClientPrivateKey:          vslClientPrivateKey,
		VSLVerifierAddress:           vslVerifierAddress,
		VSLVerifierPrivateKey:        vslVerifierPrivateKey,
	}

	return app
}
