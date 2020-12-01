package master

import (
	"database/sql"
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/ClickHouse/clickhouse-go"
	"time"
)

type MetadataClickHouse struct {
	db *sql.DB
}

func NewClickHouseMetaStore(config *config.Config) (c *MetadataClickHouse, err error) {
	c = new(MetadataClickHouse)
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
		CREATE TABLE IF NOT EXISTS mastermeta (
			vid UInt64,
			nid UInt64,
			filePath String,
			datetime DateTime
		) engine=MergeTree() 
		PARTITION BY toYYYYMM(datetime)
		ORDER BY (filePath)
	`); err != nil {
		fmt.Printf("Clickhouse create filemeta failed, err = %v", err)
	}
	return c, nil
}

func (c *MetadataClickHouse) GetEntries(filePath string) (Entries []*Entry, err error) {
	panic("implement me")
}

func (c *MetadataClickHouse) Set(filePath string, vid, nid uint64) error {
	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("InsertEntry: begin transaction error %v", err)
	}
	stmt, err := tx.Prepare("INSERT INTO mastermeta (vid, nid ,filePath, datetime) VALUES(?,?,?,?)")
	if err != nil {
		return fmt.Errorf("InsertEntry error: %s", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(vid, nid, filePath, time.Now().Format("2006-01-02 15:04:05")); err != nil {
		return fmt.Errorf("InsertEntry error: %s", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("InsertEntry: transaction commit error %s", err)
	}
	return nil
}

func (c *MetadataClickHouse) Delete(filePath string) error {
	if _, err := c.db.Exec(
		"ALTER TABLE mastermeta DELETE WHERE filePath = ?", filePath); err != nil {
		return fmt.Errorf("delete %s : %v", filePath, err)
	}
	return nil
}

func (c *MetadataClickHouse) Get(filePath string) (vid, nid uint64, err error) {
	rows, err := c.db.Query("SELECT vid, nid FROM mastermeta WHERE filePath=?", filePath)
	if err != nil {
		return 0,0, fmt.Errorf("select meta from filemeta where vid = %d and nid = %d err : %v",
			vid, nid, err)
	}
	defer rows.Close()
	var found bool
	for rows.Next() {
		found = true
		if err := rows.Scan(&vid, &nid); err != nil {
			return 0,0, err
		}
	}
	if !found {
		return 0,0, fmt.Errorf("vid and nid of filePath %s not found", filePath)
	}
	return vid, nid, nil
}

func (c *MetadataClickHouse) Close() error {
	return c.db.Close()
}

