package repoimpl

import (
	"go-crawl/database"
	"go-crawl/repository"
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
