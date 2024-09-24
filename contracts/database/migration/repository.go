package migration

type Repository interface {
	//// CreateRepository Create the migration repository data store.
	//CreateRepository()
	//// Delete Remove a migration from the log.
	//Delete(migration string)
	//// DeleteRepository Delete the migration repository data store.
	//DeleteRepository()
	//// GetLast Get the last migration batch.
	//GetLast()
	//// GetMigrationBatches Get the completed migrations with their batch numbers.
	//GetMigrationBatches()
	//// GetMigrations Get the list of migrations.
	//GetMigrations(steps int)
	//// GetMigrationsByBatch Get the list of the migrations by batch.
	//GetMigrationsByBatch(batch int)
	//// GetNextBatchNumber Get the next migration batch number.
	//GetNextBatchNumber()
	//// GetRan Get the completed migrations.
	//GetRan()
	//// Log that a migration was run.
	//Log(file, batch string)
	//// RepositoryExists Determine if the migration repository exists.
	//RepositoryExists()
}
