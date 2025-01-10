// First Program: API Logger
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SensorData struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	CO2         float64 `json:"co2"`
	NH3         float64 `json:"nh3"`
	NOx         float64 `json:"nox"`
}

func discoverDevices() ([]string, error) {
	// Simulate discovering devices on the network
	// Replace with actual network discovery logic if needed
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var devices []string
	for _, iface := range interfaces {
		if strings.HasPrefix(iface.HardwareAddr.String(), "34:94:54") {
			devices = append(devices, iface.Name)
		}
	}
	return devices, nil
}

func main() {
	// Flags
	debug := flag.Bool("debug", false, "Enable debug logging")
	host := flag.String("host", "192.168.1.100", "API server host")
	flag.Parse()

	// Logger setup
	logFile, err := os.OpenFile("api_logger.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	if *debug {
		logger = log.New(io.MultiWriter(os.Stdout, logFile), "", log.LstdFlags)
	}

	// SQLite3 setup
	db, err := sql.Open("sqlite3", "sensor_data.db")
	if err != nil {
		logger.Fatalf("DB ERROR: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS sensor_data (
		timestamp DATETIME NOT NULL,
		temperature REAL,
		humidity REAL,
		co2 REAL,
		nh3 REAL,
		nox REAL
	)`) 
	if err != nil {
		logger.Fatalf("DB ERROR: %v", err)
	}

	logger.Println("INFO - Starting program")
	for {
		logger.Println("INFO - Fetching data from API")
		resp, err := http.Get(fmt.Sprintf("http://%s/api/data", *host))
		if err != nil {
			logger.Printf("FATAL - CONNECTION TO WEBSERVER FAILED: %v", err)
			time.Sleep(3 * time.Minute)
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Printf("FATAL - Failed to read response body: %v", err)
			time.Sleep(1 * time.Minute)
			continue
		}

		var data SensorData
		if err := json.Unmarshal(body, &data); err != nil {
			logger.Printf("FATAL -  Failed to parse JSON: %v", err)
			time.Sleep(1 * time.Minute)
			continue
		}
		logger.Printf("INFO - Data received: %+v", data)

		_, err = db.Exec(`INSERT INTO sensor_data (timestamp, temperature, humidity, co2, nh3, nox) VALUES (?, ?, ?, ?, ?, ?)`,
			time.Now(), data.Temperature, data.Humidity, data.CO2, data.NH3, data.NOx)
		if err != nil {
			logger.Printf("FATAL - DB ERROR: %v", err)
		}

		logger.Println("INFO - Data written to DB")
		time.Sleep(30 * time.Minute)
	}
}

