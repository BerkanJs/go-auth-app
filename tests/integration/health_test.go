package integration_test

import (
	"net/http"
	"testing"
)

func TestHealthEndpoint_Returns200(t *testing.T) {
	resp, err := http.Get(testServer.URL + "/api/health")
	if err != nil {
		t.Fatalf("istek gönderilemedi: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("200 beklendi, alınan: %d", resp.StatusCode)
	}
}
