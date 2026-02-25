package protocol

import (
	"encoding/json"

	"github.com/google/uuid"
)

// Version is the protocol version.
const Version = "1.0"

// MessageType represents the type of a message.
type MessageType string

const (
	// MessageTypeCall is a request to call a method.
	MessageTypeCall MessageType = "call"
	// MessageTypeResponse is a response to a call.
	MessageTypeResponse MessageType = "response"
	// MessageTypeError is an error response.
	MessageTypeError MessageType = "error"
	// MessageTypeEvent is an asynchronous event.
	MessageTypeEvent MessageType = "event"
)

// Request represents a protocol request message.
type Request struct {
	Version string      `json:"version"` // Protocol version
	ID      string      `json:"id"`      // Request ID
	Type    MessageType `json:"type"`    // Message type
	Method  string      `json:"method"`  // Method name
	Params  any         `json:"params"`  // Method parameters
}

// Response represents a protocol response message.
type Response struct {
	Version string      `json:"version"` // Protocol version
	ID      string      `json:"id"`      // Request ID (correlated)
	Type    MessageType `json:"type"`    // Message type
	Result  any         `json:"result"`  // Result value (success)
	Error   *Error      `json:"error"`   // Error value (failure)
}

// Error represents a protocol error.
type Error struct {
	Code    int    `json:"code"`    // Error code
	Message string `json:"message"` // Error message
	Data    any    `json:"data"`    // Additional error data
}

// NewRequest creates a new request with auto-generated ID.
func NewRequest(method string, params any) *Request {
	return &Request{
		Version: Version,
		ID:      uuid.New().String(),
		Type:    MessageTypeCall,
		Method:  method,
		Params:  params,
	}
}

// NewResponse creates a new response for a request.
func NewResponse(id string, result any) *Response {
	return &Response{
		Version: Version,
		ID:      id,
		Type:    MessageTypeResponse,
		Result:  result,
	}
}

// NewErrorResponse creates a new error response for a request.
func NewErrorResponse(id string, code int, message string) *Response {
	return &Response{
		Version: Version,
		ID:      id,
		Type:    MessageTypeError,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}
}

// NewErrorWithData creates a new error response with additional data.
func NewErrorWithData(id string, code int, message string, data any) *Response {
	return &Response{
		Version: Version,
		ID:      id,
		Type:    MessageTypeError,
		Error: &Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// Encode encodes a request to JSON bytes.
func (r *Request) Encode() ([]byte, error) {
	return json.Marshal(r)
}

// DecodeRequest decodes JSON bytes to a request.
func DecodeRequest(data []byte) (*Request, error) {
	var req Request
	err := json.Unmarshal(data, &req)
	return &req, err
}

// Encode encodes a response to JSON bytes.
func (r *Response) Encode() ([]byte, error) {
	return json.Marshal(r)
}

// DecodeResponse decodes JSON bytes to a response.
func DecodeResponse(data []byte) (*Response, error) {
	var resp Response
	err := json.Unmarshal(data, &resp)
	return &resp, err
}

// IsSuccess returns true if the response is successful (no error).
func (r *Response) IsSuccess() bool {
	return r.Error == nil
}

// IsError returns true if the response is an error.
func (r *Response) IsError() bool {
	return r.Error != nil
}
