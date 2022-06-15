package relp

import "bufio"
import "fmt"
import "log/slog"
import "os"
import "path/filepath"

import "github.com/thierry-f-78/go-relp"

import "github.com/thierry-f-78/relp-log-collector/pkg/config"
import "github.com/thierry-f-78/relp-log-collector/pkg/utilities"

func flushSpool(spool []*relp.Message, log *slog.Logger) bool {
	var err error
	var outputName string
	var outputFile string
	var tempOutputFile string
	var fh *os.File
	var bufferFh *bufio.Writer
	var msg *relp.Message
	var outData []byte
	var handleError bool

	// Create unique file name which id prefixed by unique id derived from date
	// it ensure processing order and unicity. second parameter is number of logs
	// in this file. It allow fast accounting for total number of logs
	outputName = fmt.Sprintf("%016x", utilities.GetUniqueTime()) + "." + fmt.Sprintf("%08x", len(spool)) + ".log"

	// Create temporary file which exists while writing log duration. It just the real name
	// prefixed by ".". Note while the dot file existing, log are not acquited.
	tempOutputFile = filepath.Join(config.Cf.Spool.Path, "."+outputName)

	// Open file for write
	fh, err = os.Create(filepath.Clean(tempOutputFile))
	if err != nil {
		log.Error(fmt.Sprintf("can't open output log file %q: %s. close connection without ACK messages", tempOutputFile, err.Error()))
		return false
	}

	// Create buffer for buffered write because we will write some
	// values byte per byte.
	bufferFh = bufio.NewWriter(fh)

	// Process each message. To avoid confusion, syslog message is
	// cleaned: all non ASCII printable characters (and \) are escaped.
	// So, the \n is not a real separator, and we are sure of its role.
	for _, msg = range spool {
		outData = utilities.EscapeNonASCIIPrintable(msg.Data)
		_, err = bufferFh.Write(outData)
		if err != nil {
			log.Error(fmt.Sprintf("can't write to %q: %s. close connection without ACK messages", tempOutputFile, err.Error()))
			handleError = true
			break
		}
		err = bufferFh.WriteByte('\n')
		if err != nil {
			log.Error(fmt.Sprintf("can't write to %q: %s. close connection without ACK messages", tempOutputFile, err.Error()))
			handleError = true
			break
		}
	}

	// Flush to ensure all data were written on disk.
	err = bufferFh.Flush()
	if err != nil {
		log.Error(fmt.Sprintf("can't flush data in file %q: %s. close connection without ACK messages", tempOutputFile, err.Error()))
		handleError = true
		goto closeFileHandler
	}

	// Sync data on disk, to ensure the os will physically write data on disk.
	err = fh.Sync()
	if err != nil {
		log.Error(fmt.Sprintf("can't sync file %q: %s. close connection without ACK messages", tempOutputFile, err.Error()))
		handleError = true
	}

closeFileHandler:

	// Close connexion, on error, we could have a physicall on storage support,
	// so we prefer to close connection on error and without ACK messages
	err = fh.Close()
	if err != nil {
		log.Error(fmt.Sprintf("can't close file %q: %s. close connection without ACK messages", tempOutputFile, err.Error()))
		handleError = true
	}

	if handleError {
		return false
	}

	// Rename the temporary file to its final name in order to be processed
	// by the dispatcher.
	outputFile = filepath.Join(config.Cf.Spool.Path, outputName)
	err = os.Rename(tempOutputFile, outputFile)
	if err != nil {
		log.Error(fmt.Sprintf("can't rename spool file from %q to %q: %s. close connection without ACK messages", tempOutputFile, outputFile, err.Error()))
		return false
	}

	return true
}
