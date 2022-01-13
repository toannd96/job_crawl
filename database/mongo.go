package database

import (
	"gopkg.in/mgo.v2"
)

const (
	host   = "localhost"
	dbName = "tek4vn"
)

type Mongo struct {
	Db *mgo.Database
}

func (m *Mongo) CreateConn() {
	sess, err := mgo.Dial(host)
	if err != nil {
		panic(err)
	}
	m.Db = sess.DB(dbName)
}
