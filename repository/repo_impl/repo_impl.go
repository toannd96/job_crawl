package repoimpl

import (
	"go-crawl/database"
	"go-crawl/repository"

	"gopkg.in/mgo.v2/bson"
)

type RepoImpl struct {
	mg *database.Mongo
}

func NewRepo(mg *database.Mongo) repository.Repository {
	return &RepoImpl{
		mg: mg,
	}
}

func (rp *RepoImpl) Save(data interface{}, collection string) error {
	return rp.mg.Db.C(collection).Insert(data)
}

func (rp *RepoImpl) FindByUrl(url string, collection string) (int, error) {
	count, err := rp.mg.Db.C(collection).Find(bson.M{"url": url}).Count()
	if err != nil {
		return 0, err
	}
	return count, nil
}
