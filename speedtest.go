package main

import (
	"database/sql"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const externalSpeedTestExecutable = "speedtest"

func externalSpeedTestParams() []string {
	return []string{"-p", "no"}
}

const (
	// TODO: Make this configurable (and don't put sensitive data to GitHub)
	host      = "localhost"
	port      = 5432
	user      = "speedtest"
	password  = "speedtest"
	dbname    = "speedtest"
	tableName = "measurement"
)

const (
	nilFloat  = float32(-1.0)
	nilString = "nil"
)

type measurement struct {
	server     string
	latency    float32
	download   float32
	upload     float32
	packetLoss float32
}

func main() {
	db := connectToDatabase()
	measuredAt := time.Now()
	speedtestCmd := exec.Command(externalSpeedTestExecutable, externalSpeedTestParams()...)
	out, err := speedtestCmd.CombinedOutput()
	panicOnError(err)
	measurement := parseSpeedtestResult(string(out))
	writeToDatabase(db, measuredAt, measurement.server, measurement.latency, measurement.download, measurement.upload, measurement.packetLoss)
	db.Close()
}

// connects to the database and creates the measurement table if it does not exist yet
func connectToDatabase() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	panicOnError(err)

	_, err = db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	panicOnError(err)

	_, err = db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
		measured_at TIMESTAMP WITH TIME ZONE NOT NULL,
		server TEXT,
		latency_ms FLOAT,
		download_mbps FLOAT,
		upload_mpbs FLOAT,
		packet_loss_percent FLOAT
	)`, tableName))
	panicOnError(err)

	_, err = db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS measured_at_idx ON %s (measured_at)", tableName))
	panicOnError(err)

	return db
}

func substringOrNil(input string, startSubstr string, endSubstr string) string {
	startIndex := strings.Index(input, startSubstr)
	if startIndex == -1 {
		return nilString
	}

	var endIndex int
	if endSubstr == nilString {
		endIndex = -1
	} else {
		endIndex = strings.Index(input, endSubstr)
	}

	if endIndex != -1 {
		return strings.TrimSpace(input[startIndex+1 : endIndex])
	} else {
		return strings.TrimSpace(input[startIndex+1:])
	}
}

func float32OrNil(input string) float32 {
	result, err := strconv.ParseFloat(input, 32)
	if err != nil {
		return nilFloat
	}
	return float32(result)
}

// returns server, latency, download, upload, and packet loss from the given speedtest output string
func parseSpeedtestResult(out string) *measurement {
	// Expected output is like (without the line numbers):
	//  0:
	//  1: Speedtest by Ookla
	//  2:
	//  3:  Server: TWL-KOM - Ludwigshafen (id = 10291)
	//  4:  ISP: Vodafone Germany Cable
	//  5:  Latency:    12.42 ms   (2.62 ms jitter)
	//  6:  Download:   103.41 Mbps (data used: 99.3 MB )
	//  7:  Upload:     9.41 Mbps (data used: 4.7 MB )
	//  8:  Packet Loss:     0.0%
	//  9:  Result URL: https://www.speedtest.net/result/c/8f37ffd1-121d-48cf-808f-dd0d11e0336f
	// 10:
	outLines := strings.Split(out, "\n")
	server, latency, download, upload, packetLoss := nilString, nilFloat, nilFloat, nilFloat, nilFloat

	if len(outLines) > 3 {
		server = substringOrNil(outLines[3], ":", nilString)
	}
	if len(outLines) > 5 {
		latency = float32OrNil(substringOrNil(outLines[5], ":", "ms"))
	}
	if len(outLines) > 6 {
		download = float32OrNil(substringOrNil(outLines[6], ":", "Mbps"))
	}
	if len(outLines) > 7 {
		upload = float32OrNil(substringOrNil(outLines[7], ":", "Mbps"))
	}
	if len(outLines) > 8 {
		packetLoss = float32OrNil(substringOrNil(outLines[8], ":", "%"))
	}

	return &measurement{server, float32(latency), float32(download), float32(upload), float32(packetLoss)}
}

func writeToDatabase(db *sql.DB, measuredAt time.Time, server string, latency float32, download float32, upload float32, packetLoss float32) {
	_, err := db.Exec(fmt.Sprintf(`INSERT INTO %s (measured_at, server, latency_ms, download_mbps, upload_mpbs, packet_loss_percent)
	VALUES ($1, $2, $3, $4, $5, $6)`, tableName), measuredAt, server, latency, download, upload, packetLoss)
	panicOnError(err)
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
