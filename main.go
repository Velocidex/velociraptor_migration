package main

import (
	"os"

	"github.com/alecthomas/kingpin"
	"www.velocidex.com/golang/velociraptor/config"
)

var (
	app = kingpin.New("velociraptor_migration",
		"Migrate old datastores to new format.")

	config_path = app.Flag("config", "The configuration file.").
			Short('c').String()

	verbose_flag = app.Flag(
		"verbose", "Enabled verbose logging.").Short('v').
		Default("false").Bool()

	command_handlers []CommandHandler
)

type CommandHandler func(command string) bool

func main() {
	app.HelpFlag.Short('h')
	app.UsageTemplate(kingpin.CompactUsageTemplate)
	args := os.Args[1:]

	command := kingpin.MustParse(app.Parse(args))

	for _, command_handler := range command_handlers {
		if command_handler(command) {
			break
		}
	}
}

func makeDefaultConfigLoader() *config.Loader {
	return new(config.Loader).
		WithVerbose(*verbose_flag).
		WithFileLoader(*config_path).
		WithCustomValidator(initDebugServer)
}

func FatalIfError(command *kingpin.CmdClause, cb func() error) {
	err := cb()
	kingpin.FatalIfError(err, command.FullCommand())
}
