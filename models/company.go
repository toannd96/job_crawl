package models

import (
	"gopkg.in/mgo.v2/bson"
)

type Company struct {
	ID       bson.ObjectId     `bson:"_id,omitempty"`
	Name     string            `bson:"name,omitempty"`
	TaxInfo  map[string]string `bson:"tax_info,omitempty"`
	Business []Business        `bson:"business,omitempty"`
}

type Business struct {
	ID     string `bson:"id,omitempty"`
	Carees string `bson:"carees,omitempty"`
}
