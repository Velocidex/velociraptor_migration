package main

import (
	"context"

	api_proto "www.velocidex.com/golang/velociraptor/api/proto"
	config_proto "www.velocidex.com/golang/velociraptor/config/proto"
	"www.velocidex.com/golang/velociraptor/datastore"
	"www.velocidex.com/golang/velociraptor/file_store/api"
	"www.velocidex.com/golang/velociraptor/paths"
)

func migrate_hunts(
	ctx context.Context,
	config_obj *config_proto.Config) error {
	db, err := datastore.GetDB(config_obj)
	if err != nil {
		return err
	}

	hunts, err := db.ListChildren(config_obj, paths.HUNTS_ROOT)
	if err != nil {
		return err
	}

	for _, h := range hunts {
		hunt_id := h.Base()

		hunt_path_manager := paths.NewHuntPathManager(hunt_id)
		hunt_obj := &api_proto.Hunt{}
		err = db.GetSubject(config_obj,
			hunt_path_manager.Path(), hunt_obj)
		if err != nil {
			return err
		}

		counter, _ := migrateFromCSV(ctx, config_obj,
			hunt_path_manager.Clients().
				SetType(api.PATH_TYPE_FILESTORE_CSV),
			hunt_path_manager.Clients())
		hunt_obj.Stats.TotalClientsScheduled = counter
		hunt_obj.Stats.TotalClientsWithResults = counter

		err = db.SetSubject(config_obj,
			hunt_path_manager.Path(), hunt_obj)
		if err != nil {
			return err
		}
	}

	return nil
}
