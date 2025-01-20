#define SENSOR_PIN A2
const int sampleWindow = 5;                              // Sample window width in mS (5 mS = 200Hz), human can detect from 20Hz-20KHz. This value is the minimum frequency resolution
unsigned int sample;




void setup (){
	pinMode (SENSOR_PIN, INPUT); // Set the signal pin as input for MIC
	Serial.begin(9600);
}


void loop (){
	unsigned long startMillis= millis();                   // Start of sample window
	float peakToPeak = 0;                                  // peak-to-peak level

	unsigned int signalMax = 0;                            //minimum value
	unsigned int signalMin = 1024;                         //maximum value

	// collect data for sampleWindow mS
	while (millis() - startMillis < sampleWindow){
		sample = analogRead(SENSOR_PIN);                    //get reading from microphone
		if (sample < 1024){                                  // toss out spurious readings
			if (sample > signalMax){
				signalMax = sample;                           // save just the max levels
			}else if (sample < signalMin){
				signalMin = sample;                           // save just the min levels
			}
		}
	}

	peakToPeak = signalMax - signalMin;                    // max - min = peak-peak amplitude
	int db = map(peakToPeak,20,900,49.5,90);             //calibrate for deciBels
// <60 quiet, <85 moderate, >85 noisy, >100 dangerous
}
