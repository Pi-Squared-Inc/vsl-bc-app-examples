package main

import (
	"vsl-rpc-demo/cmd"
	_ "vsl-rpc-demo/cmd/backend-server"
	_ "vsl-rpc-demo/cmd/client"
	_ "vsl-rpc-demo/cmd/verifier"
)

func main() {
	cmd.Execute()
}
