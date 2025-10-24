package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron-ui/server"
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
	var logFilename *output.File

	config.Init()

	logFilename = output.Init(config.GetPath("LOG_PATH"))
	logInstance = initializeLogger(logFilename.Filename)

	dbConfig := config.NewDatabaseConfig()

	dbPool, err := database.CreateDbPool(dbConfig, logInstance)
	if err != nil {
		logInstance.Error("SQL Server connected fail", err)
	}
	defer dbPool.Close()

	srzRep := repositories.NewSrzRepository(dbPool, logInstance)

	s, err := gocron.NewScheduler()
	if err != nil {
		logInstance.Error("Error: ", err)
		panic(err)
	}

	defer func() {
		_ = s.Shutdown()
		logInstance.Info("Служба остановлена")
	}()

	j, err := s.NewJob(
		/*
			gocron.DurationJob(
				30*time.Second,
			),*/
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(2, 30, 0),
			),
		),
		gocron.NewTask(
			func() {
				fileDate := logFilename.CurrentDate
				if fileDate != output.GetCurrentDate() {
					logInstance.Close()
					logFilename = output.Init(config.GetPath("LOG_PATH"))
					logInstance = initializeLogger(logFilename.Filename)
					logInstance.Info("Сервис в работе. Создан новый файл логов.")
				}
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

	go func() {
		newStrPort := strings.Replace(config.GetPort("PORT"), ":", "", 1)
		intPort, err := strconv.Atoi(newStrPort)
		if err != nil {
			logInstance.Error("Не возможно определить порт сервера: ", err)
		}
		srv := server.NewServer(s, intPort)
		// srv := server.NewServer(scheduler, 8080, server.WithTitle("My Custom Scheduler")) // with custom title if you want to customize the title of the UI (optional)
		log.Printf("GoCron UI available at http://localhost%s\n", config.GetPort("PORT"))
		log.Fatal(http.ListenAndServe(config.GetPort("PORT"), srv.Router))
	}()

	select {}

}

func GetOtkrep(rep *repositories.SrzRepository, logger *logger.Logger) {
	logger.Info("Получение списка МО из БД")
	moList, err := rep.GetMo()
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
	logger.Info(fmt.Sprintf("Получено записей о МО: %d\n", len(moList)))

	logger.Info("Получение данных об откреплениях из БД")
	data, err := rep.GetReport()
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
	logger.Info(fmt.Sprintf("Получено записей об откреплениях: %d\n", len(data)))

	var mo []models.Otkrep
	for _, v := range moList {
		mo = Filter(data, func(d models.Otkrep) bool {
			return d.LpuCode == v
		})
		logger.Info(fmt.Sprintf("Количество записей: %d\n", len(mo)))
		logger.Info("Экспорт данных в Excel")

		workdir, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		}
		expDirectory := path.Join(workdir, config.GetPath("EXPORT_PATH"))
		f, err := export.GenerateXLS(mo)
		if err != nil {
			log.Fatalf("Ошибка формирования файла Excel: %v", err)
		}
		file_name := fmt.Sprintf("%s/M%s_%s.xlsx", expDirectory, v, time.Now().Format("2006-01-02T15-04-05"))
		if err := f.SaveAs(file_name); err != nil {
			fmt.Println(err)
		}
		logger.Info(fmt.Sprintf("Сформирован файл: %s\n", file_name))
	}
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
