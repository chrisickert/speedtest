package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const externalSpeedTestExecutable = "/usr/local/bin/speedtest"

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
	server          string
	serverError     error
	latency         float64
	latencyError    error
	download        float64
	downloadError   error
	upload          float64
	uploadError     error
	packetLoss      float64
	packetLossError error
}

func main() {
	db := connectToDatabase()
	measuredAt := time.Now()
	speedtestCmd := exec.Command(externalSpeedTestExecutable, externalSpeedTestParams()...)
	out, err := speedtestCmd.CombinedOutput()
	panicOnError(err)
	measurement := parseSpeedtestResult(string(out))
	writeToDatabase(db, measuredAt, measurement)
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
	var m measurement

	if len(outLines) > 3 {
		m.server, m.serverError = substring(outLines[3], ":")
	} else {
		m.serverError = errors.New("no server detected")
	}

	if len(outLines) > 5 {
		latencyString, err := substring(outLines[5], ":", "ms")
		if err != nil {
			m.latencyError = err
		} else {
			m.latency, m.latencyError = strconv.ParseFloat(latencyString, 32)
		}
	} else {
		m.latencyError = errors.New("no latency detected")
	}

	if len(outLines) > 6 {
		downloadString, err := substring(outLines[6], ":", "Mbps")
		if err != nil {
			m.downloadError = err
		} else {
			m.download, m.downloadError = strconv.ParseFloat(downloadString, 32)
		}
	} else {
		m.downloadError = errors.New(("no download speed detected"))
	}

	if len(outLines) > 7 {
		uploadString, err := substring(outLines[7], ":", "Mbps")
		if err != nil {
			m.uploadError = err
		} else {
			m.upload, m.uploadError = strconv.ParseFloat(uploadString, 32)
		}
	} else {
		m.uploadError = errors.New("no upload speed detected")
	}

	if len(outLines) > 8 {
		packetLossString, err := substring(outLines[8], ":", "%")
		if err != nil {
			m.packetLossError = err
		} else {
			m.packetLoss, m.packetLossError = strconv.ParseFloat(packetLossString, 32)
		}
	} else {
		m.packetLossError = errors.New("no packet loss detected")
	}

	return &m
}

func writeToDatabase(db *sql.DB, measuredAt time.Time, m *measurement) {
	// TODO: In a transaction do 1) insert new row with only id and measured_at 2) update with measurement values one by one
	_, err := db.Exec(fmt.Sprintf(`INSERT INTO %s (measured_at, server, latency_ms, download_mbps, upload_mpbs, packet_loss_percent)
	VALUES ($1, $2, $3, $4, $5, $6)`, tableName), pgValues(measuredAt, m)...)
	panicOnError(err)
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func substring(input string, separators ...string) (string, error) {
	if len(separators) != 1 && len(separators) != 2 {
		return "", errors.New("illegal number of separators (either one or two separators are supported)")
	}

	startIndex := strings.Index(input, separators[0])
	if startIndex == -1 {
		return "", fmt.Errorf("there is no sequence %s in input %s", separators[0], input)
	}

	var endIndex int
	if len(separators) == 1 {
		endIndex = -1
	} else {
		endIndex = strings.Index(input, separators[1])
	}

	if endIndex != -1 {
		return strings.TrimSpace(input[startIndex+1 : endIndex]), nil
	} else {
		return strings.TrimSpace(input[startIndex+1:]), nil
	}
}

func pgValues(measuredAt time.Time, m *measurement) []any {
	var result [6]any
	result[0] = measuredAt

	fieldMap := map[int]string{
		1: "server",
		2: "latency",
		3: "download",
		4: "upload",
		5: "packetLoss",
	}

	r := reflect.ValueOf(m)
	for index, name := range fieldMap {
		if reflect.Indirect(r).FieldByName(name + "Error").IsNil() {
			field := reflect.Indirect(r).FieldByName(name)
			if field.CanFloat() {
				result[index] = float32(field.Float())
			} else {
				result[index] = field.String()
			}
		} else {
			result[index] = nil
		}
	}
	return result[:]
}
