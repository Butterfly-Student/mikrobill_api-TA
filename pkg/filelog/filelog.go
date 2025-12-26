// pkg/filelog/filelog.go
package filelog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const (
	MaxLogsPerLevel = 1000
	LogDir          = "logs"
)

// LogLevel represents the log level type
type LogLevel string

const (
	LevelInfo  LogLevel = "info"
	LevelError LogLevel = "error"
	LevelWarn  LogLevel = "warn"
	LevelFatal LogLevel = "fatal"
	LevelDebug LogLevel = "debug"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// FileLogger handles file-based logging with rotation
type FileLogger struct {
	mu       sync.Mutex
	logDir   string
	maxLines int
}

var (
	defaultLogger *FileLogger
	once          sync.Once
)

// getProjectRoot returns the project root directory
func getProjectRoot() string {
	// Get the directory of this source file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		// Fallback: use current working directory
		cwd, _ := os.Getwd()
		return cwd
	}

	// Navigate up from pkg/filelog/ to project root
	// Adjust the number of ".." based on your project structure
	dir := filepath.Dir(filename)           // pkg/filelog
	projectRoot := filepath.Join(dir, "..", "..") // up to root
	
	// Clean the path to resolve ".."
	projectRoot, _ = filepath.Abs(projectRoot)
	
	return projectRoot
}

// Init initializes the file logger
func Init() error {
	var err error
	once.Do(func() {
		projectRoot := getProjectRoot()
		logPath := filepath.Join(projectRoot, LogDir)
		
		defaultLogger = &FileLogger{
			logDir:   logPath,
			maxLines: MaxLogsPerLevel,
		}
		err = defaultLogger.ensureLogDir()
	})
	return err
}

// GetLogger returns the default file logger instance
func GetLogger() *FileLogger {
	if defaultLogger == nil {
		Init()
	}
	return defaultLogger
}

func (l *FileLogger) ensureLogDir() error {
	return os.MkdirAll(l.logDir, 0755)
}

func (l *FileLogger) getLogFilePath(level LogLevel) string {
	return filepath.Join(l.logDir, fmt.Sprintf("%s.log", level))
}

// writeLog writes a log entry to the appropriate file
func (l *FileLogger) writeLog(level LogLevel, msg string, fields map[string]interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   msg,
		Fields:    fields,
	}

	filePath := l.getLogFilePath(level)

	// Read existing logs
	logs, err := l.readLogsFromFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Append new log
	logs = append(logs, entry)

	// Trim to max lines (keep most recent)
	if len(logs) > l.maxLines {
		logs = logs[len(logs)-l.maxLines:]
	}

	// Write back to file
	return l.writeLogsToFile(filePath, logs)
}

func (l *FileLogger) readLogsFromFile(filePath string) ([]LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var logs []LogEntry
	scanner := bufio.NewScanner(file)
	// Increase buffer size for long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		var entry LogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue // Skip malformed lines
		}
		logs = append(logs, entry)
	}

	return logs, scanner.Err()
}

func (l *FileLogger) writeLogsToFile(filePath string, logs []LogEntry) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, log := range logs {
		data, err := json.Marshal(log)
		if err != nil {
			continue
		}
		writer.Write(data)
		writer.WriteString("\n")
	}

	return writer.Flush()
}

// Info logs an info message
func (l *FileLogger) Info(msg string, fields ...map[string]interface{}) {
	f := mergeFields(fields)
	l.writeLog(LevelInfo, msg, f)
}

// Error logs an error message
func (l *FileLogger) Error(msg string, fields ...map[string]interface{}) {
	f := mergeFields(fields)
	l.writeLog(LevelError, msg, f)
}

// Warn logs a warning message
func (l *FileLogger) Warn(msg string, fields ...map[string]interface{}) {
	f := mergeFields(fields)
	l.writeLog(LevelWarn, msg, f)
}

