package backendserver

import (
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"vsl-rpc-demo/cmd"
	"vsl-rpc-demo/cmd/backend-server/api"
	"vsl-rpc-demo/cmd/backend-server/models"
	"vsl-rpc-demo/cmd/client"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

const VERIFIER_FEE int64 = 15

var backendCmd = &cobra.Command{
	Use:   "backend-server",
	Short: "Run the backend for the TEE computation frontend website.",
	Run: func(cmd *cobra.Command, args []string) {
		// Load the backend's .env:
		err := godotenv.Load("cmd/backend-server/.env")
		if err != nil && !os.IsNotExist(err) { // Ignore if .env file does not exist
			log.Fatalf("Error loading environment variables")
		}
		// Load the general .env
		err = godotenv.Load()
		if err != nil && !os.IsNotExist(err) { // Ignore if .env file does not exist
			log.Fatalf("Error loading environment variables")
		}
		// Set variables
		port := os.Getenv("BACKEND_SERVER_PORT")
		attesterEndpointsStr := os.Getenv("ATTESTER_ENDPOINTS")
		vslHost := os.Getenv("VSL_HOST")
		vslPort := os.Getenv("VSL_PORT")
		verifierAddress := os.Getenv("VERIFIER_ADDR")
		seconds1, err1 := strconv.Atoi(os.Getenv("EXPIRY_SECONDS"))
		seconds2, err2 := strconv.Atoi(os.Getenv("CLIENT_LOOP_INTERVAL"))
		if err1 != nil || err2 != nil {
			log.Fatalf("Error loading EXPIRY_SECONDS and/or CLIENT_LOOP_INTERVAL from .env")
		}
		expirySeconds := uint64(seconds1)
		loopInterval := time.Duration(seconds2)
		bankAddress := os.Getenv("BANK_ADDR")
		bankPrivStr := os.Getenv("BANK_PRIV")
		bankPrivBytes, err := hexutil.Decode(bankPrivStr)
		if err != nil {
			log.Fatalf("Error loading bank private key")
		}
		bankPriv, err := crypto.ToECDSA(bankPrivBytes)
		if err != nil {
			log.Fatalf("Error loading bank private key")
		}

		fee := new(big.Int).Mul(big.NewInt(int64(VERIFIER_FEE)), big.NewInt(1e18))

		lb := models.NewBalancer(strings.Split(attesterEndpointsStr, ","), &http.Client{Timeout: 10 * time.Second})
		app := models.NewApp(vslHost, vslPort, verifierAddress, bankPriv, bankAddress, lb)
		// Initialize the relying party client App
		clientApp := client.NewApp(
			vslHost,
			vslPort,
			app.VerifierAddress,
			app.ClientAddress,
			app.ClientPrivateKey,
			expirySeconds,
			loopInterval,
			fee,
			false, // zero_nonce
		)

		defer func() {
			app.WorkerPool1.Release()
			app.WorkerPool2.Release()
			app.VSLClient.Close()
			clientApp.VSLClient.Close()
			app.LoadBalancer.Close()
			app.API.Shutdown()
		}()

		// Start the API server
		api.RegisterClaimAPI(app)
		api.RegisterVerificationRecordAPI(app, clientApp)
		go func() {
			err := app.API.Listen(":" + port)
			if err != nil {
				log.Fatal(err)
			}
		}()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
	},
}

func init() {
	cmd.RootCmd.AddCommand(backendCmd)
}
