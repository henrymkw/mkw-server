package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var LogFile *os.File

func InitLogFile() {
	// Use os.Executable() to get the canotical path of the logs directory
	// relative paths are unreliable when the working directory is different
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		panic(err)
	}

	exeDir := filepath.Dir(exePath)
	logDir := filepath.Join(exeDir, "logs")

	t := time.Now().Format("2006-01-02_15-04-05")
	path := filepath.Join(logDir, t+".log")

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	LogFile = f
	log.SetOutput(LogFile)
}

func Log(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)

	if LogFile == nil {
		InitLogFile()
	}

	log.SetOutput(LogFile)
	log.Println(message)
}

func CloseLogFile() {
	if LogFile != nil {
		LogFile.Close()
	}
}
