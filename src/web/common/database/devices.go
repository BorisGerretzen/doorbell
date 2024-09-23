package database

import "web/common"

func (d *Database) GetDevices() ([]common.Device, error) {
	query := "SELECT device_name, admin FROM devices"
	rows, err := d.Db.Query(query)
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

func (d *Database) UpdateLastSeen(deviceName string) error {
	s := "UPDATE devices SET last_seen = datetime('now') WHERE device_name = ?"
	_, err := d.Db.Exec(s, deviceName)
	if err != nil {
		return err
	}

	return nil
}
