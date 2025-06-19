package models

type AttesterResponse struct {
	Result      string `json:"result"`
	Attestation []byte `json:"report"`
}

type AttesterQuery struct {
	ClaimType   string   `json:"type"`
	Computation string   `json:"computation"`
	Input       []string `json:"input"`
	Nonce       []byte   `json:"nonce"`
}

type AttesterResult struct {
	PredictedClass string `json:"predicted_class"`
	ModelOutput    string `json:"model_output"`
}
