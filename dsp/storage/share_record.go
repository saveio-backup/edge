package storage

import (
	"fmt"
	"time"

	"github.com/saveio/themis/common/log"
)

type ShareRecord struct {
	Id           string
	FileName     string
	FileOwner    string
	FileHash     string
	ToWalletAddr string
	Profit       uint64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// InsertShareRecord. insert a new miner_record or replace it
func (this *SQLiteStorage) InsertShareRecord(id, fileHash, fileName, fileOwner, toWalletAddr string, profit uint64) (bool, error) {
	sql := fmt.Sprintf("INSERT OR REPLACE INTO %s (id, fileHash, fileName, fileOwner,downloader, profit, createdAt, updatedAt) VALUES(?, ?, ?, ?, ?, ?, ?, ?)", SHARE_RECORDS_TABLE_NAME)
	return this.Exec(sql, id, fileHash, fileName, fileOwner, toWalletAddr, profit, time.Now(), time.Now())
}

// IncreaseShareRecordProfit. increase miner profit by increment
func (this *SQLiteStorage) IncreaseShareRecordProfit(id, idPrefix string, added uint64) (bool, error) {
	sql := ""
	if len(idPrefix) > 0 {
		sql = fmt.Sprintf("UPDATE %s SET profit = profit + ?, updatedAt = ? WHERE id IN (SELECT id FROM %s WHERE id like '%s-%%'  ORDER BY createdAt DESC LIMIT 1)", SHARE_RECORDS_TABLE_NAME, SHARE_RECORDS_TABLE_NAME, idPrefix)
	} else if len(id) > 0 {
		sql = fmt.Sprintf("UPDATE  %s SET profit = profit + ?, updatedAt = ? where id = ?", SHARE_RECORDS_TABLE_NAME)
	}
	log.Debugf("increase profit %s, added %d, now %v", sql, added, time.Now())
	return this.Exec(sql, added, time.Now())
}

// FindShareRecordById. find miner record by id
func (this *SQLiteStorage) FindShareRecordById(id string) (*ShareRecord, error) {
	// SELECT * FROM share_records  WHERE id like 'hash-%'  ORDER BY createdAt DESC LIMIT 1;
	// sql := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", SHARE_RECORDS_TABLE_NAME)
	sql := fmt.Sprintf("SELECT * FROM %s  WHERE id like '%%?%%'  ORDER BY createdAt DESC LIMIT 1", SHARE_RECORDS_TABLE_NAME)
	rows, err := this.Query(sql, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	record := &ShareRecord{}
	for rows.Next() {
		rows.Scan(&record.Id, &record.FileHash, &record.FileName, &record.FileOwner, &record.ToWalletAddr, &record.Profit, &record.CreatedAt, &record.UpdatedAt)
		break
	}
	return record, nil
}

// FineShareRecordsByCreatedAt. find miner record by createdat interval
func (this *SQLiteStorage) FineShareRecordsByCreatedAt(beginedAt, endedAt, offset, limit int64) ([]*ShareRecord, error) {
	sql := fmt.Sprintf("SELECT * FROM %s WHERE createdAt >= ? and createdAt <= ? ", SHARE_RECORDS_TABLE_NAME)
	args := make([]interface{}, 0, 4)
	beginT := time.Unix(beginedAt, 0)
	endT := time.Unix(endedAt, 0)
	args = append(args, beginT)
	args = append(args, endT)
	if limit != 0 {
		sql += "LIMIT ? "
		args = append(args, limit)
	}
	if offset != 0 {
		sql += "OFFSET ? "
		args = append(args, offset)
	}
	rows, err := this.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := make([]*ShareRecord, 0)
	for rows.Next() {
		record := &ShareRecord{}
		err := rows.Scan(&record.Id, &record.FileHash, &record.FileName, &record.FileOwner, &record.ToWalletAddr, &record.Profit, &record.CreatedAt, &record.UpdatedAt)
		if err != nil {
			continue
		}
		records = append(records, record)
	}
	return records, nil
}

func (this *SQLiteStorage) FindLastShareTime(fileHash string) (uint64, error) {
	sql := fmt.Sprintf("SELECT createdAt FROM %s WHERE fileHash = ? ORDER BY 'createdAt' DESC LIMIT 1", SHARE_RECORDS_TABLE_NAME)
	rows, err := this.Query(sql, fileHash)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var lastTime time.Time
	for rows.Next() {
		err := rows.Scan(&lastTime)
		if err != nil {
			return 0, err
		}
		break
	}
	log.Debugf("fileHash %s lastTime :%v unix %v", fileHash, lastTime, lastTime.Unix())
	if lastTime.Unix() < 0 {
		return 0, nil
	}

	return uint64(lastTime.Unix()), nil
}

func (this *SQLiteStorage) CountRecordByFileHash(fileHash string) (uint64, error) {
	sql := fmt.Sprintf("SELECT COUNT (fileHash) FROM %s WHERE fileHash = ? and profit > 0", SHARE_RECORDS_TABLE_NAME)
	rows, err := this.Query(sql, fileHash)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	count := uint64(0)
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return 0, err
		}
		break
	}
	return count, nil
}

// SumRecordsProfit. sum profit off all files
func (this *SQLiteStorage) SumRecordsProfit() (int64, error) {
	sql := fmt.Sprintf("SELECT SUM (profit) FROM %s;", SHARE_RECORDS_TABLE_NAME)
	rows, err := this.Query(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var value interface{}
	for rows.Next() {
		err := rows.Scan(&value)
		if err != nil {
			return 0, err
		}
		break
	}
	count, _ := value.(int64)
	return count, nil
}

// SumRecordsProfitByFileHash. sum profit by one files
func (this *SQLiteStorage) SumRecordsProfitByFileHash(fileHashStr string) (uint64, error) {
	sql := fmt.Sprintf("SELECT SUM (profit) FROM %s WHERE fileHash = ?", SHARE_RECORDS_TABLE_NAME)
	rows, err := this.Query(sql, fileHashStr)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	count := uint64(0)
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return 0, err
		}
		break
	}
	return count, nil
}
