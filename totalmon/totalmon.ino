#include <WiFi.h>
#include <DHT.h>
#include <WiFiUdp.h> // Include the WiFiUDP library
#include "SdsDustSensor.h"

int rxPin = 0;
int txPin = 1;

SdsDustSensor sds(rxPin,txPin);

// WiFi credentials
const char* ssid = "SSID";
const char* password = "PASSWORD SECURE";

// Define MQ135 sensor pin
#define MQ135_PIN A0

// Define DHT22 sensor pin and type
#define DHTPIN 2
#define DHTTYPE DHT11

// Initialize DHT sensor
DHT dht(DHTPIN, DHTTYPE);

// Calibration constants for MQ135
const float RLOAD = 10.0;  // Load resistance in kilo-ohms
const float RZERO = 76.63; // Sensor resistance in fresh air (calibrated value)

// Baseline PPM levels (fresh air) for each gas
const float PPM_CO2_BASE = 400.0;
const float PPM_NH3_BASE = 0.5;
const float PPM_NOX_BASE = 0.1;

// Sensitivity constants for each gas (must be tuned experimentally)
const float CO2_SLOPE = -0.42;
const float NH3_SLOPE = -0.35;
const float NOX_SLOPE = -0.30;

// Temperature and humidity compensation factors
const float TEMP_COEFF = 0.01; // Adjust based on datasheet or experiments
const float HUMID_COEFF = 0.02; // Adjust based on datasheet or experiments

// Time threshold for WiFi reconnection (in milliseconds)
const unsigned long RECONNECT_TIMEOUT = 30000; // 30 seconds
unsigned long lastConnectionCheck = 0;
unsigned long lastHello = 0;


// WiFi server on port 80
WiFiServer server(80);

// UDP broadcast settings
WiFiUDP udp;
const unsigned int BROADCAST_PORT = 8888; // You can choose any available port
const String UDP_MESSAGE_PREFIX = "Franco-AQM:";

void helloimhere(){
  // Get the local IP address as a string
  IPAddress localIP = WiFi.localIP();
  String ipString = String(localIP[0]) + "." +
                    String(localIP[1]) + "." +
                    String(localIP[2]) + "." +
                    String(localIP[3]);
  Serial.print("IP address: ");
  Serial.println(ipString);
  // Prepare the UDP broadcast message
  String udpMessage = UDP_MESSAGE_PREFIX + ipString;
  // Determine the broadcast address
  IPAddress broadcastIP = localIP;
  broadcastIP[3] = 255; // Set the last octet to 255 for broadcast
  // Send the UDP broadcast
  udp.beginPacket(broadcastIP, BROADCAST_PORT);
  udp.write(udpMessage.c_str());
  udp.endPacket();
  Serial.print("UDP broadcast sent to ");
  Serial.print(broadcastIP);
  Serial.print(":");
  Serial.println(BROADCAST_PORT);
}

void wifireconnect(){
  int i=0;
  WiFi.disconnect();
  Serial.println("WiFi disconnected");
  WiFi.begin(ssid, password);
  Serial.println("Connecting...");
  while (WiFi.status() != WL_CONNECTED) {
    delay(1000);
    Serial.print(".");
    i++;
    if (i>15){
      Serial.println("Wifi connection failed");
      WiFi.end();
      delay(1000);
      return;  
    } 
  }
  helloimhere();
}

void setup() {
  int i=0;
  // Initialize Serial Monitor
  Serial.begin(9600);
  Serial.println("MQ135 Multi-Gas Monitor with Web Server");

  // Initialize DHT sensor
  dht.begin();

  if (WiFi.status() != WL_CONNECTED) wifireconnect();

  // Start the server
  server.begin();
  
  // Initialize UDP (optional if only sending)
  udp.begin(BROADCAST_PORT); // This is optional for sending only

  //dust sensor setup
  sds.begin();
  Serial.println(sds.queryFirmwareVersion().toString()); // prints firmware version
  Serial.println(sds.setActiveReportingMode().toString()); // ensure sensor is in 'active' reporting mode
  Serial.println(sds.setContinuousWorkingPeriod().toString());
}

