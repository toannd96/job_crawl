package main

import (
	"go-crawl/database"
	"go-crawl/feeds"
	"go-crawl/handle"
	repoimpl "go-crawl/repository/repo_impl"
	"sync"
	"time"
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

	// Schedule crawl
	go schedule(24*time.Hour, handle, 1)
	schedule(24*time.Hour, handle, 2)
}

func schedule(timeSchedule time.Duration, handle handle.Handle, inndex int) {
	ticker := time.NewTicker(timeSchedule)
	func() {
		for {
			switch inndex {
			case 1:
				<-ticker.C
				feeds.Masothue(handle.Repo)
			case 2:
				<-ticker.C
				feeds.JobStreet(handle.Repo)
			}
		}
	}()
}
