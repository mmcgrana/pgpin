package main

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"github.com/jrallison/go-workers"
	"log"
	"net/url"
	"strings"
)

func QueueStart() {
	log.Printf("queue.start")
	u, err := url.Parse(ConfigRedisUrl)
	Must(err)
	server := u.Host
	password, _ := u.User.Password()
	database := strings.TrimLeft(u.Path, "/")
	pool := fmt.Sprintf("%d", ConfigRedisPoolSize)
	process := uuid.New()
	workers.Configure(map[string]string{
		"server":   server,
		"password": password,
		"database": database,
		"pool":     pool,
		"process":  process,
	})
	workers.Middleware = workers.NewMiddleware()
}
