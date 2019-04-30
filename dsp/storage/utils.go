package storage

import "fmt"

const SHARE_RECORDS_TABLE_NAME = "share_records"

const CreateShareRecords string = "CREATE TABLE IF NOT EXISTS share_records (id VARCHAR[255] NOT NULL PRIMARY KEY, fileHash VARCHAR[255] NOT NULL, downloader VARCHAR[255] NOT NULL, profit INTEGER NOT NULL, createdAt DATE, updatedAt DATE);"
const ScriptCreateTables string = `PRAGMA foreign_keys=off;
BEGIN TRANSACTION;
%s
COMMIT;
PRAGMA foreign_keys=on;
`

func GetCreateTables() string {
	sqlStmt := fmt.Sprintf(ScriptCreateTables, CreateShareRecords)
	return sqlStmt
}
