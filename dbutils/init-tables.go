package dbutils

import (
	"database/sql"
	"log"
)

func Initialize(dbDriver *sql.DB) {
	statement, driverError := dbDriver.Prepare(employe)
	if driverError != nil {
		log.Println(driverError)
	}
	// Create employe table
	_, statementError := statement.Exec()
	if statementError != nil {
		log.Println("Table employe already exists!")
	}
	// Create timings table
	statement, _ = dbDriver.Prepare(events)
	_, statementTimings := statement.Exec()
	if statementTimings != nil {
		log.Println("Table timings already exists!")
	}
	log.Println("All tables created/initialized successfully!")
}
