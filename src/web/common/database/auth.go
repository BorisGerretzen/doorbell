package database

import (
	"database/sql"
	"errors"
	"web/common"
)

func (d *Database) Login(deviceName string, password string) (bool, error) {
	s := "SELECT password FROM devices WHERE device_name = ?"

	var hash sql.NullString
	err := d.Db.QueryRow(s, deviceName).Scan(&hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	if !hash.Valid {
		return false, nil
	}

	return common.VerifyPassword(password, hash.String), nil
}

func (d *Database) AddDevice(deviceName string, password string) error {
	hash, err := common.HashPassword(password)
	if err != nil {
		return err
	}

	s := "INSERT INTO devices (device_name, password) VALUES (?, ?)"
	_, err = d.Db.Exec(s, deviceName, hash)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) DeleteDevice(deviceName string) error {
	s := "DELETE FROM devices WHERE device_name = ?"
	_, err := d.Db.Exec(s, deviceName)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) IsAdmin(deviceName string) (bool, error) {
	s := "SELECT admin FROM devices WHERE device_name = ?"
	var admin sql.NullInt64
	err := d.Db.QueryRow(s, deviceName).Scan(&admin)
	if err != nil {
		return false, err
	}

	return admin.Valid && admin.Int64 == 1, nil
}

func (d *Database) UpdatePassword(deviceName string, password string) error {
	hash, err := common.HashPassword(password)
	if err != nil {
		return err
	}

	s := "UPDATE devices SET password = ? WHERE device_name = ?"
	_, err = d.Db.Exec(s, hash, deviceName)
	if err != nil {
		return err
	}

	return nil
}
