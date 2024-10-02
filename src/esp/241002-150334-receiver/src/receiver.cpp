#include <Arduino.h>
#define RFIN_PIN 11

void rfDetect_ISR();

void setup()
{
  Serial.begin(9600);

  // rfTimer = timerBegin(0, 8, true);
  // timerAttachInterrupt(rfTimer, &rfDetectTimer_ISR, true);
  // timerAlarmWrite(rfTimer, STREAM_BITDETECT, false);
  // timerAlarmDisable(rfTimer);

  pinMode(LED_BUILTIN, OUTPUT);
  pinMode(RFIN_PIN, INPUT_PULLUP);                 // RFIN_PIN configed as input pin
  attachInterrupt(RFIN_PIN, rfDetect_ISR, CHANGE); //   and enable edge detection interrupt
}

void loop(){}

void rfDetect_ISR()
{
  uint8_t pinState = digitalRead(RFIN_PIN); // Read the RFIN_PIN logical level
  if (pinState == LOW)
  { // if logic level is LOW, that is Falling Edge
    digitalWrite(LED_BUILTIN, LOW);
    return;
  }

  digitalWrite(LED_BUILTIN, HIGH);

  Serial.println("RF state " + String(pinState) + "\n");
}