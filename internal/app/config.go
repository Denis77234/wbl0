package app

import (
	"log"
	"os"

	"github.com/nats-io/nats.go"

	"awesomeProject/internal/postgres"
)

type Cfg struct {
	natsurl    string
	clusterId  string
	clientId   string
	serverport string
}

func makeConfigs() (Cfg, postgres.Config) {

	serport, ok := os.LookupEnv("SERVER_PORT")
	if !ok {
		log.Fatal("env variable SERVER_PORT not found")
	}

	url, ok := os.LookupEnv("NATS_URL")
	if !ok {
		log.Fatal("env variable NATS_URL not found")
	}

	if url == "default" {
		url = nats.DefaultURL
	}

	cluster, ok := os.LookupEnv("NATS_CLUSTER_ID")
	if !ok {
		log.Fatal("env variable NATS_CLUSTER_ID not found")
	}

	client, ok := os.LookupEnv("NATS_CLIENT_ID")
	if !ok {
		log.Fatal("env variable NATS_CLIENT_ID not found")
	}

	nc := Cfg{
		natsurl:    url,
		clusterId:  cluster,
		clientId:   client,
		serverport: serport,
	}

	login, ok := os.LookupEnv("PG_LOGIN")
	if !ok {
		log.Fatal("env variable PG_LOGIN not found")
	}

	pass, ok := os.LookupEnv("PG_PASSWORD")
	if !ok {
		log.Fatal("env variable PG_PASSWORD not found")
	}

	host, ok := os.LookupEnv("PG_HOST")
	if !ok {
		log.Fatal("env variable PG_HOST not found")
	}

	port, ok := os.LookupEnv("PG_PORT")
	if !ok {
		log.Fatal("env variable PG_PORT not found")
	}

	db, ok := os.LookupEnv("PG_DATABASE")
	if !ok {
		log.Fatal("env variable PG_DATABASE not found")
	}

	ssl, ok := os.LookupEnv("PG_SSL")
	if !ok {
		log.Fatal("env variable PG_SSL not found")
	}

	pc := postgres.Config{
		Login:    login,
		Password: pass,
		HostName: host,
		Port:     port,
		Database: db,
		Ssl:      ssl,
	}

	return nc, pc
}
