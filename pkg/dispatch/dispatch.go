package dispatch

import "bufio"
import "fmt"
import "io/fs"
import "log/slog"
import "os"
import "path/filepath"
import "strconv"
import "strings"
import "time"

import "github.com/thierry-f-78/relp-log-collector/pkg/backend"
import "github.com/thierry-f-78/relp-log-collector/pkg/config"
import "github.com/thierry-f-78/relp-log-collector/pkg/utilities"

var notifyChan chan bool = make(chan bool)

func Notify() {
	select {
	case notifyChan <- true:
	default:
	}
}

type FileDesc struct {
	Path      string
	StartDate time.Time
	Logs      uint
}

func dispatcherCleanup() error {
	var entries []fs.DirEntry
	var err error
	var entry fs.DirEntry
	var name string

	entries, err = os.ReadDir(config.Cf.Spool.Path)
	if err != nil {
		return err
	}

	for _, entry = range entries {
		name = entry.Name()
		if strings.HasPrefix(name, ".") {
			err = os.Remove(filepath.Join(config.Cf.Spool.Path, name))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func Dispatch(inheritLog *slog.Logger) {
	var logFiles []*FileDesc
	var logFile *FileDesc
	var logsTotal uint
	var fileDesc *FileDesc
	var parts []string
	var value uint64
	var err error
	var entries []fs.DirEntry
	var entry fs.DirEntry
	var log *slog.Logger
	var lastProcessing time.Time
	var fileScanner *bufio.Scanner
	var logLine string
	var m *backend.Message
	var batchList *backend.BatchList
	var decodeError bool

	log = inheritLog.WithGroup("dispatcher")
	lastProcessing = time.Now()

	for {

		// Empty the spool of log files and reset number of logs
		logFiles = nil
		logsTotal = 0

		// Browse spool directory to find files to process.
		entries, err = os.ReadDir(config.Cf.Spool.Path)
		if err != nil {
			log.Error(fmt.Sprintf("Cannot read directory %q: %s", config.Cf.Spool.Path, err.Error()))
			entries = nil
		}
		for _, entry = range entries {

			// Prepare new log file descriptor.
			fileDesc = &FileDesc{
				Path: filepath.Join(config.Cf.Spool.Path, entry.Name()),
			}

			// Try to decode filename. If the filename do not match expected
			// skip the entry.
			parts = strings.Split(entry.Name(), ".")
			if len(parts) != 3 || parts[2] != "log" {
				continue
			}

			// Decode the first part of the file name which is timestamp
			// in microseconds. If we cannot decode this value, skip the file.
			value, err = strconv.ParseUint(parts[0], 16, 64)
			if err != nil {
				continue
			}
			fileDesc.StartDate = time.UnixMicro(int64(value))

			// Decode last part of the file which is the number of log lines
			// in the file. If we cannot decode this value, skip the file.
			value, err = strconv.ParseUint(parts[1], 16, 32)
			if err != nil {
				continue
			}
			fileDesc.Logs = uint(value)

			// Add the file descriptor to the list of files to process.
			// And increment total amount of log line to process.
			logFiles = append(logFiles, fileDesc)
			logsTotal += fileDesc.Logs

			// If we have enough logs to process, break the loop.
			if logsTotal >= config.Cf.Dispatch.MinLogs {
				break
			}
		}

		// If we do not have sufficient logs to process, wait
		// a little bit to scan again the spool directory.
		// If we reach the max wait time, process the current batch.
		if logsTotal < config.Cf.Dispatch.MinLogs && time.Since(lastProcessing) < config.Cf.Dispatch.MaxWait {
			time.Sleep(config.Cf.Dispatch.CheckInterval)
			continue
		}

		// Initialise all plugins to process new batch of logs.
		batchList, err = backend.NewBatch()
		if err != nil {
			log.Error(fmt.Sprintf("Can't init backend: %s. Try again in %s", err.Error(), config.Cf.Dispatch.MaxWait.String()))
			time.Sleep(config.Cf.Dispatch.MaxWait)
			continue
		}

		for _, logFile := range logFiles {
			fh, err := os.Open(logFile.Path)
			if err != nil {
				log.Error(fmt.Sprintf("Cannot open file %q: %s. Try again in %s", logFile.Path, err.Error(), config.Cf.Dispatch.MaxWait.String()))
				time.Sleep(config.Cf.Dispatch.MaxWait)
				continue
			}

			fileScanner = bufio.NewScanner(fh)
			fileScanner.Split(bufio.ScanLines)

			// Process each log line
			for fileScanner.Scan() {

				// Get next log line.
				logLine = fileScanner.Text()

				// Unescape encoded log, don't care about error : in error
				// case the function return the original string
				logLine, _ = utilities.UnescapeNonASCIIPrintable(logLine)

				// Decode syslog protocol
				m, err = utilities.DecodeSyslog([]byte(logLine))
				if err != nil {
					log.Info(fmt.Sprintf("Can't decode log line %q: %s", logLine, err.Error()))
					m = &backend.Message{
						Date: time.Now(),
						Data: logLine,
					}
					decodeError = true
				} else {
					decodeError = false
				}

				err = batchList.Pick(m, decodeError)
				if err != nil {
					log.Error(fmt.Sprintf("Can't add log %q in backend: %s", logLine, err.Error()))
					return
				}
			}

			err = fileScanner.Err()
			if err != nil {
				log.Error(fmt.Sprintf("Error reading file %q: %s. Try again in %s", logFile.Path, err.Error(), config.Cf.Dispatch.MaxWait.String()))
				fh.Close()
				time.Sleep(config.Cf.Dispatch.MaxWait)
				continue
			}

			fh.Close()
		}

		err = batchList.Flush()
		if err != nil {
			log.Error(fmt.Sprintf("Cannot flush data: %s. Try again in %s", err.Error(), config.Cf.Dispatch.MaxWait.String()))
			time.Sleep(config.Cf.Dispatch.MaxWait)
			continue
		}

		for _, logFile = range logFiles {
			err = os.Remove(logFile.Path)
			if err != nil {
				log.Error(fmt.Sprintf("Cannot remove file %q: %s", logFile.Path, err.Error()))
			}
		}

		lastProcessing = time.Now()
	}
}
