package bot

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

type Logger struct {
	logfile string
}

func NewLogger(logfile string) *Logger {
	return &Logger{logfile: logfile}
}

func (l *Logger) Logfile() string {
	return l.logfile
}

func (l *Logger) Log(new_log string) {
	fmt.Println(new_log)
}

func (l *Logger) GetLogs() []string {
	return []string{}
}

func (l *Logger) Output() {}

type LoggerBacktest struct {
	logfile string
	logs    []string
}

func NewLoggerBacktest(logfile string) *LoggerBacktest {
	return &LoggerBacktest{logfile: logfile}
}
func (l *LoggerBacktest) Logfile() string {
	return l.logfile
}

func (l *LoggerBacktest) Log(new_log string) {
	l.logs = append(l.logs, new_log)
}

func (l *LoggerBacktest) GetLogs() []string {
	return l.logs
}

func (l *LoggerBacktest) Output() {
	file, err := os.OpenFile(l.logfile, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)

	for _, l := range l.logs {
		writer.Write([]string{l})
	}
	writer.Flush()
}
