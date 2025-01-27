#include "SdsDustSensor.h"

int rxPin = 0;
int txPin = 1;

SdsDustSensor sds(rxPin,txPin);

void setup() {

  Serial.begin(9600);
  sds.begin();
  Serial.println(sds.queryFirmwareVersion().toString()); // prints firmware version
  Serial.println(sds.setActiveReportingMode().toString()); // ensure sensor is in 'active' reporting mode
  Serial.println(sds.setContinuousWorkingPeriod().toString());
}


void loop() {
    PmResult pm = sds.readPm();
    if (pm.isOk()) {
      float dust25 = pm.pm25;
      float dust10 = pm.pm10;
      Serial.print("PM2.5 = ");
      Serial.print(pm.pm25);
      Serial.print(", PM10 = ");
      Serial.println(pm.pm10);
      //if you want to just print the measured values, you can use toString() method as well
      Serial.println(pm.toString());
  }   else {
    //notice that loop delay is set to .5s
    Serial.print("could not read values from sensor, reason: ");
    Serial.println(pm.statusToString());
  }
  delay(500);  
}