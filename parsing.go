package fun

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ExtractData parses an http.Response into your Response struct and also
// unmarshals the Data field into the provided target model.
// target must be a pointer (to struct or slice).
func ExtractData(httpResp *http.Response, target any) (*Response, error) {
	if httpResp == nil {
		return nil, fmt.Errorf("http response is nil")
	}
	defer httpResp.Body.Close()

	// Read the body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal into wrapper Response
	var r Response
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into Response: %w", err)
	}

	// If no Data, return as is
	if r.Data == nil {
		return &r, nil
	}

	// Marshal Data back to JSON
	dataBytes, err := json.Marshal(r.Data)
	if err != nil {
		return &r, fmt.Errorf("failed to marshal Response.Data: %w", err)
	}

	// Unmarshal into target
	if err := json.Unmarshal(dataBytes, target); err != nil {
		return &r, fmt.Errorf("failed to unmarshal Response.Data into target: %w", err)
	}

	return &r, nil
}
