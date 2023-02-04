package dbutils

// Видимость на уровне пакета dbutils
const employe = `
	CREATE TABLE IF NOT EXISTS employe (
           ID INTEGER PRIMARY KEY AUTOINCREMENT,
           FIO VARCHAR(64) NULL,
           DEPARTMENT VARCHAR(64) NULL,
           POSITION VARCHAR(64) NULL
        )
`

const events = `
	CREATE TABLE IF NOT EXISTS events (
          ID INTEGER PRIMARY KEY AUTOINCREMENT,
          ARRIVAL_TIME INTEGER NULL,
          LEAVING_TIME INTEGER NULL,
          EMPLOYE_ID,
          CONSTRAINT fk_employe
            FOREIGN KEY (EMPLOYE_ID) 
            REFERENCES employe(ID) ON DELETE CASCADE  
        )
`
