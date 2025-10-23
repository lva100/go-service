package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/lva100/go-service/config"
	database "github.com/lva100/go-service/internal/db"
	"github.com/lva100/go-service/internal/export"
	"github.com/lva100/go-service/internal/models"
	"github.com/lva100/go-service/internal/repositories"
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
	var logInstance *logger.Logger
	config.Init()
	outputLogs := output.Init(config.GetPort("LOG_PATH"))

	logInstance = initializeLogger(outputLogs.CurrentFile)
	logInstance.Info("Test record")
	dbConfig := config.NewDatabaseConfig()

	dbPool, err := database.CreateDbPool(dbConfig, logInstance)
	if err != nil {
		logInstance.Error("SQL Server connected fail", err)
	}
	defer dbPool.Close()

	srzRep := repositories.NewSrzRepository(dbPool, logInstance)

	// create a scheduler
	s, err := gocron.NewScheduler()
	if err != nil {
		logInstance.Error("Error: ", err)
		panic(err)
	}

	defer func() {
		_ = s.Shutdown()
		logInstance.Info("Служба остановлена")
	}()

	// add a job to the scheduler
	j, err := s.NewJob(
		gocron.DurationJob(
			30*time.Second,
		),
		gocron.NewTask(
			func() {
				logInstance.Info("Старт задачи")
				GetOtkrep(srzRep, logInstance)
				logInstance.Info("Остановка задачи")
			},
		),
	)
	if err != nil {
		log.Fatalf("Error %s", err)
	}

	go func() {
		logInstance.Info(fmt.Sprintf("Зарегистрирована задача с номером: %s\n", j.ID()))
		s.Start()
	}()

	select {}

}

func GetOtkrep(rep *repositories.SrzRepository, logger *logger.Logger) {
	logger.Info("Получение данных из БД о МО")
	moList, err := rep.GetMo()
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
	logger.Info(fmt.Sprintf("Получено записей из БД о МО: %d\n", len(moList)))

	logger.Info("Получение данных из БД")
	data, err := rep.GetReport()
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
	logger.Info(fmt.Sprintf("Получено записей из БД: %d\n", len(data)))

	var mo []models.Otkrep
	for _, v := range moList {
		mo = Filter(data, func(d models.Otkrep) bool {
			return d.LpuCode == v
		})
		logger.Info(fmt.Sprintf("Количество записей: %d\n", len(mo)))
		logger.Info("Экспорт данных в Excel")
		f, err := export.GenerateXLS(mo)
		if err != nil {
			log.Fatalf("Ошибка формирования файла Excel: %v", err)
		}
		file_name := fmt.Sprintf("M%s_%s.xlsx", v, time.Now().Format("2006-01-02T15-04-05"))
		if err := f.SaveAs(file_name); err != nil {
			fmt.Println(err)
		}
		logger.Info(fmt.Sprintf("Сформирован файл: %s\n", file_name))
	}
	// for _, v := range data {
	// 	fmt.Println(v.LpuNameNew)
	// }
	/*
		evens := Filter(nums, func(n int) bool {
			return n%2 == 0
		})
	*/
}

func Filter(slice []models.Otkrep, fn func(models.Otkrep) bool) []models.Otkrep {
	var result []models.Otkrep
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}
