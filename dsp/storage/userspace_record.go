package storage

import (
	"fmt"
	"time"

	"github.com/saveio/themis/common/log"
)

type UserspaceOperation uint

const (
	UserspaceOperationNone UserspaceOperation = iota
	UserspaceOperationAdd
	UserspaceOperationRevoke
)

type UserspaceTransferType uint

const (
	TransferTypeNone UserspaceTransferType = iota
	TransferTypeIn
	TransferTypeOut
)

type UserspaceRecord struct {
	Id              string
	WalletAddr      string
	Size            uint64
	SizeOperation   UserspaceOperation
	Second          uint64
	SecondOperation UserspaceOperation
	Amount          uint64
	TransferType    UserspaceTransferType
	TotalSize       uint64
	ExpiredAt       uint64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (this *SQLiteStorage) InsertUserspaceRecord(id, walletAddr string, size uint64, sizeOp UserspaceOperation, second uint64, secondOp UserspaceOperation, amount uint64, transferType UserspaceTransferType) (bool, error) {
	totalSize, expiredAt := uint64(0), uint64(0)
	records, _ := this.SelectUserspaceRecordByWalletAddr(walletAddr, 0, 1)
	if len(records) > 0 {
		totalSize = records[0].TotalSize
		expiredAt = records[0].ExpiredAt
	}
	switch sizeOp {
	case UserspaceOperationAdd:
		totalSize += size
	case UserspaceOperationRevoke:
		if totalSize >= size {
			totalSize -= size
		}
	}

	switch secondOp {
	case UserspaceOperationAdd:
		if expiredAt == 0 {
			expiredAt = uint64(time.Now().Unix())
		}
		expiredAt += second
	case UserspaceOperationRevoke:
		if expiredAt >= second {
			expiredAt -= second
		} else {
			expiredAt = uint64(time.Now().Unix())
		}
	}

	sql := fmt.Sprintf("INSERT INTO %s (id, walletAddress, size, sizeOperation, second, secondOperation, amount, transferType, totalSize, expiredAt, createdAt, updatedAt) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", USERSPACE_RECORDS_TABLE_NAME)
	log.Debugf("insert into size %d second %d totalSize %d expired %d", size, second, totalSize, expiredAt)
	return this.Exec(sql, id, walletAddr, size, sizeOp, second, secondOp, amount, transferType, totalSize, expiredAt, time.Now(), time.Now())
}

// SelectUserspaceRecordByWalletAddr.
func (this *SQLiteStorage) SelectUserspaceRecordByWalletAddr(walletAddr string, offset, limit uint64) ([]*UserspaceRecord, error) {
	log.Debugf("SelectUserspaceRecordByWalletAddr offset %d limit %d", offset, limit)
	sql := fmt.Sprintf("SELECT * FROM %s WHERE walletAddress = ? ORDER BY createdAt DESC ", USERSPACE_RECORDS_TABLE_NAME)
	args := make([]interface{}, 0, 3)
	args = append(args, walletAddr)
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
	records := make([]*UserspaceRecord, 0)
	for rows.Next() {
		record := &UserspaceRecord{}
		err := rows.Scan(&record.Id, &record.WalletAddr, &record.Size, &record.SizeOperation, &record.Second, &record.SecondOperation, &record.Amount, &record.TransferType,
			&record.TotalSize, &record.ExpiredAt, &record.CreatedAt, &record.UpdatedAt)
		if err != nil {
			continue
		}
		records = append(records, record)
	}
	return records, nil
}
