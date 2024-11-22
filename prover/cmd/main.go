package main

import (
	"flag"
	"fmt"
	"os"

	"prover/circuits"

	"github.com/brevis-network/brevis-sdk/sdk/prover"
)

var port = flag.Uint("port", 33247, "the port to start the service at")

func main() {
	flag.Parse()

	proverService, err := prover.NewService(&circuits.MellowRewardsCircuit{}, prover.ServiceConfig{
		SetupDir: "$HOME/circuitOut",
		SrsDir:   "$HOME/kzgsrs",
		RpcURL:   "https://mainnet.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161",
		ChainId:  1,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	proverService.Serve("", *port)
}
