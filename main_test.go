package main

import "testing"

func TestParseSpeedtestResultSuccess(t *testing.T) {
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
	expectedLatency := float32(12.42)
	expectedDownload := float32(103.41)
	expectedUpload := float32(9.41)
	expectedPacketLoss := float32(0.0)

	measurement := parseSpeedtestResult(mockOutput)
	verifyParseResult(t, &measurement, &expectedServer, &expectedLatency, &expectedDownload, &expectedUpload, &expectedPacketLoss)
}

func TestParseSpeedtestResultPacketLossUnavailable(t *testing.T) {
	mockOutput := `
	Speedtest by Ookla

	  Server: PVDataNet - Frankfurt (id = 40094)
		 ISP: Plusnet
	 Latency:     5.47 ms   (0.12 ms jitter)
	Download:   102.90 Mbps (data used: 47.5 MB)
	  Upload:    52.63 Mbps (data used: 83.5 MB)
 Packet Loss: Not available.
  Result URL: https://www.speedtest.net/result/c/5dcb37c2-7e47-4780-aecc-1b40ef510d95
	`
	expectedServer := "PVDataNet - Frankfurt (id = 40094)"
	expectedLatency := float32(5.47)
	expectedDownload := float32(102.90)
	expectedUpload := float32(52.63)
	var expectedPacketLoss *float32

	measurement := parseSpeedtestResult(mockOutput)
	verifyParseResult(t, &measurement, &expectedServer, &expectedLatency, &expectedDownload, &expectedUpload, expectedPacketLoss)
}

func TestParseSpeedtestResultInvalidCommand(t *testing.T) {
	mockOutput := `an error occurred`
	var expectedServer *string
	var expectedLatency, expectedDownload, expectedUpload, expectedPacketLoss *float32

	measurement := parseSpeedtestResult(mockOutput)
	verifyParseResult(t, &measurement, expectedServer, expectedLatency, expectedDownload, expectedUpload, expectedPacketLoss)
}

func TestParseSpeedtestResultDownloadUnavailable(t *testing.T) {
	mockOutput := `
	Speedtest by Ookla
	
	  Server: TWL-KOM - Ludwigshafen (id = 10291)
		 ISP: Vodafone Germany Cable
	 Latency:    12.42 ms   (2.62 ms jitter)
	Download:   not available
	  Upload:     9.41 Mbps (data used: 4.7 MB )
	Packet Loss:     0.0%
	Result URL: https://www.speedtest.net/result/c/8f37ffd1-121d-48cf-808f-dd0d11e0336f
	`
	expectedServer := "TWL-KOM - Ludwigshafen (id = 10291)"
	expectedLatency := float32(12.42)
	var expectedDownload *float32
	expectedUpload := float32(9.41)
	expectedPacketLoss := float32(0.0)

	measurement := parseSpeedtestResult(mockOutput)
	verifyParseResult(t, &measurement, &expectedServer, &expectedLatency, expectedDownload, &expectedUpload, &expectedPacketLoss)
}

func verifyParseResult(t *testing.T, m *measurement, expectedServer *string, expectedLatency *float32, expectedDownload *float32, expectedUpload *float32, expectedPacketLoss *float32) {
	if (m.server != nil && expectedServer != nil && *(m.server) != *expectedServer) || (m.server == nil && expectedServer != nil) {
		t.Errorf("Server named not parsed correctly. Expected %s, got %s", *expectedServer, *(m.server))
	}
	if (m.latency != nil && expectedLatency != nil && *(m.latency) != *expectedLatency) || (m.latency == nil && expectedLatency != nil) {
		t.Errorf("Latency not parsed correctly. Expected %f, got %f", *expectedLatency, *(m.latency))
	}
	if (m.download != nil && expectedDownload != nil && *(m.download) != *expectedDownload) || (m.download == nil && expectedDownload != nil) {
		t.Errorf("Download not parsed correctly. Expected %f, got %f", *expectedDownload, *(m.download))
	}
	if (m.upload != nil && expectedUpload != nil && *(m.upload) != *expectedUpload) || (m.upload == nil && expectedUpload != nil) {
		t.Errorf("Upload not parsed correctly. Expected %f, got %f", *expectedUpload, *(m.upload))
	}
	if (m.packetLoss != nil && expectedPacketLoss != nil && *(m.packetLoss) != *expectedPacketLoss) || (m.packetLoss == nil && m.packetLoss != nil) {
		t.Errorf("Package loss not parsed correctly. Expected %f, got %f", *expectedPacketLoss, *(m.packetLoss))
	}
}
