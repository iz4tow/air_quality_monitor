// Second Program: Graph Generator
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"image/color"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type SensorData struct {
	Timestamp   time.Time
	Temperature float64
	Humidity    float64
	CO2         float64
	NH3         float64
	NOx         float64
}

func fetchData(db *sql.DB, start, end time.Time) ([]SensorData, error) {
	rows, err := db.Query(`SELECT timestamp, temperature, humidity, co2, nh3, nox FROM sensor_data WHERE timestamp BETWEEN ? AND ?`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []SensorData
	for rows.Next() {
		var record SensorData
		if err := rows.Scan(&record.Timestamp, &record.Temperature, &record.Humidity, &record.CO2, &record.NH3, &record.NOx); err != nil {
			return nil, err
		}
		data = append(data, record)
	}
	return data, nil
}

func createPlot(data []SensorData, field string, output string) error {
	p := plot.New()
	p.Title.Text = fmt.Sprintf("%s Over Time", field)
	p.X.Label.Text = "Time"
	p.Y.Label.Text = field

	points := make(plotter.XYs, len(data))
	for i, record := range data {
		timeValue := float64(record.Timestamp.Unix())
		var yValue float64
		switch field {
		case "Temperature":
			yValue = record.Temperature
		case "Humidity":
			yValue = record.Humidity
		case "CO2":
			yValue = record.CO2
		case "NH3":
			yValue = record.NH3
		case "NOx":
			yValue = record.NOx
		}
		points[i].X = timeValue
		points[i].Y = yValue
	}

	line, err := plotter.NewLine(points)
	if err != nil {
		return err
	}
	line.Color = color.RGBA{R: 255, A: 255}
	p.Add(line)

	if err := p.Save(10*vg.Inch, 5*vg.Inch, output); err != nil {
		return err
	}
	return nil
}

func createCombinedPlot(data []SensorData, output string) error {
	fields := []struct {
		Name  string
		Color color.RGBA
		Extract func(SensorData) float64
	}{
		{"Temperature", color.RGBA{R: 255, A: 255}, func(d SensorData) float64 { return d.Temperature }},
		{"Humidity", color.RGBA{B: 255, A: 255}, func(d SensorData) float64 { return d.Humidity }},
		{"CO2", color.RGBA{G: 255, A: 255}, func(d SensorData) float64 { return d.CO2 }},
		{"NH3", color.RGBA{R: 128, G: 0, B: 128, A: 255}, func(d SensorData) float64 { return d.NH3 }},
		{"NOx", color.RGBA{R: 0, G: 128, B: 128, A: 255}, func(d SensorData) float64 { return d.NOx }},
	}

	p := plot.New()
	p.Title.Text = "Sensor Data Over Time"
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Value"

	for _, field := range fields {
		points := make(plotter.XYs, len(data))
		for i, record := range data {
			points[i].X = float64(record.Timestamp.Unix())
			points[i].Y = field.Extract(record)
		}

		line, err := plotter.NewLine(points)
		if err != nil {
			return err
		}
		line.Color = field.Color
		line.Width = vg.Points(2)
		p.Add(line)
		p.Legend.Add(field.Name, line)
	}

	p.Legend.Top = true

	if err := p.Save(15*vg.Inch, 10*vg.Inch, output); err != nil {
		return err
	}
	return nil
}

func main() {
	// Flags
	field := flag.String("field", "all", "Field to plot (all, Temperature, Humidity, CO2, NH3, NOx)")
	output := flag.String("output", "output.png", "Output image file")
	startTime := flag.String("start", "", "Start time (YYYY-MM-DD HH:MM:SS)")
	endTime := flag.String("end", "", "End time (YYYY-MM-DD HH:MM:SS)")
	oneFile := flag.Bool("onefile", false, "Generate a single PNG with all graphs")
	flag.Parse()

	// Parse time range
	var start, end time.Time
	var err error

	if *startTime == "" {
		start = time.Time{} // From the beginning
	} else {
		start, err = time.Parse("2006-01-02 15:04:05", *startTime)
		if err != nil {
			log.Fatalf("Invalid start time: %v", err)
		}
	}

	if *endTime == "" {
		end = time.Now() // Till now
	} else {
		end, err = time.Parse("2006-01-02 15:04:05", *endTime)
		if err != nil {
			log.Fatalf("Invalid end time: %v", err)
		}
	}

	// SQLite3 setup
	db, err := sql.Open("sqlite3", "sensor_data.db")
	if err != nil {
		log.Fatalf("DB ERROR: %v", err)
	}
	defer db.Close()

	data, err := fetchData(db, start, end)
	if err != nil {
		log.Fatalf("Failed to fetch data: %v", err)
	}

	if *oneFile {
		if err := createCombinedPlot(data, *output); err != nil {
			log.Fatalf("Failed to create combined plot: %v", err)
		}
	} else if *field == "all" {
		fields := []string{"Temperature", "Humidity", "CO2", "NH3", "NOx"}
		for _, f := range fields {
			outputFile := fmt.Sprintf("%s_%s.png", *output, f)
			if err := createPlot(data, f, outputFile); err != nil {
				log.Printf("Failed to create plot for %s: %v", f, err)
			}
		}
	} else {
		if err := createPlot(data, *field, *output); err != nil {
			log.Fatalf("Failed to create plot: %v", err)
		}
	}

	log.Println("Plots generated successfully")
}

