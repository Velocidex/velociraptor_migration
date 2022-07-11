# A migration tool for older versions of Velociraptor

This tool allows migration of very old datastore to the newer formats.

Currently supported:

1. Old csv based result sets (circa Velociraptor 0.3.9)
2. Old disk baased indexes

This tool is experimental! not all data may be migrated properly.

## How to use

Point the tool at the datastore directory:

```
./velociraptor_migration_tool migrate -v /Velociraptor/datastore/
```
