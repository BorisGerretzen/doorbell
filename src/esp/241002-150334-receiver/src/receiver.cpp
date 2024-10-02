#include <Arduino.h>
#define ADC_PRESCALE 16 // Set ADC clock prescaler (adjust to tune the speed)

// Peak detection variables
#define THRESHOLD 850    // Threshold for distinguishing 0 and 1
#define SUBTHRESHOLD 400 // Threshold for ignorigin mini peaks
bool isRising = false;
int lastSample = 0;

void setup()
{
  Serial.begin(115200);

  // Set up the ADC for faster sampling
  ADCSRA &= ~((1 << ADPS2) | (1 << ADPS1) | (1 << ADPS0)); // Clear prescaler bits
  ADCSRA |= (1 << ADPS1) | (1 << ADPS0);  // Set prescaler to 8 for faster sampling (can be adjusted)

  ADMUX = (1 << ADLAR) | (1 << REFS0);    // Left adjust result, use AVcc as reference
  ADCSRA |= (1 << ADEN);                  // Enable the ADC
  }

void loop()
{
  // Start ADC conversion
  ADCSRA |= (1 << ADSC);
  
  // Wait for conversion to complete
  while (ADCSRA & (1 << ADSC))
    ;

  // Read the ADC value (10-bit, right-adjusted)
  int sample = ADCH;  // Use ADCW to read both ADCL and ADCH combined

  // ignore mini peaks
  if(sample < SUBTHRESHOLD) {
    lastSample = sample;
    return;
  }

  if(sample < lastSample && isRising) {
    isRising = false;
    if(lastSample > THRESHOLD) {
      Serial.println("1");
    } else {
      Serial.println("0");
    }
  }

  if(sample > lastSample) {
    isRising = true;
  }

  lastSample = sample;
}