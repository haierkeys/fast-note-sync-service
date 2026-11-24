package main

import (
	"go/format"
	"os"
	"reflect"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/stretchr/testify/require"
)

func TestAutoMigrateSource(t *testing.T) {
	source := autoMigrateSource()
	// Preserve every generated query model, including the token models.
	queryType := reflect.TypeOf(query.Query{})
	for i := 0; i < queryType.NumField(); i++ {
		name := queryType.Field(i).Name
		if name != "db" {
			require.Contains(t, source, `case "`+name+`":`)
		}
	}
	// This manually maintained model must survive regeneration too.
	require.Contains(t, source, `case "UserOIDCIdentity":`)
	require.Contains(t, source, "return db.AutoMigrate(UserOIDCIdentity{})")
	generated, err := format.Source([]byte(source))
	require.NoError(t, err)
	checkedIn, err := os.ReadFile("../../internal/model/model.go")
	require.NoError(t, err)
	require.Equal(t, string(checkedIn), string(generated), "run go run ./cmd/model_gen from the repository root")
}
