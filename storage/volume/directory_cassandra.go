package volume

import "github.com/gocql/gocql"

type CassandraDirectory struct {
	cluster *gocql.ClusterConfig
	session *gocql.Session
}
