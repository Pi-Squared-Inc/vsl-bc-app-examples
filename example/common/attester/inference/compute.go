package inference

import (
	utils "attester/utils"
	types "base-tee/pkg/abstract_types"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var (
	_, currentFile, _, _ = runtime.Caller(0)
	inference_src        = filepath.Dir(currentFile) + "/src/"
	inference_program    = inference_src + "inference.py"
	models_dir           = inference_src + "models/"
	img_class_model      = models_dir + "pi2resnetmodel.pt"
	llama_model          = models_dir + "pi2prunedLaMa.gguf"
)

func Inference(tpm io.ReadWriteCloser, task types.InferenceTask, inputB64 string) (string, error) {
	log.Println("Inference: Saving input bytes to temporary file...")
	tmp_file, err := os.CreateTemp(models_dir, "task*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmp_file.Name())

	input_bytes, err := base64.StdEncoding.DecodeString(inputB64)
	if err != nil {
		return "", fmt.Errorf("failed to decode b64 input: %w", err)
	}
	_, err = tmp_file.Write(input_bytes)
	if err != nil {
		return "", fmt.Errorf("failed to save input to temporary file: %w", err)
	}

	pythonOk := make(chan bool)
	ioDone := make(chan bool)
	var inference_out []byte
	go func() {
		log.Println("Inference: Executing program...")
		input_path := tmp_file.Name()
		switch task {
		case types.ImageClass:
			log.Println("Inference: Task is image classification")
			inference_out, err = exec.Command("python3", inference_program, img_class_model, input_path).Output()
		case types.TextGen:
			log.Println("Inference: Task is text generation")
			inference_out, err = exec.Command("python3", inference_program, llama_model, input_path).Output()
		default:
			pythonOk <- false
			return
		}
		if err != nil {
			pythonOk <- false
			return
		}
		pythonOk <- true
	}()
	go func() {
		log.Println("Inference: Resetting PCR 23 old value...")
		utils.ResetPCR(tpm)
		log.Println("Inference: Extending PCR 23 with current measurements...")
		// A hash of the resnet model
		log.Println("Logging AI model...")
		if task == types.ImageClass {
			utils.ExtendPCRFileHash(tpm, img_class_model)
		} else if task == types.TextGen {
			utils.ExtendPCRFileHash(tpm, llama_model)
		}

		// A hash of the input
		log.Println("Logging input...")
		utils.ExtendPCR(tpm, []byte(inputB64))
		ioDone <- true
	}()

	<-ioDone
	ok := <-pythonOk
	if !ok {
		return "", fmt.Errorf("inference program failed")
	}

	// A hash of the output
	log.Println("Logging output...")
	utils.ExtendPCR(tpm, inference_out)
	log.Println("Inference: Done.")

	return string(inference_out), nil
}
