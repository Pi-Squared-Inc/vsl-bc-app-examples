package models

import (
	"base/pkg/vsl"
	"os"
)

type App struct {
	BackendEndpoint    string
	VerifierAddress    string
	VerifierPrivateKey string
	VSLRPC             string
	VSLRPCClient       *vsl.VSLRPCClient
}

func NewApp() *App {
	backendEndpoint := os.Getenv("BACKEND_ENDPOINT")

	// VSL
	vslRPC := os.Getenv("VSL_RPC")
	vslVerifierAddress := os.Getenv("VSL_VERIFIER_ADDRESS")
	vslVerifierPrivateKey := os.Getenv("VSL_VERIFIER_PRIVATE_KEY")
	vslRPCClient := vsl.NewVSLRPCClient(vslRPC, vslVerifierPrivateKey)

	app := &App{
		BackendEndpoint:    backendEndpoint,
		VerifierAddress:    vslVerifierAddress,
		VerifierPrivateKey: vslVerifierPrivateKey,
		VSLRPC:             vslRPC,
		VSLRPCClient:       vslRPCClient,
	}

	return app
}
