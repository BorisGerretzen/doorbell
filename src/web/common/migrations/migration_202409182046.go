package migrations

type Migration202409182046 struct {
}

func (m *Migration202409182046) Version() int {
	return 202409182046
}

func (m *Migration202409182046) Description() string {
	return "Initial migration"
}

func (m *Migration202409182046) Up() []string {
	return []string{
		"CREATE TABLE devices (device_id INTEGER PRIMARY KEY, device_name TEXT UNIQUE COLLATE NOCASE, password TEXT, admin INTEGER DEFAULT 0, last_seen DATETIME DEFAULT CURRENT_TIMESTAMP)",
		"CREATE TABLE telegram_users (chat_id INTEGER PRIMARY KEY, user TEXT UNIQUE)",
		"CREATE TABLE device_users (device_id INTEGER REFERENCES devices(device_id) ON DELETE CASCADE, chat_id INTEGER REFERENCES telegram_users(chat_id) ON DELETE CASCADE, PRIMARY KEY (device_id, chat_id))",
		"INSERT INTO devices (device_id, device_name, password, admin) VALUES (1, 'admin', '$2a$14$xgJTN9afqctvqgBkMbRkHuXy6cDoZAqbC9KWHaAPVaiyJ8QfeFk2.', 1)",
	}
}

func (m *Migration202409182046) Down() []string {
	return []string{
		"DROP TABLE devices",
		"DROP TABLE telegram_users",
		"DROP TABLE device_users",
	}
}
