package models

import (
	types "base-tee/pkg/abstract_types"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"
	utils "vsl-rpc-demo/utils"
	vsl "vsl-rpc-demo/vsl-wrapper"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/panjf2000/ants/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"vsl-rpc-demo/cmd/client"
)

type App struct {
	API              *fiber.App
	DB               *gorm.DB
	VSLClient        *vsl.VSL
	WorkerPool1      *ants.PoolWithFuncGeneric[SendToAttesterPayload]
	WorkerPool2      *ants.PoolWithFuncGeneric[SendToVSLPayload]
	VerifierAddress  string
	ClientAddress    string
	ClientPrivateKey *ecdsa.PrivateKey
	LoadBalancer     *LoadBalancer
}

var VALID_PAYMENT_AMOUNT *big.Int = new(big.Int).Mul(big.NewInt(20), big.NewInt(1e18))

func NewApp(vslHost string, vslPort string, verifierAddress string, bankPriv *ecdsa.PrivateKey, bankAddress string, balancer *LoadBalancer) *App {
	db, err := gorm.Open(sqlite.Open("db.sqlite"), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&ValidationRecord{})
	db.AutoMigrate(&VerificationRecord{})
	db.AutoMigrate(&UserPaymentRecord{})

	apiServer := fiber.New(fiber.Config{
		BodyLimit: 100 * 1024 * 1024, // 100MB
	})

	// CORS
	apiServer.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	pool1, err := ants.NewPoolWithFuncGeneric(100, SendToAttester, ants.WithPreAlloc(true), ants.WithNonblocking(true))
	if err != nil {
		panic("failed to create worker pool 1")
	}

	pool2, err := ants.NewPoolWithFuncGeneric(500, SendToVSL, ants.WithPreAlloc(true), ants.WithNonblocking(true))
	if err != nil {
		panic("failed to create worker pool 2")
	}

	vslClient, err := vsl.DialVSL(vslHost, vslPort)
	if err != nil {
		panic("failed to connect to VSL server")
	}

	clientAddress, clientPrivateKey, err := vslClient.NewLoadedAccount(bankPriv, bankAddress)
	if err != nil {
		panic("failed to create VSL account")
	}

	return &App{
		API:              apiServer,
		DB:               db,
		VSLClient:        vslClient,
		WorkerPool1:      pool1,
		WorkerPool2:      pool2,
		VerifierAddress:  verifierAddress,
		ClientAddress:    clientAddress,
		ClientPrivateKey: clientPrivateKey,
		LoadBalancer:     balancer,
	}
}

type SignedComputeRequest struct {
	Computation    string
	Input          []string
	SenderAddress  string
	PaymentClaimId string
	Hash           string
	R              string
	S              string
	V              string
}

type SendToAttesterPayload struct {
	App         *App
	ClientApp   *client.App
	Attester    *AttesterEndpoint
	ID          string
	Request     SignedComputeRequest
	InputSHA256 string // used for verifying the signature
	Retried     int
	MaxRetries  int
}

func SendToAttester(payload SendToAttesterPayload) {
	// First verify payment. Skip verification if this is a retry (it's been verified already)
	if payload.Retried == 0 {
		err := verifyUserPayment(payload.App, payload.Request, payload.InputSHA256)
		if err != nil {
			dbRecordFailure(payload.App, payload.ID, err.Error())
			return
		}
	}
	// Then proceed to calling the attester (retry if it errors out)
	claim, proof, err := client.AttestAndGenClaim(payload.ClientApp, payload.Attester.Url, types.Computation(payload.Request.Computation), payload.Request.Input)
	if err != nil {
		if payload.Retried < payload.MaxRetries {
			log.Printf("Retrying request %s with a different attester...\n", payload.ID)
			new_attester, err := payload.App.LoadBalancer.GetNextAttester()
			if err != nil {
				log.Printf("Cannot retry!")
				dbRecordFailure(payload.App, payload.ID, err.Error())
				return
			}
			payload.Attester = new_attester
			payload.Retried += 1
			SendToAttester(payload)
		} else {
			log.Println("error generating claim: ", err)
			dbRecordFailure(payload.App, payload.ID, err.Error())
			return
		}
		return
	}
	err = payload.Attester.FinishTask()
	if err != nil {
		log.Printf("Could not cleanly finish attester task: %s\n", err.Error())
	}

	resultString := claim.Result
	switch claim.Computation {
	case types.InferImageClass:
		resultString, err = extractPredictedClass(resultString)
		if err != nil {
			log.Println("image classification returned unexpected output: ", err)
			dbRecordFailure(payload.App, payload.ID, err.Error())
			return
		}
	}
	err = payload.App.DB.Model(&VerificationRecord{}).Where("id = ?", payload.ID).Updates(map[string]interface{}{
		"result": resultString,
		"status": "awaiting_settlement",
	}).Error
	if err != nil {
		log.Println("error updating verification record: ", err)
	}

	go func() {
		err := SubmitWithRetry(payload.App.WorkerPool2, SendToVSLPayload{
			App:         payload.App,
			ClientApp:   payload.ClientApp,
			ID:          payload.ID,
			Computation: payload.Request.Computation,
			Claim:       claim,
			Proof:       proof,
		}, 30, 10*time.Second)
		if err != nil {
			log.Printf("Failed to submit SendToVSL to worker pool: %v", err)
			dbRecordFailure(payload.App, payload.ID, err.Error())
		}
	}()
}

