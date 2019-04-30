package storage

import (
	"fmt"
	"testing"
	"time"
)

func TestInsertShareRecord(t *testing.T) {
	s, err := NewSQLiteStorage("test.db")
	if err != nil {
		t.Fatal(err)
	}
	id := fmt.Sprintf("hash-%d", time.Now().Unix())
	ret, err := s.InsertShareRecord(id, "2", "3", 5)
	if err != nil {
		t.Fatal(err)
	}
	if !ret {
		t.Fatal("insert miner record result false")
	}
}

func TestQueryShareRecordById(t *testing.T) {
	s, err := NewSQLiteStorage("test.db")
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.FindShareRecordById("2")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("id %v\n", r.id)
	fmt.Printf("hash %v\n", r.fileHash)
	fmt.Printf("to %v\n", r.toWalletAddr)
	fmt.Printf("profit %v\n", r.profit)
	fmt.Printf("createdat %v\n", r.createdAt)
}

func TestIncreaseShareRecordProfit(t *testing.T) {
	s, err := NewSQLiteStorage("test.db")
	if err != nil {
		t.Fatal(err)
	}
	ret, err := s.IncreaseShareRecordProfit("2", "2", 3)
	if err != nil {
		t.Fatal(err)
	}
	if !ret {
		t.Fatal("insert miner record result false")
	}
}

func TestFineShareRecordsByCreatedAt(t *testing.T) {
	s, err := NewSQLiteStorage("test.db")
	if err != nil {
		t.Fatal(err)
	}

	rs, err := s.FineShareRecordsByCreatedAt(1555928050, 1555928185, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("result %v\n", rs)
}

func TestCountRecordByFileHash(t *testing.T) {
	s, err := NewSQLiteStorage("test.db")
	if err != nil {
		t.Fatal(err)
	}

	rs, err := s.CountRecordByFileHash("2")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("result %v\n", rs)
}

func TestSumRecordsProfit(t *testing.T) {
	s, err := NewSQLiteStorage("test.db")
	if err != nil {
		t.Fatal(err)
	}

	rs, err := s.SumRecordsProfit()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("result %v\n", rs)
}
