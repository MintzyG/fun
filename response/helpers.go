package response

import (
	"encoding/json"
	"log"
)

func validateStatusCode(code int) error {
	if code < 100 || code > 599 {
		return &StatusCodeError{
			Code: code,
		}
	}
	return nil
}

type sizeEstimator struct {
	size int
}

func (s *sizeEstimator) Write(p []byte) (n int, err error) {
	s.size += len(p)
	return len(p), nil
}

func (r *Response) estimateSize() (int, error) {
	if r == nil {
		log.Println("WARNING: estimateSize called on nil Response")
		return 0, nil
	}
	estimator := &sizeEstimator{}
	encoder := json.NewEncoder(estimator)
	if err := encoder.Encode(r); err != nil {
		return 0, err
	}
	return estimator.size, nil
}

// validateResponseSize checks if the response size is within limits
func (r *Response) validateResponseSize() error {
	if r == nil {
		log.Println("WARNING: validateResponseSize called on nil Response")
		return nil
	}
	config := r.getResponseConfig()
	if !config.EnableSizeValidation {
		return nil
	}

	size, err := r.estimateSize()
	if err != nil {
		return &EncodingError{Inner: err}
	}

	if size > config.ResponseSizeLimit {
		return &SizeLimitError{
			Size: size,
			Max:  config.ResponseSizeLimit,
		}
	}

	return nil
}

func (r *Response) GetResponseStats() map[string]any {
	if r == nil {
		log.Println("WARNING: GetResponseStats called on nil Response")
		return nil
	}
	data, _ := json.Marshal(r)
	return map[string]any{
		"size_bytes":   len(data),
		"trace_count":  len(r.Trace),
		"content_type": r.ContentType,
		"status_code":  r.Code,
	}
}

func (r *Response) IsWithinLimits() bool {
	if r == nil {
		log.Println("WARNING: IsWithinLimits called on nil Response")
		return true
	}
	config := r.getResponseConfig()

	if len(r.Trace) > config.MaxTraceSize {
		return false
	}

	if config.EnableSizeValidation {
		if err := r.validateResponseSize(); err != nil {
			return false
		}
	}

	return true
}
