package master

import (
	"gopkg.in/redis.v2"
)

type MetadataRedis struct {
	client *redis.Client
}

