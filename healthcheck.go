package main

import (
	"net/http"
	"os"
)

func main() {
	resp, err := http.Get("http://localhost:28080/health")
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		os.Exit(0)
	}
	os.Exit(1)
}