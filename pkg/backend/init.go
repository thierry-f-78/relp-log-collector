package backend

import "fmt"
import "io/fs"
import "log/slog"
import "os"
import "plugin"

import "github.com/thierry-f-78/relp-log-collector/pkg/config"

type backendRef struct {
	newBatch func() (Batch, error)
}

var backendRefList []*backendRef
var backendRefError *backendRef

var db transaction

func Init(inheritLog *slog.Logger) error {
	var err error
	var pluginDirectory string
	var dirIndex []fs.DirEntry
	var entry fs.DirEntry
	var back *plugin.Plugin
	var init plugin.Symbol
	var newBatch plugin.Symbol
	var registerBackend *backendRef
	var ok bool
	var pluginPath string
	var initFn InitFunction

	// Browse each configured plugin directory to load all plugin avalaible.
	for _, pluginDirectory = range config.Cf.Plugins.Path {

		// Browse plugin directory and load each files found as plugin
		dirIndex, err = os.ReadDir(pluginDirectory)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can't open plugin directory %q: %s\n", pluginDirectory, err.Error())
			continue
		}
		for _, entry = range dirIndex {
			if entry.IsDir() {
				continue
			}

			pluginPath = pluginDirectory + "/" + entry.Name()

			// Open plugin
			back, err = plugin.Open(pluginPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Can't open plugin %q: %s\n", pluginPath, err.Error())
				continue
			}

			// Check if plugin implements Backend interface
			//
			// Init() error
			// NewBatch() (backend.Batch, error)
			// Pick(m *backend.Message) (bool, error)
			// Flush() error
			registerBackend = &backendRef{}
			init, err = back.Lookup("Init")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Plugin %q has an error with symbol Init: %s. Ignore the plugin\n", pluginPath, err.Error())
				continue
			}
			initFn, ok = init.(func() error)
			if !ok {
				fmt.Fprintf(os.Stderr, "Plugin %q: symbol Init doesn't match the expected type. Ignore the plugin\n", pluginPath)
				continue
			}

			newBatch, err = back.Lookup("NewBatch")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Plugin %q has an error with symbol NewBatch: %s. Ignore the plugin\n", pluginPath, err.Error())
				continue
			}
			registerBackend.newBatch, ok = newBatch.(func() (Batch, error))
			if !ok {
				fmt.Fprintf(os.Stderr, "Plugin %q: symbol NewBatch doesn't match the expected type. Ignore the plugin\n", pluginPath)
				continue
			}

			// Init plugin
			err = initFn()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Can't init plugin %q: %s. Ignore the plugin\n", pluginPath, err.Error())
				continue
			}

			fmt.Fprintf(os.Stdout, "Plugin %q loaded\n", pluginPath)
			backendRefList = append(backendRefList, registerBackend)
		}
	}

	// Register internal logs proc. syslog is always at the final position to catch all logs
	// errors is specif to handle unparsable logs
	backendRefList = append(backendRefList, syslogInit())
	backendRefError = errorsInit()

	// Init database connection. If more than one database is configured, only
	// the first in the below list is initialized.
	switch {
	case config.Cf.Clickhouse != nil:
		fmt.Printf("Init Clickhouse\n")
		db, err = clickhouseInit(inheritLog)
	case config.Cf.PostgreSQL != nil:
		fmt.Printf("Init PostgreSQL\n")
		db, err = postgresInit(inheritLog)
	case config.Cf.SQLite != nil:
		fmt.Printf("Init SQLite\n")
		db, err = sqliteInit(inheritLog)
	}
	if err != nil {
		return err
	}

	return nil
}
