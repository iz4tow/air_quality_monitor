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
	Dust25      float64 `json:"pm25"`
	Dust10      float64 `json:"pm10"`
}

func discoverHost() (string, error) {
	conn, err := net.ListenPacket("udp", ":8888")
	if err != nil {
		return "", fmt.Errorf("failed to listen for UDP packets: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	deadline := time.Now().Add(60 * time.Second)
	err = conn.SetReadDeadline(deadline)
	if err != nil {
		return "", fmt.Errorf("failed to set read deadline: %v", err)
	}

	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			if os.IsTimeout(err) {
				return "", fmt.Errorf("no UDP packets received within 60 seconds")
			}
			return "", fmt.Errorf("error reading UDP packet: %v", err)
		}

		message := string(buf[:n])
		if strings.HasPrefix(message, "Franco-AQM:") {
			host := strings.TrimSpace(strings.TrimPrefix(message, "Franco-AQM:"))
			return host, nil
		}
	}
}

func main() {
	var failures int
	// Flags
	debug := flag.Bool("debug", false, "Enable debug logging")
	host := flag.String("host", "", "API server host. If not provided data_logger will look in your network for a compatible device.")
	interval := flag.Int("interval", 30, "Interval between measurements in minutes.")
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

	// Discover host if not provided
	if *host == "" || failures > 4 {
		failures = 0
		logger.Println("INFO - No host provided, listening for UDP packets")
		var err error = fmt.Errorf("initial error to loop")
		for err != nil {
			var hostAddr string
			hostAddr, err = discoverHost()
			if err != nil {
				logger.Printf("WARNING - Failed to discover host: %v", err)
			}
			*host = hostAddr
			logger.Printf("INFO - Discovered host: %s", *host)
		}
	}

	// SQLite3 setup
	db, err := sql.Open("sqlite3", "sensor_data.db")
	if err != nil {
		logger.Fatalf("DB ERROR: %v", err)
	}
	defer db.Close()

	// Ensure the table exists
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS sensor_data (
		timestamp DATETIME NOT NULL,
		temperature REAL,
		humidity REAL,
		co2 REAL,
		nh3 REAL,
		nox REAL,
		pm25 REAL,
		pm10 REAL
	)`)
	if err != nil {
		logger.Fatalf("DB ERROR: %v", err)
	}

	// Check for missing columns and add them
	requiredColumns := map[string]string{
		"temperature": "REAL",
		"humidity":    "REAL",
		"co2":         "REAL",
		"nh3":         "REAL",
		"nox":         "REAL",
		"pm25":        "REAL",
		"pm10":        "REAL",
	}

	rows, err := db.Query(`PRAGMA table_info(sensor_data)`)
	if err != nil {
		logger.Fatalf("DB ERROR: %v", err)
	}
	defer rows.Close()

	existingColumns := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			logger.Fatalf("DB ERROR: %v", err)
		}
		existingColumns[name] = true
	}

	for col, colType := range requiredColumns {
		if !existingColumns[col] {
			logger.Printf("INFO - Adding missing column: %s %s", col, colType)
			_, err := db.Exec(fmt.Sprintf(`ALTER TABLE sensor_data ADD COLUMN %s %s`, col, colType))
			if err != nil {
				logger.Fatalf("DB ERROR: Failed to add column %s: %v", col, err)
			}
		}
	}

	logger.Println("INFO - Starting program")
	for {
		logger.Println("INFO - Fetching data from API")
		resp, err := http.Get(fmt.Sprintf("http://%s/api/data", *host))
		if err != nil {
			logger.Printf("FATAL - CONNECTION TO WEBSERVER FAILED: %v", err)
			time.Sleep(3 * time.Minute)
			failures++
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
			logger.Printf("FATAL - Failed to parse JSON: %v", err)
			time.Sleep(1 * time.Minute)
			continue
		}
		logger.Printf("INFO - Data received: %+v", data)

		_, err = db.Exec(`INSERT INTO sensor_data (timestamp, temperature, humidity, co2, nh3, nox, pm25, pm10) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			time.Now(), data.Temperature, data.Humidity, data.CO2, data.NH3, data.NOx, data.Dust25, data.Dust10)
		if err != nil {
			logger.Printf("FATAL - DB ERROR: %v", err)
		}

		logger.Println("INFO - Data written to DB")
		time.Sleep(time.Duration(*interval) * time.Minute)
	}
}