type SendToVSLPayload struct {
	App            *App
	ClientApp      *client.App
	ID             string
	Computation    string
	Claim          *types.TEEComputationClaim
	Proof          *types.TEEComputationClaimVerificationContext
	SenderAddress  string
	PaymentClaimId string
}

func SendToVSL(payload SendToVSLPayload) {
	claimId, _, err := client.VerifyClaim(payload.ClientApp, payload.Claim, payload.Proof)
	if err != nil {
		log.Printf("claim not verified: %v", err)
		dbRecordClaimIdFailure(payload.App, payload.ID, claimId, err.Error())
		return
	} else {
		log.Println("Verified!")
	}

	err = payload.App.DB.Model(&VerificationRecord{}).Where("id = ?", payload.ID).Updates(map[string]interface{}{
		"claim_id": claimId,
		"status":   "completed",
	}).Error
	if err != nil {
		log.Println("error updating verification record: ", err)
	}
}

func dbRecordFailure(app *App, request_id string, reason string) {
	err := app.DB.Model(&VerificationRecord{}).Where("id = ?", request_id).Updates(map[string]interface{}{
		"status": "failed: " + reason,
	}).Error
	if err != nil {
		log.Println("error updating verification record: ", err)
	}

}

func dbRecordClaimIdFailure(app *App, request_id string, claim_id string, reason string) {
	err := app.DB.Model(&VerificationRecord{}).Where("id = ?", request_id).Updates(map[string]interface{}{
		"claim_id": claim_id,
		"status":   "failed: " + reason,
	}).Error
	if err != nil {
		log.Println("error updating verification record: ", err)
	}
}

func verifyUserPayment(app *App, request SignedComputeRequest, inputSHA256 string) error {
	// Check in database if payment id has already been used
	var num_claims int64
	err := app.DB.Model(&UserPaymentRecord{}).Where("id = ?", request.PaymentClaimId).Count(&num_claims).Error
	if err != nil {
		return err
	}
	if num_claims != 0 {
		return fmt.Errorf("cannot use same payment_claim_id twice")
	}

	// Verify request signature
	rlpRequest, err := rlp.EncodeToBytes(
		[]interface{}{
			request.Computation, request.SenderAddress, request.PaymentClaimId, inputSHA256,
		},
	)
	if err != nil {
		return fmt.Errorf("could not rlp encode request")
	}
	err = utils.VerifySign(rlpRequest, request.SenderAddress, request.Hash, request.R, request.S, request.V)
	if err != nil {
		return fmt.Errorf("signature verification failed: %s", err.Error())
	}

	claim, err := app.VSLClient.PollSettledByID(
		request.PaymentClaimId,
		time.Now(),
		30,               // poll for at most 30s for payment claim
		time.Duration(5), // poll frequency: 5 s
	)
	if err != nil {
		return err
	}
	claimStr := claim.VerifiedClaim.Claim
	payMessage := types.PaymentClaim{}
	err = json.Unmarshal([]byte(claimStr), &payMessage)
	if err != nil {
		return fmt.Errorf("claim id provided does not correspond to a payment")
	}
	if payMessage.Payment.From != request.SenderAddress {
		return fmt.Errorf("payment not made from claimed sender address")
	}
	if payMessage.Payment.To != app.ClientAddress {
		return fmt.Errorf("payment not made to backend address")
	}
	bigIntAmount, ok := new(big.Int).SetString(payMessage.Payment.Amount, 0)
	if !ok {
		return fmt.Errorf("could not parse payment amount for claim id")
	}
	if bigIntAmount.Cmp(VALID_PAYMENT_AMOUNT) != 0 {
		return fmt.Errorf("wrong payment amount")
	}

	// Store payment id in database to prevent it from being used again:
	err = app.DB.Create(UserPaymentRecord{
		ID:          request.PaymentClaimId,
		UserAddress: request.SenderAddress,
	}).Error
	if err != nil {
		return fmt.Errorf("error creating user payment record: %s", err.Error())
	}

	return nil
}
