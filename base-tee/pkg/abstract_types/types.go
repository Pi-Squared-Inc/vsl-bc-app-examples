package types

type ClaimType string

const (
	// A claim that a given attestation report is coming from a legitimate TEE
	ClaimTypeTEEComputation ClaimType = "TEEComputation"
)

/**
 * TEEComputation claim and its verification context
 */
type TEEComputationClaim struct {
	ClaimType     ClaimType           `json:"type"`
	Computation   Computation         `json:"computation"`
	Input         []string            `json:"input"`
	Result        string              `json:"result"`
	DigestHistory map[string][]string `json:"digest_history"`
	Nonce         []byte              `json:"nonce"`
}

// VerificationContext for TEEComputation
type TEEComputationClaimVerificationContext struct {
	Attestation []byte `json:"report"`
	// ^ As byte array; must be unmarshaled into an *attest.Attestation.
	// The reason this is not directly an *attest.Attestation for now is that
	// *attest.Attestation can't be marshaled to JSON, but only to protobuf.
}

// VSL data types:

type Timestamp struct {
	Seconds     uint64 `json:"seconds"`
	Nanoseconds uint32 `json:"nanos"`
}

func (t *Timestamp) Tick() Timestamp {
	if t.Nanoseconds == 999_999_999 {
		return Timestamp{
			Seconds:     t.Seconds + 1,
			Nanoseconds: 0,
		}
	} else {
		return Timestamp{
			Seconds:     t.Seconds,
			Nanoseconds: t.Nanoseconds + 1,
		}
	}
}

func MaxT(t1 Timestamp, t2 Timestamp) Timestamp {
	if t1.Seconds > t2.Seconds {
		return t1
	} else if t1.Seconds == t2.Seconds {
		if t1.Nanoseconds > t2.Nanoseconds {
			return t1
		}
		return t2
	}
	return t2
}

type SubmittedClaim struct {
	Claim     string    `json:"claim"`      /// the claim to be verified
	ClaimType string    `json:"claim_type"` /// the claim type
	Proof     string    `json:"proof"`      /// the proof of the claim
	Nonce     string    `json:"nonce"`      /// the client nonce
	Verifiers []string  `json:"to"`         /// the list of verifiers (currently a singleton list)
	Quorum    uint16    `json:"quorum"`     /// the minimum quorum of signatures (currently 1)
	Client    string    `json:"from"`       /// the client account requesting verification
	Expires   Timestamp `json:"expires"`    /// the time after which the claim is dropped if not enough verifications are received
	Fee       string    `json:"fee"`        /// the fee for verification
}

type SignedSubmittedClaim struct {
	Claim     string    `json:"claim"`      /// the claim to be verified
	ClaimType string    `json:"claim_type"` /// the claim type
	Proof     string    `json:"proof"`      /// the proof of the claim
	Nonce     string    `json:"nonce"`      /// the client nonce
	Verifiers []string  `json:"to"`         /// the list of verifiers (currently a singleton list)
	Quorum    uint16    `json:"quorum"`     /// the minimum quorum of signatures (currently 1)
	Client    string    `json:"from"`       /// the client account requesting verification
	Expires   Timestamp `json:"expires"`    /// the time after which the claim is dropped if not enough verifications are received
	Fee       string    `json:"fee"`        /// the fee for verification
	Hash      string    `json:"hash"`
	R         string    `json:"r"`
	S         string    `json:"s"`
	V         string    `json:"v"`
}

type TimestampedSignedSubmittedClaim struct {
	Data      SignedSubmittedClaim `json:"data"`
	ID        string               `json:"id"`
	Timestamp Timestamp            `json:"timestamp"`
}
type VerifiedClaim struct {
	Claim  string `json:"claim"` /// the verified claim
	Nonce  string `json:"nonce"` /// the client nonc
	Client string `json:"to"`    /// the client interested in this claim
}

type SignedSettleClaimMessage struct {
	Verifier string `json:"from"`            /// The address of the verifier requesting claim settlement
	Nonce    string `json:"nonce"`           /// The nonce of the verifier requesting claim settlement
	ClaimID  string `json:"target_claim_id"` /// The id of the claim for which claim settlement is requested
	Hash     string `json:"hash"`
	R        string `json:"r"`
	S        string `json:"s"`
	V        string `json:"v"`
}

type SettleClaimMessage struct {
	Verifier string `json:"from"`            /// The address of the verifier requesting claim settlement
	Nonce    string `json:"nonce"`           /// The nonce of the verifier requesting claim settlement
	ClaimID  string `json:"target_claim_id"` /// The id of the claim for which claim settlement is requested
}

type SettledVerifiedClaim struct {
	VerifiedClaim VerifiedClaim `json:"verified_claim"` /// the claim which was verified
	Verifiers     []string      `json:"verifiers"`      /// the addresses of the verifiers which have signed the `verified_claim` object
}

type SignedSettledVerifiedClaim struct {
	VerifiedClaim VerifiedClaim `json:"verified_claim"` /// the claim which was verified
	Verifiers     []string      `json:"verifiers"`      /// the addresses of the verifiers which have signed the `verified_claim` object
	Hash          string        `json:"hash"`
	R             string        `json:"r"`
	S             string        `json:"s"`
	V             string        `json:"v"`
}

type TimestampedSignedSettledVerifiedClaim struct {
	Data      SignedSettledVerifiedClaim `json:"data"`
	ID        string                     `json:"id"`
	Timestamp Timestamp                  `json:"timestamp"`
}

type CreateAccountMessage struct {
	OwnerAddress string `json:"owner_address"`
	Script       string `json:"script"`
	Label        string `json:"label"`
}

type SignedCreateAccountMessage struct {
	OwnerAddress string `json:"owner_address"`
	Script       string `json:"script"`
	Label        string `json:"label"`
	Hash         string `json:"hash"`
	R            string `json:"r"`
	S            string `json:"s"`
	V            string `json:"v"`
}

type PayMessage struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
	Nonce  string `json:"nonce"`
}

type SignedPayMessage struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
	Nonce  string `json:"nonce"`
	Hash   string `json:"hash"`
	R      string `json:"r"`
	S      string `json:"s"`
	V      string `json:"v"`
}

type PaymentClaim struct {
	Payment PayMessage `json:"Payment"`
}

// Sum type for TEE computation kinds
type Computation string

const (
	InferImageClass      Computation = "img_class"
	InferTextGen         Computation = "text_gen"
	BlockProcessingKReth Computation = "block_processing_kreth"
)

// AI-Inference specific types
type InferenceTask int

const (
	ImageClass InferenceTask = iota
	TextGen
)

var TaskToName = map[InferenceTask]string{
	ImageClass: "image-classification",
	TextGen:    "text-generation",
}

var NameToTask = map[string]InferenceTask{
	"image-classification": ImageClass,
	"text-generation":      TextGen,
}
