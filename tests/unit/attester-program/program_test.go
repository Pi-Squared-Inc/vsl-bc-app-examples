package main

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestAttesterProgram(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..")

	// 1. Inference tests
	inferenceDir := filepath.Join(repoRoot, "example/common/attester/inference/src")

	tests := []struct {
		name string
		cmd  []string
	}{
		{
			name: "Image classification",
			cmd:  []string{"python3", "inference.py", "models/pi2resnetmodel.pt", "sample/goldfish.jpeg"},
		},
		{
			name: "LLM inference",
			cmd:  []string{"python3", "inference.py", "models/pi2prunedLaMa.gguf", "sample/prompt.txt"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(tc.cmd[0], tc.cmd[1:]...)
			cmd.Dir = inferenceDir
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Failed %s: %v\nOutput:\n%s", tc.name, err, string(output))
			}
		})
	}

}
