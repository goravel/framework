package migration

//func TestMigrateFreshCommand(t *testing.T) {
//	if env.IsWindows() {
//		t.Skip("Skipping tests of using docker")
//	}
//
//	testQueries := gorm.NewTestQueries().Queries()
//	for driver, testQuery := range testQueries {
//		query := testQuery.Query()
//		mockConfig := testQuery.MockConfig()
//		mockConfig.EXPECT().GetString("database.migrations.table").Return("migrations").Once()
//		mockConfig.EXPECT().GetString("database.migrations.driver").Return(contractsmigration.MigratorSql).Once()
//		mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.charset", testQuery.Docker().Driver().String())).Return("utf8bm4").Once()
//
//		mockSchema := mocksmigration.NewSchema(t)
//		migration.CreateTestMigrations(driver)
//
//		mockContext := mocksconsole.NewContext(t)
//		mockArtisan := mocksconsole.NewArtisan(t)
//
//		migrateCommand := NewMigrateCommand(nil, mockConfig, mockSchema)
//		require.NotNil(t, migrateCommand)
//		assert.Nil(t, migrateCommand.Handle(mockContext))
//
//		mockContext.EXPECT().OptionBool("seed").Return(false).Once()
//
//		migrateFreshCommand := NewMigrateFreshCommand(mockConfig, mockArtisan)
//		assert.Nil(t, migrateFreshCommand.Handle(mockContext))
//
//		var agent migration.Agent
//		err := query.Where("name", "goravel").First(&agent)
//		assert.Nil(t, err)
//		assert.True(t, agent.ID > 0)
//
//		// Test MigrateFreshCommand with --seed flag and seeders specified
//		mockContext = mocksconsole.NewContext(t)
//		mockConfig.EXPECT().GetString("database.migrations.table").Return("migrations").Once()
//		mockConfig.EXPECT().GetString("database.migrations.driver").Return(contractsmigration.MigratorSql).Once()
//
//		mockArtisan = mocksconsole.NewArtisan(t)
//		mockContext.EXPECT().OptionBool("seed").Return(true).Once()
//		mockContext.EXPECT().OptionSlice("seeder").Return([]string{"MockSeeder"}).Once()
//		mockArtisan.EXPECT().Call("db:seed --seeder MockSeeder").Once()
//
//		migrateFreshCommand = NewMigrateFreshCommand(mockConfig, mockArtisan)
//		assert.Nil(t, migrateFreshCommand.Handle(mockContext))
//
//		var agent1 migration.Agent
//		err = query.Where("name", "goravel").First(&agent1)
//		assert.Nil(t, err)
//		assert.True(t, agent1.ID > 0)
//
//		// Test MigrateFreshCommand with --seed flag and no seeders specified
//		mockContext = mocksconsole.NewContext(t)
//		mockArtisan = mocksconsole.NewArtisan(t)
//		mockContext.EXPECT().OptionBool("seed").Return(true).Once()
//		mockContext.EXPECT().OptionSlice("seeder").Return([]string{}).Once()
//		mockArtisan.EXPECT().Call("db:seed").Once()
//
//		migrateFreshCommand = NewMigrateFreshCommand(mockConfig, mockArtisan)
//		assert.Nil(t, migrateFreshCommand.Handle(mockContext))
//
//		var agent2 migration.Agent
//		err = query.Where("name", "goravel").First(&agent2)
//		assert.Nil(t, err)
//		assert.True(t, agent2.ID > 0)
//	}
//
//	defer assert.Nil(t, file.Remove("database"))
//}
