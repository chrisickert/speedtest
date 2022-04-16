package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

const speedtestExecutable = "speedtest"

const (
	host     = "localhost"
	port     = 5432
	user     = "speedtest"
	password = "<read-from-env>"
	dbname   = "speedtest"
)

func main() {
	// speedtestCmd := exec.Command("speedtest", "-p", "no")
	// out, err := speedtestCmd.CombinedOutput()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	out := `
Speedtest by Ookla

  Server: TWL-KOM - Ludwigshafen (id = 10291)
	 ISP: Vodafone Germany Cable
 Latency:    12.42 ms   (2.62 ms jitter)
Download:   103.41 Mbps (data used: 99.3 MB )
  Upload:     9.41 Mbps (data used: 4.7 MB )
Packet Loss:     0.0%
Result URL: https://www.speedtest.net/result/c/8f37ffd1-121d-48cf-808f-dd0d11e0336f
`
	fmt.Printf("%s\n", out)
	parse_speedtest_result(out)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
}

// returns server, latency, download, upload, and packet loss from the given speedtest output string
func parse_speedtest_result(out string) (string, float32, float32, float32, float32) {
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
	server := strings.TrimSpace(strings.TrimLeft(outLines[3], "Server: "))
	latency, _ := strconv.ParseFloat(strings.TrimSpace(outLines[5][strings.Index(outLines[5], ":")+1:strings.Index(outLines[5], "ms")]), 32)
	download, _ := strconv.ParseFloat(strings.TrimSpace(outLines[6][strings.Index(outLines[6], ":")+1:strings.Index(outLines[6], "Mbps")]), 32)
	upload, _ := strconv.ParseFloat(strings.TrimSpace(outLines[7][strings.Index(outLines[7], ":")+1:strings.Index(outLines[7], "Mbps")]), 32)
	packetLoss, _ := strconv.ParseFloat(strings.TrimSpace(outLines[8][strings.Index(outLines[8], ":")+1:strings.Index(outLines[8], "%")]), 32)
	return server, float32(latency), float32(download), float32(upload), float32(packetLoss)
}
