package api

// swagger:ignore
type SignRequest struct {
	MessageHashHex string `json:"message_hash_hex" validate:"required"`
}

// swagger:ignore
type SignResponse struct {
	SignatureHex string `json:"signature_hex"`
	SignerID     string `json:"signer_id"`
}

// swagger:ignore
type VerifyRequest struct {
	MessageHashHex string `json:"message_hash_hex" validate:"required"`
	SignatureHex   string `json:"signature_hex" validate:"required"`
	SignerID       string `json:"signer_id" validate:"required"`
}
