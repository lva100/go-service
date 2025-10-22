package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/lva100/go-service/config"
	database "github.com/lva100/go-service/internal/db"
	"github.com/lva100/go-service/pkg/logger"
	"github.com/lva100/go-service/pkg/logger/output"
)

func initializeLogger(fn string) *logger.Logger {
	logInstance, err := logger.NewLogger(fn)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	return logInstance
}

func main() {
	config.Init()
	outputLogs := output.Init(config.GetPort("LOG_PATH"))

	logInstance := initializeLogger(outputLogs.CurrentFile)
	logInstance.Info("Test record")
	dbConfig := config.NewDatabaseConfig()

	fmt.Println(outputLogs.CurrentFile)
	// servicePort := config.GetPort("PORT")

	dbPool, err := database.CreateDbPool(dbConfig, logInstance)
	if err != nil {
		logInstance.Error("SQL Server connected fail", err)
	}
	defer dbPool.Close()

	// r := chi.NewRouter()
	// r.Use(middleware.Logger)
	// r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("welcome"))
	// })
	// fmt.Printf("Service starting on port %s\n", servicePort)
	// http.ListenAndServe(servicePort, r)

	// create a scheduler
	s, err := gocron.NewScheduler()
	if err != nil {
		log.Fatalf("Error %s", err)
	}

	defer func() { _ = s.Shutdown() }()

	// add a job to the scheduler
	j, err := s.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func() {
				fmt.Println("Job working...")
			},
		),
	)
	if err != nil {
		log.Fatalf("Error %s", err)
	}
	// each job has a unique id
	fmt.Println(j.ID())

	// start the scheduler
	s.Start()
	time.Sleep(60 * time.Second)

}
