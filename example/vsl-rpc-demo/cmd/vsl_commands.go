package cmd

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	vsl "vsl-rpc-demo/vsl-wrapper"
)

var (
	VSL_HOST  string
	VSL_PORT  string
	BANK_PRIV *ecdsa.PrivateKey
	BANK_ADDR string
)

func PerformGenAddress(party string, bank_priv *ecdsa.PrivateKey, bank_addr string) (string, error) {
	log.Printf("Creating %s account on VSL...\n", party)

	vsl, err := vsl.DialVSL(VSL_HOST, VSL_PORT)
	if err != nil {
		return "", fmt.Errorf("failed VSL connection: %v", err)
	}

	addr, privKey, err := vsl.NewLoadedAccount(bank_priv, bank_addr)
	if err != nil {
		return "", fmt.Errorf("failed creating account: %s", err)
	}
	priv_key_str := hexutil.Encode(crypto.FromECDSA(privKey))
	err = set_env(party, addr, priv_key_str)
	if err != nil {
		return "", fmt.Errorf("failed creating account: %s", err)
	}
	return addr, nil
}

func PerformCheckBalance(addr string) (*big.Int, error) {
	log.Printf("Checking balance for %s on VSL...\n", addr)

	vsl, err := vsl.DialVSL(VSL_HOST, VSL_PORT)
	if err != nil {
		return big.NewInt(0), fmt.Errorf("failed VSL connection: %v", err)
	}

	balance, err := vsl.GetBalance(addr)
	if err != nil {
		return big.NewInt(0), fmt.Errorf("failed checking balance: %s", err)
	}

	log.Println("Balance (in attos):", balance)
	return balance, nil
}

func PerformFundBalance(addr string, amount *big.Int, bank_priv *ecdsa.PrivateKey, bank_addr string) error {
	log.Printf("Sending %d tokens from bank to %s on VSL...\n", amount, addr)

	vsl, err := vsl.DialVSL(VSL_HOST, VSL_PORT)
	if err != nil {
		return fmt.Errorf("failed VSL connection: %v", err)
	}

	if err := vsl.FundBalance(addr, amount, bank_priv, bank_addr); err != nil {
		return fmt.Errorf("failed funding more balance: %s", err)
	}
	log.Printf("Successfully sent %d VSL tokens from bank to %s\n", amount, addr)

	return nil
}

func get_env() {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) { // Ignore if .env file does not exist
		log.Fatalf("Error loading environment variables")
	}
	VSL_HOST = os.Getenv("VSL_HOST")
	VSL_PORT = os.Getenv("VSL_PORT")
	if VSL_HOST == "" || VSL_PORT == "" || err != nil {
		log.Fatalf("Error loading environment variables")
	}

	BANK_ADDR = os.Getenv("BANK_ADDR")
	bankPrivStr := os.Getenv("BANK_PRIV")
	bankPrivBytes, err := hexutil.Decode(bankPrivStr)
	if err != nil {
		log.Fatalf("Error loading bank private key")
	}
	BANK_PRIV, err = crypto.ToECDSA(bankPrivBytes)
	if err != nil {
		log.Fatalf("Error loading bank private key")
	}

}

func set_env(party string, addr string, privKey string) error {
	// Store account address in .env
	env, err := godotenv.Read()
	if err != nil && !os.IsNotExist(err) { // Ignore if .env file does not exist
		return fmt.Errorf("error loading environment variables")
	}

	if party == "client" {
		env["CLIENT_ADDR"] = addr
		env["CLIENT_PRIV"] = privKey
	} else {
		env["VERIFIER_ADDR"] = addr
		env["VERIFIER_PRIV"] = privKey
	}
	err = godotenv.Write(env, ".env")
	if err != nil {
		return fmt.Errorf("error writing environment variables")
	}
	return nil
}

var accCmd = &cobra.Command{
	Use:       "gen-address",
	Short:     "Generates a VSL address and private key, preloads it with 1000 tokens, and stores results in the environment file.",
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: []string{"client", "verifier"},
	Run: func(cmd *cobra.Command, args []string) {
		get_env()
		_, err := PerformGenAddress(args[0], BANK_PRIV, BANK_ADDR)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var balanceCmd = &cobra.Command{
	Use:   "check-balance",
	Short: "Checks balance for a particular account on the VSL. Amount is returned in atto-VSL (= 10^-18 VSL).",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		get_env()
		_, err := PerformCheckBalance(args[0])
		if err != nil {
			log.Fatal(err)
		}
	},
}

var fundBalanceCmd = &cobra.Command{
	Use:   "fund-balance [address] [amount]",
	Short: "Funds balance for a particular account on the VSL. Amount is specified in atto-VSL (= 10^-18 VSL).",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		get_env()
		addr := args[0]
		amountAttos, ok := new(big.Int).SetString(args[1], 10)
		if !ok {
			log.Fatalf("failed parsing ammount")
		}
		err := PerformFundBalance(addr, amountAttos, BANK_PRIV, BANK_ADDR)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(accCmd)
	RootCmd.AddCommand(balanceCmd)
	RootCmd.AddCommand(fundBalanceCmd)
}
