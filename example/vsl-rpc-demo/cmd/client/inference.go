package client

import (
	types "base-tee/pkg/abstract_types"
	"encoding/base64"
	"log"
	"math/big"
	"os"

	"github.com/spf13/cobra"
)

var (
	ImagePath   string
	InputPrompt string
)

var imgClassCmd = &cobra.Command{
	Use:   "img_class",
	Short: "Requests an image classification of an input image on a TEE, and attests it.",
	Run: func(cmd *cobra.Command, args []string) {
		rpc_host, rpc_port, verifier_addr, client_addr, client_priv, exp_seconds, loop_interval := get_env()
		bigIntFee := new(big.Int).SetUint64(Fee)
		APP = NewApp(
			rpc_host,
			rpc_port,
			verifier_addr,
			client_addr,
			client_priv,
			exp_seconds,
			loop_interval,
			bigIntFee,
			Zero_Nonce,
		)
		computationInput := make([]string, 1)
		computationInput[0] = imgToB64()
		err := PerformRelyingPartyCLI(APP, ATTESTER_ENDPOINT, types.InferImageClass, computationInput)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var llamaCmd = &cobra.Command{
	Use:   "llama",
	Short: "Prompts a LLaMa model running on a TEE, and attests it.",
	Run: func(cmd *cobra.Command, args []string) {
		rpc_host, rpc_port, verifier_addr, client_addr, client_priv, exp_seconds, loop_interval := get_env()
		bigIntFee := new(big.Int).SetUint64(Fee)
		APP = NewApp(
			rpc_host,
			rpc_port,
			verifier_addr,
			client_addr,
			client_priv,
			exp_seconds,
			loop_interval,
			bigIntFee,
			Zero_Nonce,
		)
		computationInput := make([]string, 1)
		computationInput[0] = base64.StdEncoding.EncodeToString([]byte(InputPrompt))
		err := PerformRelyingPartyCLI(APP, ATTESTER_ENDPOINT, types.InferTextGen, computationInput)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func imgToB64() string {
	imgBytes, err := os.ReadFile(ImagePath)
	if err != nil {
		log.Fatalf("failed reading file: %v", err)
	}
	return base64.StdEncoding.EncodeToString(imgBytes)
}

func init() {
	imgClassCmd.Flags().StringVar(&ImagePath, "img", "", "Image path to classify")
	imgClassCmd.MarkFlagRequired("img")
	imgClassCmd.Args = cobra.NoArgs
	imgClassCmd.Flags().BoolVar(&Zero_Nonce, "zero-nonce", false, "Sets nonce to zero when requesting attestation report. Useful for testing.")
	imgClassCmd.Flags().Uint64Var(&Fee, "fee", 1*(1e18), "Fee promised for claim verification (in atto-VSL).")
	clientCmd.AddCommand(imgClassCmd)

	llamaCmd.Flags().StringVar(&InputPrompt, "prompt", "", "Input prompt")
	llamaCmd.MarkFlagRequired("prompt")
	llamaCmd.Args = cobra.NoArgs
	llamaCmd.Flags().BoolVar(&Zero_Nonce, "zero-nonce", false, "Sets nonce to zero when requesting attestation report. Useful for testing.")
	llamaCmd.Flags().Uint64Var(&Fee, "fee", 1*(1e18), "Fee promised for claim verification (in atto-VSL).")
	clientCmd.AddCommand(llamaCmd)
}
