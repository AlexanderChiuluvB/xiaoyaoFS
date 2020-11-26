package volume

import (
	"database/sql"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/ClickHouse/clickhouse-go"

	"time"
)

type ClickHouseDirectory struct {
	db *sql.DB
}

func NewClickHouseDirectory(config *config.Config) (c *ClickHouseDirectory, err error) {
	c = new(ClickHouseDirectory)
	if c.db, err = sql.Open("clickhouse", config.ClickHouseHost); err != nil {
		fmt.Printf("connect to clickhouse %s error: %v", config.ClickHouseHost, err)
		return nil, err
	}
	if err := c.db.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			panic(err)
		}
	}
	if _, err = c.db.Exec(`
		CREATE TABLE IF NOT EXISTS filemeta (
			vid UInt64,
			nid UInt64,
			meta String,
			datetime DateTime
		) engine=MergeTree() 
		PARTITION BY toYYYYMM(datetime)
		ORDER BY (vid, nid)
	`); err != nil {
		fmt.Printf("Clickhouse create filemeta failed, err = %v", err)
	}
	return c, nil
}

func (c *ClickHouseDirectory) Get(vid, nid uint64) (n *Needle, err error) {
	var data []byte
	rows, err := c.db.Query("SELECT meta FROM filemeta WHERE vid=? AND nid=?", vid, nid)
	if err != nil {
		return nil, fmt.Errorf("select meta from filemeta where vid = %d and nid = %d err : %v",
			vid, nid, err)
	}
	defer rows.Close()
	var found bool
	for rows.Next() {
		found = true
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
	}
	if !found {
		return nil, fmt.Errorf("needle with vid %d and nid %d not found", vid, nid)
	}

	return UnMarshalBinary(data)
}

func (c *ClickHouseDirectory) Has(vid, nid uint64) (has bool) {
	_, err := c.Get(vid, nid)
	return err == nil
}

func (c *ClickHouseDirectory) Set(vid, nid uint64, needle *Needle) (err error) {
	meta, err := MarshalBinary(needle)
	if err != nil {
		return err
	}
	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("InsertEntry: begin transaction error %v", err)
	}
	stmt, err := tx.Prepare("INSERT INTO filemeta (vid, nid ,meta, datetime) VALUES(?,?,?,?)")
	if err != nil {
		return fmt.Errorf("InsertEntry error: %s", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(vid, nid, meta, time.Now().Format("2006-01-02 15:04:05")); err != nil {
		return fmt.Errorf("InsertEntry error: %s", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("InsertEntry: transaction commit error %s", err)
	}

	return nil
}

func (c *ClickHouseDirectory) Del(vid, nid uint64) (err error) {
	if _, err := c.db.Exec(
		"ALTER TABLE filemeta DELETE WHERE vid = ? AND nid = ?", vid, nid); err != nil {
		return fmt.Errorf("delete %d %d : %v", vid, nid, err)
	}
	return nil
}

func (c *ClickHouseDirectory) Close() {
	c.db.Close()
}