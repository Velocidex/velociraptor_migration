package main

import (
	"fmt"
	"strings"
	"time"

	api_proto "www.velocidex.com/golang/velociraptor/api/proto"
	"www.velocidex.com/golang/velociraptor/datastore"
	"www.velocidex.com/golang/velociraptor/paths"
	"www.velocidex.com/golang/vfilter/utils"
)

var (
	index_command = app.Command("index", "Migrate index data")
)

func doIndexMigration() error {
	config_obj, err := makeDefaultConfigLoader().
		WithRequiredFrontend().
		WithRequiredUser().LoadAndValidate()
	if err != nil {
		return fmt.Errorf("Unable to load config file: %w", err)
	}

	labels := make(map[string][]string)

	db, err := datastore.GetDB(config_obj)
	if err != nil {
		return err
	}

	terms, err := db.ListChildren(
		config_obj, paths.CLIENT_INDEX_URN_DEPRECATED)
	if err != nil {
		return err
	}

	for _, t := range terms {
		term := t.Base()
		if strings.HasPrefix(term, "host:") ||
			strings.HasPrefix(term, "c.") ||
			term == "all" {
			continue
		}

		clients, err := db.ListChildren(config_obj, t)
		if err != nil {
			continue
		}

		for _, c := range clients {
			client_id := c.Base()
			client_labels, _ := labels[client_id]
			if !utils.InString(&client_labels, term) {
				client_labels = append(client_labels, term)
			}
			labels[client_id] = client_labels
		}
	}

	for client_id, client_labels := range labels {
		record := &api_proto.ClientLabels{
			Timestamp: uint64(time.Now().UnixNano()),
			Label:     client_labels,
		}
		client_path_manager := paths.NewClientPathManager(client_id)

		db.SetSubject(config_obj, client_path_manager.Labels(), record)
	}

	return nil
}

func init() {
	command_handlers = append(command_handlers, func(command string) bool {
		switch command {
		case index_command.FullCommand():
			FatalIfError(index_command, doIndexMigration)

		default:
			return false
		}
		return true
	})
}
