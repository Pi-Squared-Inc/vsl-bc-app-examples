package client

import (
	"crypto/ecdsa"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"vsl-rpc-demo/cmd"

	vsl "vsl-rpc-demo/vsl-wrapper"
)

type App struct {
	VSLClient        *vsl.VSL
	VerifierAddress  string
	ClientAddress    string
	ClientPrivateKey *ecdsa.PrivateKey
	ExpirySeconds    uint64
	LoopInterval     time.Duration
	Fee              *big.Int
	ZeroNonce        bool
}

var (
	Zero_Nonce bool
	Fee        uint64
)

func NewApp(rpc_host string, rpc_port string, verifier_addr string, client_addr string, client_priv *ecdsa.PrivateKey, expiry_seconds uint64, loop_interval time.Duration, fee *big.Int, zero_nonce bool) *App {
	vslClient, err := vsl.DialVSL(rpc_host, rpc_port)
	if err != nil {
		panic("failed to connect to VSL server")
	}
	return &App{
		VSLClient:        vslClient,
		VerifierAddress:  verifier_addr,
		ClientAddress:    client_addr,
		ClientPrivateKey: client_priv,
		ExpirySeconds:    expiry_seconds,
		LoopInterval:     loop_interval,
		Fee:              fee,
		ZeroNonce:        zero_nonce,
	}
}

var (
	APP               *App = new(App)
	ATTESTER_ENDPOINT string
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Queries an attester within a TEE and validates the attestation report on the VSL.",
}

func get_env() (string, string, string, string, *ecdsa.PrivateKey, uint64, time.Duration) {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) { // Ignore if .env file does not exist
		log.Fatalf("Error loading environment variables")
	}
	ATTESTER_ENDPOINT = os.Getenv("ATTESTER_ENDPOINT")
	rpc_host := os.Getenv("VSL_HOST")
	rpc_port := os.Getenv("VSL_PORT")
	client_addr := os.Getenv("CLIENT_ADDR")
	privStr := os.Getenv("CLIENT_PRIV")
	verifier_addr := os.Getenv("VERIFIER_ADDR")
	seconds1, err1 := strconv.Atoi(os.Getenv("EXPIRY_SECONDS"))
	seconds2, err2 := strconv.Atoi(os.Getenv("CLIENT_LOOP_INTERVAL"))
	if ATTESTER_ENDPOINT == "" || rpc_host == "" || rpc_port == "" || err1 != nil || err2 != nil {
		log.Fatalf("Error loading environment variables")
	}
	exp_seconds := uint64(seconds1)
	loop_interval := time.Duration(seconds2)

	var client_priv *ecdsa.PrivateKey
	if privStr != "" {
		privBytes, err := hexutil.Decode(privStr)
		if err != nil {
			log.Fatalf("Error loading private key")
		}
		client_priv, err = crypto.ToECDSA(privBytes)
		if err != nil {
			log.Fatalf("Error loading private key")
		}
	}

	return rpc_host,
		rpc_port,
		verifier_addr,
		client_addr,
		client_priv,
		exp_seconds,
		loop_interval
}

func init() {
	cmd.RootCmd.AddCommand(clientCmd)
}