// Fatal logs a fatal message
func (l *FileLogger) Fatal(msg string, fields ...map[string]interface{}) {
	f := mergeFields(fields)
	l.writeLog(LevelFatal, msg, f)
}

// Debug logs a debug message
func (l *FileLogger) Debug(msg string, fields ...map[string]interface{}) {
	f := mergeFields(fields)
	l.writeLog(LevelDebug, msg, f)
}

func mergeFields(fields []map[string]interface{}) map[string]interface{} {
	if len(fields) == 0 {
		return nil
	}
	result := make(map[string]interface{})
	for _, f := range fields {
		for k, v := range f {
			result[k] = v
		}
	}
	return result
}

// Package-level convenience functions
func Info(msg string, fields ...map[string]interface{}) {
	GetLogger().Info(msg, fields...)
}

func Error(msg string, fields ...map[string]interface{}) {
	GetLogger().Error(msg, fields...)
}

func Warn(msg string, fields ...map[string]interface{}) {
	GetLogger().Warn(msg, fields...)
}

func Fatal(msg string, fields ...map[string]interface{}) {
	GetLogger().Fatal(msg, fields...)
}

func Debug(msg string, fields ...map[string]interface{}) {
	GetLogger().Debug(msg, fields...)
}

// GetLogs returns logs for a specific level with pagination
func (l *FileLogger) GetLogs(level LogLevel, limit, offset int) ([]LogEntry, int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	filePath := l.getLogFilePath(level)
	logs, err := l.readLogsFromFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []LogEntry{}, 0, nil
		}
		return nil, 0, err
	}

	total := len(logs)

	// Reverse to show newest first
	reversed := make([]LogEntry, len(logs))
	for i, j := 0, len(logs)-1; j >= 0; i, j = i+1, j-1 {
		reversed[i] = logs[j]
	}

	// Apply pagination
	if offset >= len(reversed) {
		return []LogEntry{}, total, nil
	}

	end := offset + limit
	if end > len(reversed) {
		end = len(reversed)
	}

	return reversed[offset:end], total, nil
}

// GetAllLevelLogs returns logs from all levels combined
func (l *FileLogger) GetAllLevelLogs(limit, offset int) ([]LogEntry, int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var allLogs []LogEntry
	levels := []LogLevel{LevelInfo, LevelError, LevelWarn, LevelFatal, LevelDebug}

	for _, level := range levels {
		filePath := l.getLogFilePath(level)
		logs, err := l.readLogsFromFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, 0, err
		}
		allLogs = append(allLogs, logs...)
	}

	// Sort by timestamp (newest first)
	sortLogsByTimestamp(allLogs)

	total := len(allLogs)

	// Apply pagination
	if offset >= len(allLogs) {
		return []LogEntry{}, total, nil
	}

	end := offset + limit
	if end > len(allLogs) {
		end = len(allLogs)
	}

	return allLogs[offset:end], total, nil
}

func sortLogsByTimestamp(logs []LogEntry) {
	// Simple bubble sort for simplicity (newest first)
	for i := 0; i < len(logs)-1; i++ {
		for j := 0; j < len(logs)-i-1; j++ {
			if logs[j].Timestamp < logs[j+1].Timestamp {
				logs[j], logs[j+1] = logs[j+1], logs[j]
			}
		}
	}
}

// GetStats returns log statistics
func (l *FileLogger) GetStats() (map[LogLevel]int, int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	stats := make(map[LogLevel]int)
	total := 0
	levels := []LogLevel{LevelInfo, LevelError, LevelWarn, LevelFatal, LevelDebug}

	for _, level := range levels {
		filePath := l.getLogFilePath(level)
		logs, err := l.readLogsFromFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				stats[level] = 0
				continue
			}
			return nil, 0, err
		}
		stats[level] = len(logs)
		total += len(logs)
	}

	return stats, total, nil
}

// GetLogFiles returns list of available log files
func (l *FileLogger) GetLogFiles() ([]string, error) {
	entries, err := os.ReadDir(l.logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".log" {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}