int hello=0;
void loop() {

  unsigned long currentMillis = millis();
  if (hello==0){
    lastHello = millis();
    hello=1;
  } else if (currentMillis - lastHello >= RECONNECT_TIMEOUT) {
  //send UDP packet to be discovered
    helloimhere();
    hello=0;
  }


  // Check WiFi status
  if (WiFi.status() != WL_CONNECTED) {
    if (lastConnectionCheck == 0) {
      lastConnectionCheck = currentMillis; // Initialize the timer
    } else if (currentMillis - lastConnectionCheck >= RECONNECT_TIMEOUT) {
      wifireconnect();
    }
  } else {
    // Reset the timer when WiFi is connected
    lastConnectionCheck = 0;
  }

  // Check for client connection
  WiFiClient client = server.available();
  if (client) {
    Serial.println("New client connected");
    String request = client.readStringUntil('\r');
    Serial.println(request);

    // Check if the request is a GET request
    if (request.indexOf("GET /api/data") >= 0) {
      sendJsonResponse(client);
    } else {
      sendNotFoundResponse(client);
    }

    // Close the connection
    client.stop();
    Serial.println("Client disconnected");
  }
}

void sendJsonResponse(WiFiClient& client) {
  // Read temperature and humidity from DHT11 sensor
  float temperature = dht.readTemperature();
  float humidity = dht.readHumidity();

  // If DHT fails, set default values
  if (isnan(temperature)) temperature = 20;
  if (isnan(humidity)) humidity = 50;

  // Read analog value from MQ135 sensor
  int analogValue = analogRead(MQ135_PIN);

  // Convert analog reading to voltage
  float voltage = analogValue * (5.0 / 1023.0);

  // Calculate sensor resistance (RS)
  float RS = ((5.0 - voltage) / voltage) * RLOAD;

  // Calculate RS/RZERO ratio
  float ratio = RS / RZERO;

  // Apply temperature and humidity compensation to the ratio
  float tempCompensation = 1 + TEMP_COEFF * (temperature - 20); // Reference temperature: 20°C
  float humidCompensation = 1 - HUMID_COEFF * (humidity - 50);  // Reference humidity: 50%
  float compensatedRatio = ratio * tempCompensation * humidCompensation;

  // Estimate gas concentrations using compensated ratio
  float ppmCO2 = PPM_CO2_BASE * pow(compensatedRatio, CO2_SLOPE);
  float ppmNH3 = PPM_NH3_BASE * pow(compensatedRatio, NH3_SLOPE);
  float ppmNOx = PPM_NOX_BASE * pow(compensatedRatio, NOX_SLOPE);
  
  //dust data
  PmResult pm = sds.readPm();
  float dust25,dust10;
    if (pm.isOk()) {
      dust25 = pm.pm25; 
      dust10 = pm.pm25;
      Serial.print("PM2.5 = ");
      Serial.print(pm.pm25);
      Serial.print(", PM10 = ");
      Serial.println(pm.pm10);
      //if you want to just print the measured values, you can use toString() method as well
      Serial.println(pm.toString());
  }   else {
    //notice that loop delay is set to .5s
    Serial.print("could not read values from PM sensor");
      dust25 = NAN;
      dust10 = NAN;
  }
  // Build JSON response
  String jsonResponse = "{";
  jsonResponse += "\"temperature\": " + String(temperature, 2) + ",";
  jsonResponse += "\"humidity\": " + String(humidity, 2) + ",";
  jsonResponse += "\"co2\": " + String(ppmCO2, 2) + ",";
  jsonResponse += "\"nh3\": " + String(ppmNH3, 2) + ",";
  jsonResponse += "\"nox\": " + String(ppmNOx, 2) + ",";
  jsonResponse += "\"PM2.5\": " + String(dust25, 2) + ",";
  jsonResponse += "\"PM10\": " + String(dust10, 2);
  jsonResponse += "}";

  // Send HTTP response
  client.println("HTTP/1.1 200 OK");
  client.println("Content-Type: application/json");
  client.println("Connection: close");
  client.println("Content-Length: " + String(jsonResponse.length())); // Explicit content length
  client.println();
  client.println(jsonResponse);

  Serial.println("JSON response sent:");
  Serial.println(jsonResponse);
}

void sendNotFoundResponse(WiFiClient& client) {
  // Get the local IP address as a string
  IPAddress localIP = WiFi.localIP();
  String ipString = String(localIP[0]) + "." +
                    String(localIP[1]) + "." +
                    String(localIP[2]) + "." +
                    String(localIP[3]);

  // Prepare the 404 response content
  String response = "Please visit <a href=http://" + ipString + "/api/data>http://" + ipString + "/api/data</a>";

  // Send the HTTP response
  client.println("HTTP/1.1 404 Not Found");
  client.println("Content-Type: text/html");
  client.println("Connection: close");
  client.println("Content-Length: " + String(response.length())); // Explicit content length
  client.println();
  client.println(response); // Response body

  // Debugging output
  Serial.println("404 Not Found response sent:");
  Serial.println(response);
}
