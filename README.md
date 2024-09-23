# Wifi doorbell

This is a simple project to connect a radio doorbell to the internet and send a notification to users using Telegram.
The Go webapp listens to topics on a MQTT broker and sends a message to a Telegram bot when a message is received.

To capture the doorbell signal an ESP32 is combined with a CC1101 433 MHz RF module, when a doorbell signal is detected a message is published to the MQTT broker.

## How to use
Make sure you have a MQTT broker running and a Telegram bot created.

### Docker
Create a file called `prod.env` as shown below, make sure to replace the values with your own.
If you don't have certificate authentication enabled on your MQTT broker, you can remove the `MQTT_CA_PATH`, `MQTT_CERT_PATH` and `MQTT_KEY_PATH` variables.
```env
MQTT_HOST=mqtt
MQTT_PORT=8883
MQTT_USERNAME=<mqtt_username> 
MQTT_PASSWORD=<mqtt_password>
MQTT_CA_PATH=/certs/ca.crt
MQTT_CERT_PATH=/certs/server.crt
MQTT_KEY_PATH=/certs/server.key
DEVELOPMENT=false
```

Finally create an environment variable for the Telegram bot token.
This is a little clunky but I couldn't get it to work with the Github Actions secrets otherwise.
```bash
export TELEGRAM_KEY=<your_telegram_bot_token>
```

You can then use the docker-compose file to start the webapp.
```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up
```
### Standalone
You can also run the webapp standalone. 
In this case make sure that the settings you would have put in the `prod.env` file are set as environment variables.
Make sure that the certificate paths are correct.
```bash
cd src/web
go build -o .app ./app
./app
```