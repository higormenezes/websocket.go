package internal

import (
	"fmt"
	"testing"
)

func TestGetWebSocketAcceptKey(t *testing.T) {
	var tests = []struct {
		initialKey, expectedAcceptKey string
	}{
		{"dGhlIHNhbXBsZSBub25jZQ==", "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="},
		{"   dGhlIHNhbXBsZSBub25jZQ==   ", "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("for the initial key '%s' it should return the accept key '%s'", test.initialKey, test.expectedAcceptKey)
		t.Run(testName, func(t *testing.T) {
			acceptKey, err := getWebSocketAcceptKey(test.initialKey)
			if err != nil {
				t.Errorf("An error occurred while trying to get the accept key")
			}
			if acceptKey != test.expectedAcceptKey {
				t.Errorf("Expected %s, got %s", test.expectedAcceptKey, acceptKey)
			}
		})
	}
}
