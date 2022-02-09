package models

import "gopkg.in/mgo.v2/bson"

type Recruitment struct {
	ID             bson.ObjectId `bson:"_id,omitempty"`
	Title          string        `bson:"title,omitempty"`
	Company        string        `bson:"company,omitempty"`
	NumberRecruits string        `bson:"number_recruits,omitempty"`
	WorkForm       string        `bson:"work_form,omitempty"`
	Sex            string        `bson:"sex,omitempty"`
	Rank           string        `bson:"rank,omitempty"`
	Experience     string        `bson:"experience,omitempty"`
	Location       string        `bson:"location,omitempty"`
	Address        []string        `bson:"address,omitempty"`
	JobKeyword     []string      `bson:"job_keyword,omitempty"`
	SkillKeyword   []string      `bson:"skill_keyword,omitempty"`
	Descript       string        `bson:"descript,omitempty"`
	Salary         string        `bson:"salary,omitempty"`
	Url            string        `bson:"url,omitempty"`
	Site           string        `bson:"site,omitempty"`
	CreatedAt      string        `bson:"created_at,omitempty"`
	JobDeadline    string        `bson:"job_deadline,omitempty"`
}
