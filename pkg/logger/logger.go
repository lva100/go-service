package logger

import (
	"log"
	"os"
)

type Logger struct {
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	file        *os.File
}

func NewLogger(logFilePath string) (*Logger, error) {
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		// infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		infoLogger:  log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLogger:  log.New(file, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		file:        file,
	}, nil
}

func (l *Logger) Info(msg string) {
	l.infoLogger.Printf("%s", msg)
}

func (l *Logger) Warn(msg string) {
	l.warnLogger.Printf("%s", msg)
}

func (l *Logger) Error(msg string, err error) {
	l.errorLogger.Printf("%s: %v", msg, err)
}

func (l *Logger) Close() {
	l.infoLogger.Printf("Запись в текущий файл логов завершена.")
	l.file.Close()
}
