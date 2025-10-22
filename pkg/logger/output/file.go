package output

import (
	"fmt"
	"os"
	"path"
	"time"
)

type file struct {
	LogDir      string
	CurrentFile string
	NewFile     string
}

func Init(logDir string) *file {
	workdir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	logDirectory := path.Join(workdir, logDir)
	return &file{
		LogDir:      logDirectory,
		CurrentFile: fmt.Sprintf("%s/otkrep_service_%s.log", logDirectory, time.Now().Format("2006-01-02")),
		NewFile:     fmt.Sprintf("%s/otkrep_service_%s.log", logDirectory, time.Now().Format("2006-01-02")),
	}
}

func (f *file) GetFileName() *file {
	return &file{
		LogDir:      f.LogDir,
		CurrentFile: f.CurrentFile,
		NewFile:     fmt.Sprintf("%s/otkrep_service_%s.log", f.LogDir, time.Now().Format("2006-01-02")),
	}
}
