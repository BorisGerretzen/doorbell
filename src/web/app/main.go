package main

import (
	"log"
	"web/common"
	"web/common/database"
)

type Services struct {
	db  *database.Database
	env common.Environment
}

var services Services

func main() {
	// Setup DB and run migrations
	db, err := database.NewDatabase("./db/settings.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	if err = common.NewMigrator().Migrate(db.Db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	env := common.GetEnvironment()
	if env.MqttHost == "" || env.MqttPort == "" {
		log.Fatal("MQTT_HOST and MQTT_PORT must be set")
	}

	services = Services{
		db:  db,
		env: env,
	}

	RunMqtt(&services)
	RunHttp(&services)
}
