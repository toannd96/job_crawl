package main

import (
	"go-crawl/database"
	"go-crawl/feeds"
	"go-crawl/handle"
	repoimpl "go-crawl/repository/repo_impl"
	"sync"
)

func main() {
	mg := &database.Mongo{}
	mg.CreateConn()

	handle := handle.Handle{
		Repo: repoimpl.NewRepo(mg),
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		feeds.Masothue(handle.Repo)
	}()

	go func() {
		defer wg.Done()
		feeds.JobStreet(handle.Repo)
	}()

	wg.Wait()
}
