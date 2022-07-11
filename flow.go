package main

import (
	"context"

	config_proto "www.velocidex.com/golang/velociraptor/config/proto"
	"www.velocidex.com/golang/velociraptor/datastore"
	"www.velocidex.com/golang/velociraptor/file_store"
	"www.velocidex.com/golang/velociraptor/file_store/api"
	"www.velocidex.com/golang/velociraptor/file_store/directory"
	flows_proto "www.velocidex.com/golang/velociraptor/flows/proto"
	"www.velocidex.com/golang/velociraptor/json"
	"www.velocidex.com/golang/velociraptor/logging"
	"www.velocidex.com/golang/velociraptor/paths"
	"www.velocidex.com/golang/velociraptor/result_sets"
)

func migrateFlow(
	ctx context.Context,
	config_obj *config_proto.Config,
	flow_id, client_id string) error {
	db, err := datastore.GetDB(config_obj)
	if err != nil {
		return err
	}

	collection_context := &flows_proto.ArtifactCollectorContext{}
	flow_path_manager := paths.NewFlowPathManager(client_id, flow_id)

	err = db.GetSubject(config_obj, flow_path_manager.Path(),
		collection_context)
	if err != nil {
		return err
	}

	collection_context.ClientId = client_id
	defer func() {
		db.SetSubject(config_obj, flow_path_manager.Path(),
			collection_context)
		json.Dump(collection_context)
	}()

	// Logs
	counter, _ := migrateFromCSV(ctx, config_obj,
		flow_path_manager.LogLegacy(),
		flow_path_manager.Log())
	collection_context.TotalLogs = uint64(counter)
	collection_context.TotalCollectedRows = 0

	// Uploads
	counter, _ = migrateFromCSV(ctx, config_obj,
		flow_path_manager.UploadMetadata().
			SetType(api.PATH_TYPE_FILESTORE_CSV),
		flow_path_manager.UploadMetadata())
	collection_context.TotalUploadedFiles = uint64(counter)

	// Artifact results.
	for _, artifact_with_result := range collection_context.ArtifactsWithResults {
		artifact_name, artifact_source := paths.SplitFullSourceName(artifact_with_result)
		var path_spec api.FSPathSpec

		if artifact_source != "" {
			path_spec = paths.CLIENTS_ROOT.AsFilestorePath().
				AddChild(client_id, "artifacts", artifact_name,
					flow_id, artifact_source).
				SetType(api.PATH_TYPE_FILESTORE_JSON)
		} else {
			path_spec = paths.CLIENTS_ROOT.AsFilestorePath().
				AddChild(client_id, "artifacts", artifact_name,
					flow_id).
				SetType(api.PATH_TYPE_FILESTORE_JSON)
		}

		counter, _ := migrateFromCSV(ctx, config_obj,
			path_spec.SetType(api.PATH_TYPE_FILESTORE_CSV),
			path_spec)
		collection_context.TotalCollectedRows += uint64(counter)

	}

	return nil
}

func migrateFromCSV(
	ctx context.Context,
	config_obj *config_proto.Config,
	csv_pathspec api.FSPathSpec,
	json_pathspec api.FSPathSpec) (row_count uint64, err error) {
	logger := logging.GetLogger(config_obj, &logging.FrontendComponent)
	logger.Info("Migrating %v to %v", csv_pathspec.AsClientPath(),
		json_pathspec.AsClientPath())

	file_store_factory := file_store.GetFileStore(config_obj)

	// Read the logs
	log_chan, err := directory.ReadRowsCSV(
		ctx, file_store_factory,
		csv_pathspec,
		0, 9999999999)
	if err != nil {
		return 0, err
	}

	rs_writer, err := result_sets.NewResultSetWriter(
		file_store_factory,
		json_pathspec,
		nil,
		nil,
		result_sets.TruncateMode)
	if err != nil {
		return 0, err
	}
	defer rs_writer.Close()

	var counter uint64
	for l := range log_chan {
		rs_writer.Write(l)
		counter++
	}

	return counter, nil
}
