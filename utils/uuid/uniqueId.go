package uuid

import "time"

func UniqueId() (id uint64) {
	//TODO 分布式唯一自增id snowflake
	return uint64(time.Now().UnixNano())
}