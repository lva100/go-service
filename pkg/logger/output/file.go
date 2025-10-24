package output

import (
	"fmt"
	"os"
	"path"
	"time"
)

type File struct {
	CurrentDate string
	LogDir      string
	Filename    string
}

func Init(logDir string) *File {
	workdir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	logDirectory := path.Join(workdir, logDir)
	return &File{
		CurrentDate: time.Now().Format("2006-01-02"),
		LogDir:      logDirectory,
		Filename:    fmt.Sprintf("%s/otkrep_service_%s.log", logDirectory, time.Now().Format("2006-01-02")),
	}
}

func GetCurrentDate() string {
	return time.Now().Format("2006-01-02")
}

func (f *File) GetCurrentFileName() string {
	return fmt.Sprintf("%s/otkrep_service_%s.log", f.LogDir, time.Now().Format("2006-01-02"))
}
