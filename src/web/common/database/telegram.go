package database

import (
	"database/sql"
	"web/common"
)

type DeviceConfig struct {
	Users []common.TelegramUser `json:"users"`
}

func (d *Database) GetDeviceUsers(deviceName string) ([]common.TelegramUser, error) {
	// Get users
	query := `
SELECT telegram_users.chat_id, telegram_users.user 
FROM telegram_users 
    JOIN device_users ON telegram_users.chat_id = device_users.chat_id 
    JOIN devices ON device_users.device_id = devices.device_id 
WHERE devices.device_name = ?`
	rows, err := d.Db.Query(query, deviceName)
	if err != nil {
		return nil, err
	}

	// Convert to strings
	var users []common.TelegramUser
	for rows.Next() {
		var user sql.NullString
		var chatId sql.NullString
		err = rows.Scan(&chatId, &user)
		if err != nil {
			return nil, err
		}

		if user.Valid && chatId.Valid {
			users = append(users, common.TelegramUser{
				ChatId: chatId.String,
				User:   user.String,
			})
		}
	}

	return users, nil
}

func (d *Database) DeleteUser(deviceName string, chatId string) error {
	query := "DELETE FROM device_users WHERE device_id = (SELECT device_id FROM devices WHERE device_name = ?) AND chat_id = ?"
	_, err := d.Db.Exec(query, deviceName, chatId)
	return err
}

func (d *Database) AddUser(deviceName string, user common.TelegramUser) error {
	// add user to telegram_users replace into
	query := "INSERT INTO telegram_users (chat_id, user) VALUES (?, ?) ON CONFLICT (chat_id) DO UPDATE SET user = ?"
	_, err := d.Db.Exec(query, user.ChatId, user.User, user.User)
	if err != nil {
		return err
	}

	query = "INSERT OR REPLACE INTO device_users (device_id, chat_id) VALUES ((SELECT device_id FROM devices WHERE device_name = ?), ?)"
	_, err = d.Db.Exec(query, deviceName, user.ChatId)
	return err
}

func (d *Database) GetDevicesByUser(chatId string) ([]common.Device, error) {
	query := `
SELECT devices.device_name, devices.admin FROM devices
JOIN device_users du on devices.device_id = du.device_id
WHERE du.chat_id = ?`
	rows, err := d.Db.Query(query, chatId)
	if err != nil {
		return nil, err
	}

	var devices []common.Device
	for rows.Next() {
		var device common.Device
		var admin int
		err = rows.Scan(&device.DeviceName, &admin)
		if err != nil {
			return nil, err
		}
		device.Admin = admin == 1
		devices = append(devices, device)
	}

	return devices, nil
}
