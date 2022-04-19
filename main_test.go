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
	expectedLatency := 12.42
	expectedDownload := 103.41
	expectedUpload := 9.41
	expectedPacketLoss := 0.0

	measurement := parseSpeedtestResult(mockOutput)
	if measurement.server != expectedServer {
		t.Errorf("Server named not parsed correctly. Expected %s, got %s", expectedServer, measurement.server)
	}
	if measurement.latency != float32(expectedLatency) {
		t.Errorf("Latency not parsed correctly. Expected %f, got %f", expectedLatency, measurement.latency)
	}
	if measurement.download != float32(expectedDownload) {
		t.Errorf("Download not parsed correctly. Expected %f, got %f", expectedDownload, measurement.download)
	}
	if measurement.upload != float32(expectedUpload) {
		t.Errorf("Upload not parsed correctly. Expected %f, got %f", expectedUpload, measurement.upload)
	}
	if measurement.packetLoss != float32(expectedPacketLoss) {
		t.Errorf("Package loss not parsed correctly. Expected %f, got %f", expectedPacketLoss, measurement.packetLoss)
	}
}

// func TestParseSpeedtestResultPacketLossUnavailable(t *testing.T) {
// 	mockOutput := `
// 	Speedtest by Ookla

// 	  Server: PVDataNet - Frankfurt (id = 40094)
// 		 ISP: Plusnet
// 	 Latency:     5.47 ms   (0.12 ms jitter)
// 	Download:   102.90 Mbps (data used: 47.5 MB)
// 	  Upload:    52.63 Mbps (data used: 83.5 MB)
//  Packet Loss: Not available.
//   Result URL: https://www.speedtest.net/result/c/5dcb37c2-7e47-4780-aecc-1b40ef510d95
// 	`
// 	expectedServer := "TPVDataNet - Frankfurt"
// 	expectedLatency := 5.47
// 	expectedDownload := 102.90
// 	expectedUpload := 52.63
// 	expectedPacketLoss := "Not available"

// 	parsedServer, parsedLatency, parsedDownload, parsedUpload, parsedPackageLoss := parseSpeedtestResult(mockOutput)
// 	if parsedServer != expectedServer {
// 		t.Errorf("Server named not parsed correctly. Expected %s, got %s", expectedServer, parsedServer)
// 	}
// 	if parsedLatency != float32(expectedLatency) {
// 		t.Errorf("Latency not parsed correctly. Expected %f, got %f", expectedLatency, parsedLatency)
// 	}
// 	if parsedDownload != float32(expectedDownload) {
// 		t.Errorf("Download not parsed correctly. Expected %f, got %f", expectedDownload, parsedDownload)
// 	}
// 	if parsedUpload != float32(expectedUpload) {
// 		t.Errorf("Upload not parsed correctly. Expected %f, got %f", expectedUpload, parsedUpload)
// 	}
// 	if parsedPackageLoss != float32(expectedPacketLoss) {
// 		t.Errorf("Package loss not parsed correctly. Expected %f, got %f", expectedPacketLoss, parsedPackageLoss)
// 	}
// }
