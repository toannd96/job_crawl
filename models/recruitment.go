package models

import "gopkg.in/mgo.v2/bson"

type Recruitment struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	Title     string        `bson:"title,omitempty"`
	Company   string        `bson:"company,omitempty"`
	Location  string        `bson:"location,omitempty"`
	Descript  string        `bson:"descript,omitempty"`
	Url       string        `bson:"url,omitempty"`
	Site      string        `bson:"site,omitempty"`
	CreatedAt string        `bson:"created_at,omitempty"`
}
