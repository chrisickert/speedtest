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

type measurement struct {
	server     *string
	latency    *float32
	download   *float32
	upload     *float32
	packetLoss *float32
}

func main() {
	db := connectToDatabase()
	measuredAt := time.Now()
	speedtestCmd := exec.Command(externalSpeedTestExecutable, externalSpeedTestParams()...)
	out, err := speedtestCmd.CombinedOutput()
	panicOnError(err)
	measurement := parseSpeedtestResult(string(out))
	writeToDatabase(db, measuredAt, &measurement)
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

func parseSpeedtestResult(out string) measurement {
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
	var server *string = nil
	var latency, download, upload, packetLoss *float32 = nil, nil, nil, nil

	if len(outLines) > 3 {
		startSubstr := ":"
		server = substringOrNil(&outLines[3], &startSubstr, nil)
	}
	if len(outLines) > 5 {
		startSubstr, endSubstr := ":", "ms"
		latency = float32OrNil(substringOrNil(&outLines[5], &startSubstr, &endSubstr))
	}
	if len(outLines) > 6 {
		startSubstr, endSubstr := ":", "Mbps"
		download = float32OrNil(substringOrNil(&outLines[6], &startSubstr, &endSubstr))
	}
	if len(outLines) > 7 {
		startSubstr, endSubstr := ":", "Mbps"
		upload = float32OrNil(substringOrNil(&outLines[7], &startSubstr, &endSubstr))
	}
	if len(outLines) > 8 {
		startSubstr, endSubstr := ":", "%"
		packetLoss = float32OrNil(substringOrNil(&outLines[8], &startSubstr, &endSubstr))
	}

	return measurement{server, latency, download, upload, packetLoss}
}

func writeToDatabase(db *sql.DB, measuredAt time.Time, m *measurement) {
	_, err := db.Exec(fmt.Sprintf(`INSERT INTO %s (measured_at, server, latency_ms, download_mbps, upload_mpbs, packet_loss_percent)
	VALUES ($1, $2, $3, $4, $5, $6)`, tableName), measuredAt, *(m.server), *(m.latency), *(m.download), *(m.upload), *(m.packetLoss))
	panicOnError(err)
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func substringOrNil(input *string, startSubstr *string, endSubstr *string) *string {
	startIndex := strings.Index(*input, *startSubstr)
	if startIndex == -1 {
		return nil
	}

	var endIndex int
	if endSubstr == nil {
		endIndex = -1
	} else {
		endIndex = strings.Index(*input, *endSubstr)
	}

	if endIndex != -1 {
		result := strings.TrimSpace((*input)[startIndex+1 : endIndex])
		return &result
	} else {
		result := strings.TrimSpace((*input)[startIndex+1:])
		return &result
	}
}

func float32OrNil(input *string) *float32 {
	result, err := strconv.ParseFloat(*input, 32)
	if err != nil {
		return nil
	}
	resultFloat32 := float32(result)
	return &resultFloat32
}
