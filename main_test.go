package main

import "testing"

func TestParseSpeedtestResult(t *testing.T) {
	mockOutput := `
	Speedtest by Ookla
	
	  Server: TWL-KOM - Ludwigshafen (id = 10291)
		 ISP: Vodafone Germany Cable
	 Latency:    12.42 ms   (2.62 ms jitter)
	Download:   103.41 Mbps (data used: 99.3 MB )
	  Upload:     9.41 Mbps (data used: 4.7 MB )
	Packet Loss:     0.0%
	Result URL: https://www.speedtest.net/result/c/8f37ffd1-121d-48cf-808f-dd0d11e0336f
	`
	expectedServer := "TWL-KOM - Ludwigshafen (id = 10291)"
	expectedLatency := 12.42
	expectedDownload := 103.41
	expectedUpload := 9.41
	expectedPacketLoss := 0.0

	parsedServer, parsedLatency, parsedDownload, parsedUpload, parsedPackageLoss := parseSpeedtestResult(mockOutput)
	if parsedServer != expectedServer {
		t.Errorf("Server named not parsed correctly. Expected %s, got %s", expectedServer, parsedServer)
	}
	if parsedLatency != float32(expectedLatency) {
		t.Errorf("Latency not parsed correctly. Expected %f, got %f", expectedLatency, parsedLatency)
	}
	if parsedDownload != float32(expectedDownload) {
		t.Errorf("Download not parsed correctly. Expected %f, got %f", expectedDownload, parsedDownload)
	}
	if parsedUpload != float32(expectedUpload) {
		t.Errorf("Upload not parsed correctly. Expected %f, got %f", expectedUpload, parsedUpload)
	}
	if parsedPackageLoss != float32(expectedPacketLoss) {
		t.Errorf("Package loss not parsed correctly. Expected %f, got %f", expectedPacketLoss, parsedPackageLoss)
	}
}
