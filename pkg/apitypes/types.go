package apitypes

// OkResponse represents a successful response with a message
type OkResponse struct {
	Response string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
