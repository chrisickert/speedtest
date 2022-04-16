package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	speedtestCmd := exec.Command("speedtest", "-p", "no")
	out, err := speedtestCmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", out)
}
