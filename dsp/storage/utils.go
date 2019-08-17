package storage

import "fmt"

const SHARE_RECORDS_TABLE_NAME = "share_records"
const USERSPACE_RECORDS_TABLE_NAME = "userspace_records"

const CreateShareRecords string = "CREATE TABLE IF NOT EXISTS share_records (id VARCHAR[255] NOT NULL PRIMARY KEY, fileHash VARCHAR[255] NOT NULL, fileName VARCHAR[255] NOT NULL, fileOwner VARCHAR[255] NOT NULL, downloader VARCHAR[255] NOT NULL, profit INTEGER NOT NULL, createdAt DATE, updatedAt DATE);"
const CreateUserspaceRecords string = "CREATE TABLE IF NOT EXISTS userspace_records (id VARCHAR[255] NOT NULL PRIMARY KEY, walletAddress VARCHAR[255] NOT NULL,  size INTEGER, sizeOperation INTEGER NOT NULL, second INTEGER, secondOperation INTEGER NOT NULL, amount INTEGER NOT NULL, transferType INTEGER NOT NULL, totalSize INTEGER NOT NULL, expiredAt INTEGER NOT NULL, createdAt DATE, updatedAt DATE);"
const ScriptCreateTables string = `PRAGMA foreign_keys=off;
BEGIN TRANSACTION;
%s%s
COMMIT;
PRAGMA foreign_keys=on;
`

func GetCreateTables() string {
	sqlStmt := fmt.Sprintf(ScriptCreateTables, CreateShareRecords, CreateUserspaceRecords)
	return sqlStmt
}
