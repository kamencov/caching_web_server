package main

import (
	"caching_web_server/internal/apps"

	_ "github.com/lib/pq"
)

func main() {
	apps := apps.NewRun()

	if err := apps.Run(); err != nil {
		panic(err)
	}
}
