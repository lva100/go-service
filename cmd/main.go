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

func initializeLogger(fname string) *logger.Logger {
	logInstance, err := logger.NewLogger(fname)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	return logInstance
}

func main() {
	var logInstance *logger.Logger
	var logFilename *output.File
	var lastInsertId int64

	config.Init()

	logFilename = output.Init(config.GetPath("LOG_PATH"))
	logInstance = initializeLogger(logFilename.Filename)

	dbConfig := config.NewDatabaseConfig()

	dbPool, err := database.CreateDbPool(dbConfig, logInstance)
	if err != nil {
		logInstance.Error("SQL Server connected fail", err)
	}
	defer func() {
		err := dbPool.Close()
		if err != nil {
			logInstance.Error("Closed dbPool error", err)
		}
	}()

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

	j1, err := s.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(20, 0, 0),
			),
		),
		gocron.NewTask(
			func(lastid int64) {
				fileDate := logFilename.CurrentDate
				if fileDate != output.GetCurrentDate() {
					logInstance.Close()
					logFilename = output.Init(config.GetPath("LOG_PATH"))
					logInstance = initializeLogger(logFilename.Filename)
					logInstance.Info("Сервис в работе. Создан новый файл логов.")
				}
				logInstance.Info("Старт задачи")
				id, ok := srzRep.CheckRequest(time.Now().Format("2006-01-02T15:04:05"))
				if !ok {
					logInstance.Info("Актульный запрос среза в БД не найден, отправляем запрос в ФЕРЗЛ")
					lastId, err := srzRep.CreateRequest(config.GetApiVersion())
					if err != nil {
						log.Fatalf("Server failed: %v\n", err)
					}
					lastInsertId = lastId
				} else {
					lastInsertId = id
				}
				logInstance.Info(fmt.Sprintf("ID запроса: %d\n", lastInsertId))
				logInstance.Info("Остановка задачи")
			},
			lastInsertId,
		),
		gocron.WithName("Запрос среза открепления из ФЕРЗЛ"),
	)
	if err != nil {
		log.Fatalf("Error %s", err)
	}

	j2, err := s.NewJob(
		/*gocron.DurationJob(
			30*time.Second,
		),*/
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(21, 0, 0),
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
				GetOtkrep(srzRep, logInstance, lastInsertId)
				logInstance.Info("Остановка задачи")
			},
		),
		gocron.WithName("Формирование списков открепленных для МО"),
	)
	if err != nil {
		log.Fatalf("Error %s", err)
	}

	go func() {
		logInstance.Info(fmt.Sprintf("Зарегистрирована задача с номером: %s\n", j1.ID()))
		logInstance.Info(fmt.Sprintf("Зарегистрирована задача с номером: %s\n", j2.ID()))
		s.Start()
	}()

	go func() {
		newStrPort := strings.Replace(config.GetPort(), ":", "", 1)
		intPort, err := strconv.Atoi(newStrPort)
		if err != nil {
			logInstance.Error("Не возможно определить порт сервера: ", err)
		}
		srv := server.NewServer(s, intPort, server.WithTitle("Go Service"))
		log.Printf("UI available at http://localhost%s\n", config.GetPort())
		log.Fatal(http.ListenAndServe(config.GetPort(), srv.Router))
	}()

	select {}

}

func GetOtkrep(rep *repositories.SrzRepository, logger *logger.Logger, lastInsertId int64) {
	logger.Info("Получение списка МО из БД")
	currentDate := time.Now().Format("20060102")
	moList, err := rep.GetMo(lastInsertId, currentDate)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
	if len(moList) == 0 {
		logger.Info("Нет данных, нечего выгружать")
		return
	}
	logger.Info(fmt.Sprintf("Получено записей о МО: %d\n", len(moList)))

	logger.Info("Получение данных об откреплениях из БД")
	data, err := rep.GetReport(lastInsertId, currentDate)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
	logger.Info(fmt.Sprintf("Получено записей об откреплениях: %d\n", len(data)))

	logger.Info("Загрузка истории отправки открепления в БД OtkrepLog")
	for _, v := range data {
		var err error
		pid, err := strconv.Atoi(v.PID)
		if err != nil {
			logger.Error("Ошибка конвертации PID из строки в число", err)
		}
		err = rep.InsertLog(pid, v.ENP, v.LpuCodeNew, v.LpuNameNew, v.LpuStart.Format("2006-01-02"),
			v.LpuCode, v.LpuFinish.Format("2006-01-02"), lastInsertId, time.Now().Format("2006-01-02"), fmt.Sprintf("M%s_%s.xlsx", v.LpuCode, time.Now().Format("2006-01-02")))
		if err != nil {
			logger.Error("Ошибка при вставке данных в БД OtkrepLog", err)
		}
	}
	logger.Info("Загрузка истории отправки открепления в БД OtkrepLog выполнена.")
	logger.Info("Закрытие прикрепления в БД по срезу из ФЕРЗЛ.")
	err = rep.ClosePrikrep(lastInsertId)
	if err != nil {
		logger.Error("Ошибка закрытия прикрепления в БД по срезу из ФЕРЗЛ", err)
	} else {
		logger.Info("Закрытие прикрепления в БД по срезу из ФЕРЗЛ выполнено.")
	}

	var mo []models.Otkrep
	for _, v := range moList {
		mo = Filter(data, func(d models.Otkrep) bool {
			return d.LpuCode == v
		})
		logger.Info(fmt.Sprintf("Количество записей для МО: %s - %d\n", v, len(mo)))
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
		// file_name := fmt.Sprintf("%s/M%s_%s.xlsx", expDirectory, v, time.Now().Format("2006-01-02T15-04-05"))
		file_name := fmt.Sprintf("%s/M%s_%s.xlsx", expDirectory, v, time.Now().Format("2006-01-02"))
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
