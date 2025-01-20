package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
	"strconv"

)

type SensorData struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	CO2         float64 `json:"co2"`
	NH3         float64 `json:"nh3"`
	NOx         float64 `json:"nox"`
	Dust25      float64 `json:"pm2.5"`
	Dust10      float64 `json:"pm10"`
	CO          float64 `json:"CO"`
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
	wa_contact := flag.String("number", "", "Whatsapp contact number whitout +, es 393312345654")
	flag.Parse()

	// Logger setup
	logFile, err := os.OpenFile("alarm_aqi.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
		
		if data.CO > 20.0 {
			level:=strconv.FormatFloat(data.CO, 'f', 2, 64) // 'f' format, 2 decimal places
			logger.Printf("ALARM - CO CRITICAL LEVEL!!!")
			cmd := exec.Command("whatsapp/send_whatsapp", "LIVELLO CO CRITICO: "+level+"ppm", *wa_contact)
			_, err := cmd.CombinedOutput()
			if err != nil {
			fmt.Println("Error:", err)
			}
		}
		if data.CO2 > 1000.0 {
			level:=strconv.FormatFloat(data.CO2, 'f', 2, 64) // 'f' format, 2 decimal places
			logger.Printf("ALARM - CO2 CRITICAL LEVEL!!!")
			cmd := exec.Command("whatsapp/send_whatsapp", "LIVELLO CO2 CRITICO: "+level+"ppm", *wa_contact)
			_, err := cmd.CombinedOutput()
			if err != nil {
			fmt.Println("Error:", err)
			}
		}
		if data.NOx > 1.5 {
			level:=strconv.FormatFloat(data.NOx, 'f', 2, 64) // 'f' format, 2 decimal places
			logger.Printf("ALARM - NOx CRITICAL LEVEL!!!")
			cmd := exec.Command("whatsapp/send_whatsapp", "LIVELLO NOx CRITICO: "+level+"ppm", *wa_contact)
			_, err := cmd.CombinedOutput()
			if err != nil {
			fmt.Println("Error:", err)
			}
		}
		if data.Dust25 > 100.0 {
			level:=strconv.FormatFloat(data.Dust25, 'f', 2, 64) // 'f' format, 2 decimal places
			logger.Printf("ALARM - PM2.5 CRITICAL LEVEL!!!")
			cmd := exec.Command("whatsapp/send_whatsapp", "LIVELLO PM2.5 CRITICO: "+level+"ppm", *wa_contact)
			_, err := cmd.CombinedOutput()
			if err != nil {
			fmt.Println("Error:", err)
			}
		}
		if data.Dust10 > 100.0 {
			level:=strconv.FormatFloat(data.Dust10, 'f', 2, 64) // 'f' format, 2 decimal places
			logger.Printf("ALARM - PM10 CRITICAL LEVEL!!!")
			cmd := exec.Command("whatsapp/send_whatsapp", "LIVELLO PM10 CRITICO: "+level+"ppm", *wa_contact)
			_, err := cmd.CombinedOutput()
			if err != nil {
			fmt.Println("Error:", err)
			}
		}
		if data.Temperature > 30.0 {
			level:=strconv.FormatFloat(data.Temperature, 'f', 2, 64) // 'f' format, 2 decimal places
			logger.Printf("ALARM - TEMP CRITICAL LEVEL!!!")
			cmd := exec.Command("whatsapp/send_whatsapp", "LIVELLO TEMPERATURA CRITICO: "+level+"C", *wa_contact)
			_, err := cmd.CombinedOutput()
			if err != nil {
			fmt.Println("Error:", err)
			}
		}
		if (data.Humidity < 30.0 || data.Humidity > 80.0) {
			level:=strconv.FormatFloat(data.Humidity, 'f', 2, 64) // 'f' format, 2 decimal places
			logger.Printf("ALARM - HUM CRITICAL LEVEL!!!")
			cmd := exec.Command("whatsapp/send_whatsapp", "LIVELLO UMIDITA' CRITICO: "+level+"%", *wa_contact)
			_, err := cmd.CombinedOutput()
			if err != nil {
			fmt.Println("Error:", err)
			}
		}
		time.Sleep(time.Duration(*interval) * time.Minute)
	}
}

