package main

import (
	"log"
	"web/common"
)

func RunMqtt(services *Services) {
	log.Println("Connecting to MQTT")

	if services.env.TelegramKey == "" {
		log.Println("!!! Telegram key not set, skipping telegram notifications !!!")
	}

	mqtt := common.NewMqttHandler(services.env)
	if err := mqtt.Connect(); err != nil {
		log.Fatalf("Failed to connect to MQTT: %v", err)
	}
	log.Println("Connected to MQTT")

	mqtt.ListenDoorbell(func(deviceName string, payload string) {
		log.Printf("Received doorbell message from %s: %s", deviceName, payload)
		err := services.db.UpdateLastSeen(deviceName)
		if err != nil {
			log.Printf("Failed to update last seen for %s: %v", deviceName, err)
			return
		}

		// Get device config
		users, err := services.db.GetDeviceUsers(deviceName)
		if err != nil {
			log.Printf("Failed to get config for %s: %v", deviceName, err)
			return
		}

		// Get chat ids
		chatIds := make([]string, len(users))
		for i, user := range users {
			chatIds[i] = user.ChatId
		}

		// Send notification
		if services.env.TelegramKey != "" {
			log.Printf("Sending telegram notification")
			err = common.SendTelegramDoorbell(services.env.TelegramKey, "Ding dong er is iemand aan de deur!", chatIds)
			if err != nil {
				log.Printf("Failed to send telegram notification: %v", err)
			}
		}
	})
}
