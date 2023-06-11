package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/hung0913208/go-algorithm/lib/container"
	//"github.com/hung0913208/go-algorithm/modules"
)

func main() {
	//modules.Init(map[string]bool{
	//	// @NOTE: this place is used to define database
	//	"planetscale": len(os.Getenv("PLANETSCALE_DATABASE")) > 0,
	//	"elephansql":  len(os.Getenv("ELEPHANSQL_DATABASE")) > 0,
	//	"supabase":    len(os.Getenv("SUPABASE_DATABASE")) > 0,
	//	"yugabyte":    len(os.Getenv("YUGABYTE_DATABASE")) > 0,
	//	"influxdb":    len(os.Getenv("INFLUXDB_URI")) > 0,
	//	"redislab":    len(os.Getenv("REDIS_URI")) > 0,
	//	"memcachier":  len(os.Getenv("MEMCACHIER_HOST")) > 0,
	//	"mariadb":     len(os.Getenv("MARIADB_DATABASE")) > 0,
	//	"mysql":       len(os.Getenv("MYSQL_DATABASE")) > 0,
	//	"excel":       false,

	//	// @NOTE: this place is used to define microservice
	//	"spawn": len(os.Getenv("SPAWN_DATABASE")) > 0,
	//})

	// @NOTE: wait and serve http requests
	httpServer := http.Server{
		Addr: ":8080",
	}

	idleConnectionsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP Server Shutdown Error: %v", err)
		}

		close(idleConnectionsClosed)
	}()

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe Error: %v", err)
	}

	// @NOTE: closed everything... bye :)
	<-idleConnectionsClosed

	container.Terminate("Bye bye :)", 0)
}
