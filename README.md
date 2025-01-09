# air_quality_monitor
A simple air quality monitor with Arduino UNO WIFI R2

## Hardware
- Arduino UNO WIFI R2
- DHT11 sensor (PIN 2)
- MQ135 sensor (PIN A0)

## Features
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

## TO DO
- Install SDS11 sensor to monitor PM
- build a server to collect data from this sensor host and draw graphics 
