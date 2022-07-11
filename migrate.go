package main

import (
	"fmt"

	"www.velocidex.com/golang/velociraptor/config"
	"www.velocidex.com/golang/velociraptor/datastore"
	"www.velocidex.com/golang/velociraptor/paths"

	_ "www.velocidex.com/golang/velociraptor/result_sets/simple"
	_ "www.velocidex.com/golang/velociraptor/result_sets/timed"
)

var (
	migrate_command = app.Command("migrate", "Migrate data")

	migrate_command_datastore = migrate_command.
					Arg("datastore", "Path to the datastore to migrate").
					Required().String()
)

func doMigration() error {
	config_obj := config.GetDefaultConfig()
	config_obj.Datastore.Implementation = "FileBaseDataStore"
	config_obj.Datastore.Location = *migrate_command_datastore
	config_obj.Datastore.FilestoreDirectory = *migrate_command_datastore

	ctx, cancel := install_sig_handler()
	defer cancel()

	db, err := datastore.GetDB(config_obj)
	clients, err := db.ListChildren(config_obj, paths.CLIENTS_ROOT)
	if err != nil {
		return err
	}

	// Migrate hunts
	err = migrate_hunts(ctx, config_obj)
	if err != nil {
		return err
	}

	for _, c := range clients {
		client_id := c.Base()
		client_path_manager := paths.NewClientPathManager(client_id)
		flows, err := db.ListChildren(config_obj, client_path_manager.Path().AddChild("collections"))
		if err != nil {
			continue
		}

		fmt.Printf("Migrating client %v\n", client_id)

		for _, f := range flows {
			flow_id := f.Base()
			if flow_id == "F.Monitoring" {
				continue
			}

			fmt.Printf("Migrating client_id %v flow_id %v\n", client_id, flow_id)
			err = migrateFlow(ctx, config_obj, flow_id, client_id)
			if err != nil {
				return err
			}
		}

		// Allow interrupt cancellations
		select {
		case <-ctx.Done():
			return nil
		default:
		}

	}

	return nil
}

func init() {
	command_handlers = append(command_handlers, func(command string) bool {
		switch command {
		case migrate_command.FullCommand():
			FatalIfError(migrate_command, doMigration)

		default:
			return false
		}
		return true
	})
}
