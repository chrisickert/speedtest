package main

import (
	"errors"
	"testing"
	"time"
)

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
	expectedMeasurement := measurement{
		server:     "TWL-KOM - Ludwigshafen (id = 10291)",
		latency:    float64(float32(12.42)),
		download:   float64(float32(103.41)),
		upload:     float64(float32(9.41)),
		packetLoss: float64(float32(0)),
	}
	measurement := parseSpeedtestResult(mockOutput)
	verifyParseResult(t, measurement, &expectedMeasurement)
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
	expectedMeasurement := measurement{
		server:          "PVDataNet - Frankfurt (id = 40094)",
		latency:         float64(float32(5.47)),
		download:        float64(float32(102.90)),
		upload:          float64(float32(52.63)),
		packetLossError: errors.New("not available"),
	}
	measurement := parseSpeedtestResult(mockOutput)
	verifyParseResult(t, measurement, &expectedMeasurement)
}

func TestParseSpeedtestResultInvalidCommand(t *testing.T) {
	mockOutput := `an error occurred`
	expectedMeasurement := measurement{
		serverError:     errors.New("not available"),
		latencyError:    errors.New("not available"),
		downloadError:   errors.New("not available"),
		uploadError:     errors.New("not available"),
		packetLossError: errors.New("not available"),
	}
	measurement := parseSpeedtestResult(mockOutput)
	verifyParseResult(t, measurement, &expectedMeasurement)
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
	expectedMeasurement := measurement{
		server:        "TWL-KOM - Ludwigshafen (id = 10291)",
		latency:       float64(float32(12.42)),
		downloadError: errors.New("not available"),
		upload:        float64(float32(9.41)),
		packetLoss:    float64(float32(0)),
	}
	measurement := parseSpeedtestResult(mockOutput)
	verifyParseResult(t, measurement, &expectedMeasurement)
}

func verifyParseResult(t *testing.T, m *measurement, expected *measurement) {
	if (m.serverError == nil && expected.serverError == nil && m.server != expected.server) || (m.serverError != nil && expected.serverError == nil) || (m.serverError == nil && expected.serverError != nil) {
		t.Errorf("Server not parsed correctly. Expected (%s, %s), got (%s, %s)", expected.server, expected.serverError, m.server, m.serverError)
	}
	if (m.latencyError == nil && expected.latencyError == nil && m.latency != expected.latency) || (m.latencyError != nil && expected.latencyError == nil) || (m.latencyError == nil && expected.latencyError != nil) {
		t.Errorf("Latency not parsed correctly. Expected (%f, %s), got (%f, %s)", expected.latency, expected.latencyError, m.latency, m.latencyError)
	}
	if (m.downloadError == nil && expected.downloadError == nil && m.download != expected.download) || (m.downloadError != nil && expected.downloadError == nil) || (m.downloadError == nil && expected.downloadError != nil) {
		t.Errorf("Download not parsed correctly. Expected (%f, %s), got (%f, %s)", expected.download, expected.downloadError, m.download, m.downloadError)
	}
	if (m.uploadError == nil && expected.uploadError == nil && m.upload != expected.upload) || (m.uploadError != nil && expected.uploadError == nil) || (m.uploadError == nil && expected.uploadError != nil) {
		t.Errorf("Upload not parsed correctly. Expected (%f, %s), got (%f, %s)", expected.upload, expected.uploadError, m.upload, m.uploadError)
	}
	if (m.packetLossError == nil && expected.packetLossError == nil && m.packetLoss != expected.packetLoss) || (m.packetLossError != nil && expected.packetLossError == nil) || (m.packetLossError == nil && expected.packetLossError != nil) {
		t.Errorf("Package loss not parsed correctly. Expected (%f, %s), got (%f, %s)", expected.packetLoss, expected.packetLossError, m.packetLoss, m.packetLossError)
	}
}

func TestPgValuesAllSet(t *testing.T) {
	measuredAt := time.Now()
	m := measurement{
		server:     "TWL-KOM - Ludwigshafen (id = 10291)",
		latency:    float64(float32(12.42)),
		download:   float64(float32(103.41)),
		upload:     float64(float32(9.41)),
		packetLoss: float64(float32(0)),
	}
	values := pgValues(measuredAt, &m)

	if values[0] != measuredAt {
		t.Error("measured-at value not set correctly")
	}
	if values[1] != "TWL-KOM - Ludwigshafen (id = 10291)" {
		t.Error("server value not set correctly")
	}
	if values[2] != float32(12.42) {
		t.Error("latency value not set correctly")
	}
	if values[3] != float32(103.41) {
		t.Error("download value not set correctly")
	}
	if values[4] != float32(9.41) {
		t.Error("upload value not set correctly")
	}
	if values[5] != float32(0.0) {
		t.Error("packet loss value not set correctly")
	}
}

func TestPgValuesWithNil(t *testing.T) {
	measuredAt := time.Now()
	m := measurement{
		server:          "TWL-KOM - Ludwigshafen (id = 10291)",
		latency:         float64(float32(12.42)),
		download:        float64(float32(103.41)),
		upload:          float64(float32(9.41)),
		packetLossError: errors.New("not available"),
	}
	values := pgValues(measuredAt, &m)

	if values[0] != measuredAt {
		t.Error("measured-at value not set correctly")
	}
	if values[1] != "TWL-KOM - Ludwigshafen (id = 10291)" {
		t.Error("server value not set correctly")
	}
	if values[2] != float32(12.42) {
		t.Error("latency value not set correctly")
	}
	if values[3] != float32(103.41) {
		t.Error("download value not set correctly")
	}
	if values[4] != float32(9.41) {
		t.Error("upload value not set correctly")
	}
	if values[5] != nil {
		t.Error("packet loss value not set correctly")
	}
}
