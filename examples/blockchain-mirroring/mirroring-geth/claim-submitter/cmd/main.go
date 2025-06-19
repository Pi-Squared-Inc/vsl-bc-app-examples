package main

import (
	"log"
	"mirroring-geth-claim-submitter/models"
	"mirroring-geth-claim-submitter/utils"
)

func main() {
	app, err := models.NewApp()
	if err != nil {
		log.Fatalf("Error initializing app: %+v", err)
	}
	err = utils.ObserveBlocks(app)
	if err != nil {
		log.Fatalf("Error observing blocks: %+v", err)
	}

}
