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

func main() {
	db := connectToDatabase()
	measuredAt := time.Now()
	speedtestCmd := exec.Command(externalSpeedTestExecutable, externalSpeedTestParams()...)
	out, err := speedtestCmd.CombinedOutput()
	handleError(err)
	server, latency, download, upload, packetLoss := parseSpeedtestResult(string(out))
	writeToDatabase(db, measuredAt, server, latency, download, upload, packetLoss)
	err = db.Close()
	handleError(err)
}

// connects to the database and creates the measurement table if it does not exist yet
func connectToDatabase() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	handleError(err)

	_, err = db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	handleError(err)

	_, err = db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
		measured_at TIMESTAMP WITH TIME ZONE NOT NULL,
		server TEXT,
		latency_ms FLOAT,
		download_mbps FLOAT,
		upload_mpbs FLOAT,
		packet_loss_percent FLOAT
	)`, tableName))
	handleError(err)

	_, err = db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS measured_at_idx ON %s (measured_at)", tableName))
	handleError(err)

	return db
}

// returns server, latency, download, upload, and packet loss from the given speedtest output string
func parseSpeedtestResult(out string) (string, float32, float32, float32, float32) {
	// Output is like (without the line numbers):
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
	// TODO: Make the parsing more robust ...
	server := strings.TrimSpace(outLines[3][strings.Index(outLines[3], ":")+1:])
	latency, _ := strconv.ParseFloat(strings.TrimSpace(outLines[5][strings.Index(outLines[5], ":")+1:strings.Index(outLines[5], "ms")]), 32)
	download, _ := strconv.ParseFloat(strings.TrimSpace(outLines[6][strings.Index(outLines[6], ":")+1:strings.Index(outLines[6], "Mbps")]), 32)
	upload, _ := strconv.ParseFloat(strings.TrimSpace(outLines[7][strings.Index(outLines[7], ":")+1:strings.Index(outLines[7], "Mbps")]), 32)
	packetLoss, _ := strconv.ParseFloat(strings.TrimSpace(outLines[8][strings.Index(outLines[8], ":")+1:strings.Index(outLines[8], "%")]), 32)
	return server, float32(latency), float32(download), float32(upload), float32(packetLoss)
}

func writeToDatabase(db *sql.DB, measuredAt time.Time, server string, latency float32, download float32, upload float32, packetLoss float32) {
	_, err := db.Exec(fmt.Sprintf(`INSERT INTO %s (measured_at, server, latency_ms, download_mbps, upload_mpbs, packet_loss_percent)
	VALUES ($1, $2, $3, $4, $5, $6)`, tableName), measuredAt, server, latency, download, upload, packetLoss)
	handleError(err)
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
