package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"bufio"
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


func main() {
	var failures int
	// Flags
	debug := flag.Bool("debug", false, "Enable debug logging")
	host := flag.String("host", "", "API server host. If not provided data_logger will look in your network for a compatible device.")
	interval := flag.Int("interval", 5, "Interval between measurements in minutes.")
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

	// Read host in pipefile if not provided
	if *host == "" || failures > 4 {
		failures = 0
		logger.Println("INFO - No host provided, listening for UDP packets")
		pipe, err := os.Open("/tmp/airmonpipe")
		if err != nil {
			log.Fatalf("Error opening pipe: %v", err)
			return
		}
		defer pipe.Close()
		reader := bufio.NewReader(pipe)
		message, _ := reader.ReadString('\n')
		*host=message
		logger.Println("INFO - host:", *host)
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

