package test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MintzyG/FastUtilitiesNet/response"
)

func TestInvalidStatusCodeHandling(t *testing.T) {
	// Capture log output
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(nil) // Note: this might affect other tests if run in parallel, 
	                        // but for a simple case it works.

	t.Run("WithCode invalid code", func(t *testing.T) {
		logBuf.Reset()
		w := httptest.NewRecorder()
		resp := response.Base().WithCode(99)
		
		// Check if warning was logged during WithCode
		if !strings.Contains(logBuf.String(), "WARNING: Invalid status code 99 set") {
			t.Errorf("Expected warning log for invalid status code in WithCode, got: %s", logBuf.String())
		}
		
		logBuf.Reset()
		resp.Send(w)
		
		// Check if warning was logged during Send and it was sent as 500
		if !strings.Contains(logBuf.String(), "Response will be sent. as 500") {
			t.Errorf("Expected warning log for invalid status code in Send, got: %s", logBuf.String())
		}
		
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code 500, got %d", w.Code)
		}
	})

	t.Run("OK helper with custom invalid code", func(t *testing.T) {
		logBuf.Reset()
		w := httptest.NewRecorder()
		
		// Directly setting Code to something invalid (since it's exported)
		resp := response.OK()
		resp.Code = 600 
		
		resp.Send(w)
		
		if !strings.Contains(logBuf.String(), "Response will be sent. as 500") {
			t.Errorf("Expected warning log for invalid status code in Send, got: %s", logBuf.String())
		}
		
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code 500, got %d", w.Code)
		}
	})

	t.Run("Base uninitialized code 0", func(t *testing.T) {
		logBuf.Reset()
		w := httptest.NewRecorder()
		resp := response.Base()
		
		resp.Send(w)
		
		if !strings.Contains(logBuf.String(), "Response will be sent. as 500") {
			t.Errorf("Expected warning log for invalid status code 0, got: %s", logBuf.String())
		}
		
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code 500, got %d", w.Code)
		}
	})
}
