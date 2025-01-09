#include <WiFi.h>
#include <DHT.h>

// WiFi credentials
const char* ssid = "WIFI_AP_SSID";
const char* password = "WIFI PASSWORD";

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

// WiFi server on port 80
WiFiServer server(80);

void setup() {
  // Initialize Serial Monitor
  Serial.begin(9600);
  Serial.println("MQ135 Multi-Gas Monitor with Web Server");

  // Initialize DHT sensor
  dht.begin();

  // Connect to WiFi
  WiFi.begin(ssid, password);
  Serial.print("Connecting to WiFi");
  while (WiFi.status() != WL_CONNECTED) {
    delay(1000);
    Serial.print(".");
  }
  Serial.println("\nWiFi connected");

  // Get the local IP address as a string
  IPAddress localIP = WiFi.localIP();
  String ipString = String(localIP[0]) + "." +
                    String(localIP[1]) + "." +
                    String(localIP[2]) + "." +
                    String(localIP[3]);

  Serial.print("IP address: ");
  Serial.println(ipString);

  // Start the server
  server.begin();
}

void loop() {
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
  // Read temperature and humidity from DHT22 sensor
  float temperature = dht.readTemperature();
  float humidity = dht.readHumidity();

  //if DHT fails it set hum 50, temp 20
  if (isnan(temperature)) temperature=20;
  if (isnan(humidity)) humidity=50;

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

  // Build JSON response
  String jsonResponse = "{";
  jsonResponse += "\"temperature\": " + String(temperature, 2) + ",";
  jsonResponse += "\"humidity\": " + String(humidity, 2) + ",";
  jsonResponse += "\"co2\": " + String(ppmCO2, 2) + ",";
  jsonResponse += "\"nh3\": " + String(ppmNH3, 2) + ",";
  jsonResponse += "\"nox\": " + String(ppmNOx, 2);
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
  String response = "Please visit http://" + ipString + "/api/data";

  // Send the HTTP response
  client.println("HTTP/1.1 404 Not Found");
  client.println("Content-Type: text/plain");
  client.println("Connection: close");
  client.println("Content-Length: " + String(response.length())); // Explicit content length
  client.println(); // Blank line separating headers from body
  client.println(response); // Response body

  // Debugging output
  Serial.println("404 Not Found response sent:");
  Serial.println(response);
}
