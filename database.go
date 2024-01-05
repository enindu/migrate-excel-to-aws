package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func createDatabaseConnection() *gorm.DB {
	dsn := DB_USERNAME + ":" + DB_PASSWORD + "@" + DB_PROTOCOL + "(" + DB_HOST + ")/" + DB_NAME + "?charset=" + DB_CHARSET + "&loc=" + DB_LOCALE + "&parseTime=" + DB_PARSE_TIME

	database, exception := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	handle(exception)

	exception = migrateDatabase(database)
	handle(exception)

	return database
}

func migrateDatabase(d *gorm.DB) error {
	if !d.Migrator().HasTable(&User{}) {
		return d.AutoMigrate(&User{})
	}

	if !d.Migrator().HasTable(&CandidateCvs{}) {
		return d.AutoMigrate(&CandidateCvs{})
	}

	return nil
}
