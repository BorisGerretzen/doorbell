package common

import "os"

type Environment struct {
	MqttHost     string
	MqttPort     string
	MqttUsername string
	MqttPassword string
	MqttCaPath   string
	MqttCertPath string
	MqttKeyPath  string
	TelegramKey  string
	Development  bool
	HttpPort     string
}

func GetEnvironment() Environment {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "9000"
	}

	return Environment{
		MqttHost:     os.Getenv("MQTT_HOST"),
		MqttPort:     os.Getenv("MQTT_PORT"),
		MqttUsername: os.Getenv("MQTT_USERNAME"),
		MqttPassword: os.Getenv("MQTT_PASSWORD"),
		MqttCaPath:   os.Getenv("MQTT_CA_PATH"),
		MqttCertPath: os.Getenv("MQTT_CERT_PATH"),
		MqttKeyPath:  os.Getenv("MQTT_KEY_PATH"),
		TelegramKey:  os.Getenv("TELEGRAM_KEY"),
		Development:  os.Getenv("DEVELOPMENT") == "true",
		HttpPort:     port,
	}
}
