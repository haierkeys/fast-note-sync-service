package model

import (
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAutoMigrateUserOIDCIdentity(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "migration.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })

	require.NoError(t, AutoMigrate(db, "User"))
	user := User{UID: 1, Username: "test-user", Email: "user@example.invalid"}
	require.NoError(t, db.Create(&user).Error)
	require.False(t, db.Migrator().HasTable(&UserOIDCIdentity{}))
	require.NoError(t, AutoMigrate(db, "UserOIDCIdentity"))
	require.True(t, db.Migrator().HasTable(&UserOIDCIdentity{}))
	require.True(t, db.Migrator().HasIndex(&UserOIDCIdentity{}, "idx_user_oidc_identity_uid"))
	require.True(t, db.Migrator().HasIndex(&UserOIDCIdentity{}, "idx_user_oidc_identity_issuer_subject"))

	identity := UserOIDCIdentity{UID: user.UID, Issuer: "https://idp.example.invalid", Subject: "test-subject"}
	require.NoError(t, db.Create(&identity).Error)
	for range 2 {
		require.NoError(t, AutoMigrate(db, "User"))
		require.NoError(t, AutoMigrate(db, "UserOIDCIdentity"))
	}
	var savedUser User
	require.NoError(t, db.First(&savedUser, "uid = ?", user.UID).Error)
	require.Equal(t, user, savedUser)
	var savedIdentity UserOIDCIdentity
	require.NoError(t, db.First(&savedIdentity, identity.ID).Error)
	require.Equal(t, identity, savedIdentity)

	// The binding is unique by issuer AND subject, not by either alone.
	duplicate := UserOIDCIdentity{UID: 2, Issuer: identity.Issuer, Subject: identity.Subject}
	require.Error(t, db.Create(&duplicate).Error)
	otherSubject := UserOIDCIdentity{UID: user.UID, Issuer: identity.Issuer, Subject: "another-subject"}
	require.NoError(t, db.Create(&otherSubject).Error)
	otherIssuer := UserOIDCIdentity{UID: user.UID, Issuer: "https://other-idp.example.invalid", Subject: identity.Subject}
	require.NoError(t, db.Create(&otherIssuer).Error)
}
