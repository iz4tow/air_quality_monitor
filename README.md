# Air Quality Monitor
## With Arduino WIFI R.2 sensor and Go Server
A simple air quality monitor with Arduino UNO WIFI R2 and a Go server

## Hardware
- Arduino UNO WIFI R2
- DHT11 sensor (PIN 2)
- MQ135 sensor (PIN A0)

## Sensor Features
Web server providing json data for:
- Temperature C
- Humidity %
- CO2 ppm
- NH3 ppm
- NOx ppm

example of json response:
```
{
	"temperature": 20.00,
	"humidity": 50.00,
	"co2": 551.50,
	"nh3": 0.65,
	"nox": 0.13
}
```
On the Serial Console (9600bd) you can find device IP.
Arduino Wifi will reconnect automatically in case of connection lost.

## Server
### Data Logger
Save data to a sqlite3 file
#### Compile
```
go build --trimpath data_logger.go
```

#### How to use
```
Usage of ./data_logger:
  -debug
    	Enable debug logging
  -host string
    	API server host (default "192.168.1.100")
```
Normal run
```
./data_logger -host 192.168.1.100
```

Debug (verbose logging)
```
./data_logger -host 192.168.1.100 -debug
```

### Data Plotter
Plot graph from sqlite3 data
#### Compile
```
go build --trimpath data_plotter.go
```

#### How to use
```
Usage of /tmp/go-build3139716189/b001/exe/data_plotter:
  -end string
    	End time (YYYY-MM-DD HH:MM:SS)
  -field string
    	Field to plot (all, Temperature, Humidity, CO2, NH3, NOx) (default "all")
  -onefile
    	Generate a single PNG with all graphs
  -output string
    	Output image file (default "output.png")
  -start string
    	Start time (YYYY-MM-DD HH:MM:SS)
```

Multiple image output
```
./data_plotter
```

Single image with all graph
```
./data_plotter -onefile
```



## TO DO
- Install SDS11 sensor to monitor PM
