package handler

import (
	"fmt"
	"testing"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/notification"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/export"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestExportFlags(t *testing.T) {
	f := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(f)

	tmpDB1, dbErr1 := db.DB()
	if dbErr1 != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB1.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("happy code path", func(t *testing.T) {
		tmpDB := entity.NewTestDB()
		tmpDB2, dbErr2 := tmpDB.DB()
		if dbErr2 != nil {
			t.Errorf("Failed to get database")
		}

		defer tmpDB2.Close()

		exportFlags(tmpDB)
		tmpFlag := entity.Flag{}
		tmpDB.First(&tmpFlag)
		assert.NotZero(t, tmpFlag.ID)
	})

	t.Run("fetchAllFlags error code path", func(t *testing.T) {
		defer gostub.StubFunc(&fetchAllFlags, nil, fmt.Errorf("error")).Reset()
		tmpDB := entity.NewTestDB()
		tmpDB2, dbErr2 := tmpDB.DB()
		if dbErr2 != nil {
			t.Errorf("Failed to get database")
		}

		defer tmpDB2.Close()

		err := exportFlags(tmpDB)
		assert.Error(t, err)
	})
}

func TestExportFlagSnapshots(t *testing.T) {
	f := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(f)
	entity.SaveFlagSnapshot(db, f.ID, "flagr-test@example.com", notification.OperationUpdate, notification.ComponentFlag, f.ID, f.Key)

	tmpDB1, dbErr1 := db.DB()
	if dbErr1 != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB1.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("happy code path", func(t *testing.T) {
		tmpDB := entity.NewTestDB()
		tmpDB2, dbErr2 := tmpDB.DB()
		if dbErr2 != nil {
			t.Errorf("Failed to get database")
		}

		defer tmpDB2.Close()

		exportFlagSnapshots(tmpDB)
		fs := entity.FlagSnapshot{}
		tmpDB.First(&fs)
		assert.NotZero(t, fs.ID)
	})
}

func TestExportSQLiteFile(t *testing.T) {
	f := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(f)
	entity.SaveFlagSnapshot(db, f.ID, "flagr-test@example.com", notification.OperationUpdate, notification.ComponentFlag, f.ID, f.Key)

	tmpDB1, dbErr1 := db.DB()
	if dbErr1 != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB1.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("happy code path and export everything in db", func(t *testing.T) {
		f, done, err := exportSQLiteFile(nil)
		defer done()

		assert.NoError(t, err)
		assert.NotNil(t, f)
	})

	t.Run("happy code path and exclude_snapshots", func(t *testing.T) {
		f, done, err := exportSQLiteFile(new(true))
		defer done()

		assert.NoError(t, err)
		assert.NotNil(t, f)
	})
}

func TestExportSQLiteHandler(t *testing.T) {
	f := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(f)
	entity.SaveFlagSnapshot(db, f.ID, "flagr-test@example.com", notification.OperationUpdate, notification.ComponentFlag, f.ID, f.Key)

	tmpDB1, dbErr1 := db.DB()
	if dbErr1 != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB1.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	t.Run("happy code path", func(t *testing.T) {
		res := exportSQLiteHandler(export.GetExportSqliteParams{})
		assert.IsType(t, res.(*export.GetExportSqliteOK), res)
	})

	t.Run("fetchAllFlags error code path", func(t *testing.T) {
		defer gostub.StubFunc(&fetchAllFlags, nil, fmt.Errorf("error")).Reset()

		res := exportSQLiteHandler(export.GetExportSqliteParams{})
		assert.IsType(t, res.(*export.GetExportSqliteDefault), res)
	})
}

func TestExportEvalCacheJSONHandler(t *testing.T) {
	fixtureFlag := entity.GenFixtureFlag()
	db := entity.PopulateTestDB(fixtureFlag)
	tmpDB1, dbErr1 := db.DB()
	if dbErr1 != nil {
		t.Errorf("Failed to get database")
	}

	defer tmpDB1.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	ec := GetEvalCache()
	ec.lastSnapshotMaxID = 0
	ec.reloadMapCache()

	t.Run("happy code path", func(t *testing.T) {
		res := exportEvalCacheJSONHandler(export.GetExportEvalCacheJSONParams{})
		assert.IsType(t, res.(*export.GetExportEvalCacheJSONOK), res)
	})

	t.Run("query by id", func(t *testing.T) {
		res := exportEvalCacheJSONHandler(export.GetExportEvalCacheJSONParams{
			Ids: new("100"),
		})
		ok, ok2 := res.(*export.GetExportEvalCacheJSONOK)
		assert.True(t, ok2)
		payload := ok.Payload.(EvalCacheJSON)
		assert.Len(t, payload.Flags, 1)
		assert.Equal(t, "flag_key_100", payload.Flags[0].Key)
	})

	t.Run("query by id no match", func(t *testing.T) {
		res := exportEvalCacheJSONHandler(export.GetExportEvalCacheJSONParams{
			Ids: new("999"),
		})
		ok, ok2 := res.(*export.GetExportEvalCacheJSONOK)
		assert.True(t, ok2)
		payload := ok.Payload.(EvalCacheJSON)
		assert.Len(t, payload.Flags, 0)
	})

	t.Run("query by key", func(t *testing.T) {
		res := exportEvalCacheJSONHandler(export.GetExportEvalCacheJSONParams{
			Keys: new("flag_key_100"),
		})
		ok, ok2 := res.(*export.GetExportEvalCacheJSONOK)
		assert.True(t, ok2)
		payload := ok.Payload.(EvalCacheJSON)
		assert.Len(t, payload.Flags, 1)
		assert.Equal(t, uint(100), payload.Flags[0].ID)
	})

	t.Run("query by enabled true", func(t *testing.T) {
		res := exportEvalCacheJSONHandler(export.GetExportEvalCacheJSONParams{
			Enabled: new(true),
		})
		ok, ok2 := res.(*export.GetExportEvalCacheJSONOK)
		assert.True(t, ok2)
		payload := ok.Payload.(EvalCacheJSON)
		assert.Len(t, payload.Flags, 1)
		assert.True(t, payload.Flags[0].Enabled)
	})

	t.Run("query by enabled false", func(t *testing.T) {
		res := exportEvalCacheJSONHandler(export.GetExportEvalCacheJSONParams{
			Enabled: new(false),
		})
		ok, ok2 := res.(*export.GetExportEvalCacheJSONOK)
		assert.True(t, ok2)
		payload := ok.Payload.(EvalCacheJSON)
		assert.Len(t, payload.Flags, 0)
	})

	t.Run("query by tags ANY", func(t *testing.T) {
		res := exportEvalCacheJSONHandler(export.GetExportEvalCacheJSONParams{
			Tags: new("tag1"),
		})
		ok, ok2 := res.(*export.GetExportEvalCacheJSONOK)
		assert.True(t, ok2)
		payload := ok.Payload.(EvalCacheJSON)
		assert.Len(t, payload.Flags, 1)
	})

	t.Run("query by tags no match", func(t *testing.T) {
		res := exportEvalCacheJSONHandler(export.GetExportEvalCacheJSONParams{
			Tags: new("nonexistent"),
		})
		ok, ok2 := res.(*export.GetExportEvalCacheJSONOK)
		assert.True(t, ok2)
		payload := ok.Payload.(EvalCacheJSON)
		assert.Len(t, payload.Flags, 0)
	})

	t.Run("query by tags ALL match", func(t *testing.T) {
		res := exportEvalCacheJSONHandler(export.GetExportEvalCacheJSONParams{
			Tags: new("tag1,tag2"),
			All:  new(true),
		})
		ok, ok2 := res.(*export.GetExportEvalCacheJSONOK)
		assert.True(t, ok2)
		payload := ok.Payload.(EvalCacheJSON)
		assert.Len(t, payload.Flags, 1)
	})

	t.Run("query by tags ALL no match", func(t *testing.T) {
		res := exportEvalCacheJSONHandler(export.GetExportEvalCacheJSONParams{
			Tags: new("tag1,tag3"),
			All:  new(true),
		})
		ok, ok2 := res.(*export.GetExportEvalCacheJSONOK)
		assert.True(t, ok2)
		payload := ok.Payload.(EvalCacheJSON)
		assert.Len(t, payload.Flags, 0)
	})

	t.Run("query by enabled and tags", func(t *testing.T) {
		res := exportEvalCacheJSONHandler(export.GetExportEvalCacheJSONParams{
			Enabled: new(true),
			Tags:    new("tag1"),
		})
		ok, ok2 := res.(*export.GetExportEvalCacheJSONOK)
		assert.True(t, ok2)
		payload := ok.Payload.(EvalCacheJSON)
		assert.Len(t, payload.Flags, 1)
	})
}

func TestEvalCacheQuery_UnionIDsAndKeys(t *testing.T) {
	f1 := entity.GenFixtureFlag()
	f2 := f1
	f2.ID = 101
	f2.Key = "flag_key_101"
	f2.Enabled = false
	f2.Tags = []entity.Tag{{Value: "tag3"}}
	f2.PrepareEvaluation()

	db := entity.NewTestDB()
	db.Create(&f1)
	db.Create(&f2)
	tmpDB, dbErr := db.DB()
	if dbErr != nil {
		t.Errorf("Failed to get database")
	}
	defer tmpDB.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	ec := GetEvalCache()
	ec.lastSnapshotMaxID = 0
	ec.reloadMapCache()

	t.Run("ids + keys union", func(t *testing.T) {
		flags := ec.Query(ExportQuery{
			IDs:  []string{"100"},
			Keys: []string{"flag_key_101"},
		})
		assert.Len(t, flags, 2)
	})

	t.Run("enabled false with matching tags", func(t *testing.T) {
		flags := ec.Query(ExportQuery{
			Enabled: new(false),
			Tags:    []string{"tag3"},
		})
		assert.Len(t, flags, 1)
		assert.Equal(t, "flag_key_101", flags[0].Key)
	})

	t.Run("tags ANY partial match", func(t *testing.T) {
		flags := ec.Query(ExportQuery{
			Tags: []string{"tag1", "nonexistent"},
		})
		assert.Len(t, flags, 1)
		assert.Equal(t, "flag_key_100", flags[0].Key)
	})

	t.Run("tags ALL single tag match", func(t *testing.T) {
		flags := ec.Query(ExportQuery{
			Tags:   []string{"tag1"},
			TagsOp: TagOpAll,
		})
		assert.Len(t, flags, 1)
		assert.Equal(t, "flag_key_100", flags[0].Key)
	})

	t.Run("response body structure", func(t *testing.T) {
		res := exportEvalCacheJSONHandler(export.GetExportEvalCacheJSONParams{
			Ids: new("100"),
		})
		ok, ok2 := res.(*export.GetExportEvalCacheJSONOK)
		assert.True(t, ok2)
		payload := ok.Payload.(EvalCacheJSON)
		if !assert.Len(t, payload.Flags, 1) {
			return
		}
		f := payload.Flags[0]
		assert.Equal(t, uint(100), f.ID)
		assert.Equal(t, "flag_key_100", f.Key)
		assert.True(t, f.Enabled)
		assert.NotEmpty(t, f.Tags)
	})
}
