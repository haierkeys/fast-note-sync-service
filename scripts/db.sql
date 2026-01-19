# sqlite3
PRAGMA foreign_keys = false;

-- ----------------------------
-- Table structure for pre_user
-- ----------------------------
DROP TABLE IF EXISTS "user";

CREATE TABLE "user" (
    `uid` integer PRIMARY KEY AUTOINCREMENT,
    `email` text DEFAULT "",
    `username` text DEFAULT "",
    `password` text DEFAULT "",
    `salt` text DEFAULT "",
    `token` text DEFAULT "",
    `avatar` text DEFAULT "",
    `is_deleted` integer DEFAULT 0,
    `updated_at` datetime DEFAULT NULL,
    `created_at` datetime DEFAULT NULL,
    `deleted_at` datetime DEFAULT NULL
);

CREATE INDEX `idx_pre_user_email` ON "user"(`email`);

/*
 动作 create modify delete
 */
DROP TABLE IF EXISTS "note";

CREATE TABLE "note" (
    "id" integer PRIMARY KEY AUTOINCREMENT,
    "vault_id" integer NOT NULL DEFAULT 0,
    "action" text DEFAULT '',
    "rename" integer DEFAULT 0,
    "path" text DEFAULT '',
    "path_hash" text DEFAULT '',
    "content" text DEFAULT '',
    "content_hash" text DEFAULT '',
    "content_last_snapshot" text NOT NULL DEFAULT '',
    "content_last_snapshot_hash" text NOT NULL DEFAULT '',
    "version" integer DEFAULT 0,
    "client_name" text NOT NULL DEFAULT '',
    "size" integer DEFAULT 0,
    "ctime" integer DEFAULT 0,
    "mtime" integer DEFAULT 0,
    "updated_timestamp" integer DEFAULT 0,
    "created_at" datetime DEFAULT NULL,
    "updated_at" datetime DEFAULT NULL
);

CREATE INDEX "idx_vault_id_action_rename" ON "note" ("vault_id", "action", "rename" DESC);

CREATE INDEX "idx_vault_id_rename" ON "note" ("vault_id", "rename" DESC);

CREATE INDEX "idx_vault_id_updated_at" ON "note" ("vault_id", "updated_at" DESC);

CREATE INDEX "idx_vault_id_updated_timestamp" ON "note" ("vault_id", "updated_timestamp" DESC);

CREATE INDEX `idx_vault_id_path` ON `note`(`vault_id`, `path`);

DROP TABLE IF EXISTS "note_history";

CREATE TABLE "note_history" (
    "id" integer PRIMARY KEY AUTOINCREMENT,
    "note_id" integer NOT NULL DEFAULT 0,
    "vault_id" integer NOT NULL DEFAULT 0,
    "path" text DEFAULT '',
    "content" text DEFAULT '',
    "content_hash" text NOT NULL DEFAULT '',
    "diff_patch" text DEFAULT '',
    "client_name" text DEFAULT '',
    "version" integer DEFAULT 0,
    "created_at" datetime DEFAULT NULL,
    "updated_at" datetime DEFAULT NULL
);

CREATE INDEX "idx_note_history_note_id" ON "note_history" ("note_id");

CREATE INDEX "idx_note_history_version" ON "note_history" ("note_id", "version");

CREATE INDEX "idx_note_history_content_hash" ON "note_history" ("note_id", "content_hash");

DROP TABLE IF EXISTS "vault";

CREATE TABLE "vault" (
    "id" integer PRIMARY KEY AUTOINCREMENT,
    "vault" text DEFAULT '',
    "note_count" integer DEFAULT 0,
    "note_size" integer DEFAULT 0,
    "file_count" integer DEFAULT 0,
    "file_size" integer DEFAULT 0,
    "is_deleted" integer DEFAULT 0,
    "created_at" datetime DEFAULT NULL,
    "updated_at" datetime DEFAULT NULL
);

CREATE INDEX "idx_vault_uid" ON "vault" ("vault" ASC);

DROP TABLE IF EXISTS "file";

CREATE TABLE "file" (
    "id" integer PRIMARY KEY AUTOINCREMENT,
    "vault_id" integer NOT NULL DEFAULT 0,
    "action" text DEFAULT '',
    "path" text DEFAULT '',
    "path_hash" text DEFAULT '',
    "content_hash" text DEFAULT '',
    "save_path" text DEFAULT '',
    "size" integer NOT NULL DEFAULT 0,
    "ctime" integer NOT NULL DEFAULT 0,
    "mtime" integer NOT NULL DEFAULT 0,
    "updated_timestamp" integer NOT NULL DEFAULT 0,
    "created_at" datetime DEFAULT NULL,
    "updated_at" datetime DEFAULT NULL
);

CREATE INDEX "idx_file_vault_id_path_hash" ON "file" ("vault_id", "path_hash" DESC);

CREATE INDEX "idx_file_vault_id_updated_at" ON "file" ("vault_id", "updated_at" DESC);

CREATE INDEX "idx_file_vault_id_updated_timestamp" ON "file" ("vault_id", "updated_timestamp" DESC);

CREATE INDEX `idx_file_vault_id_path` ON `file`(`vault_id`, `path`);

DROP TABLE IF EXISTS "setting";

CREATE TABLE "setting" (
    "id" integer PRIMARY KEY AUTOINCREMENT,
    "vault_id" integer NOT NULL DEFAULT 0,
    "action" text DEFAULT '',
    "path" text DEFAULT '',
    "path_hash" text DEFAULT '',
    "content" text DEFAULT '',
    "content_hash" text DEFAULT '',
    "size" integer DEFAULT 0,
    "ctime" integer DEFAULT 0,
    "mtime" integer DEFAULT 0,
    "updated_timestamp" integer DEFAULT 0,
    "created_at" datetime DEFAULT NULL,
    "updated_at" datetime DEFAULT NULL
);

CREATE INDEX "idx_setting_id_path_hash" ON "setting" ("id", "path_hash" DESC);

CREATE INDEX "idx_setting_id_updated_at" ON "setting" ("id", "updated_at" DESC);

CREATE INDEX "idx_setting_id_updated_timestamp" ON "setting" ("id", "updated_timestamp" DESC);

CREATE INDEX `idx_setting_id_path` ON `setting`(`id`, `path`);

DROP TABLE IF EXISTS "user_share";

CREATE TABLE "user_share" (
    "id" integer PRIMARY KEY AUTOINCREMENT,
    "uid" integer NOT NULL DEFAULT 0,
    "res" text NOT NULL DEFAULT '',
    -- 资源列表 (JSON: {"note":["id1"],"file":["id2"]})
    "status" integer DEFAULT 1,
    -- 1-有效, 2-已撤销
    "view_count" integer DEFAULT 0,
    -- 访问次数
    "last_viewed_at" datetime DEFAULT NULL,
    "expires_at" datetime DEFAULT NULL,
    "created_at" datetime DEFAULT NULL,
    "updated_at" datetime DEFAULT NULL
);

CREATE INDEX "idx_user_share_uid" ON "user_share" ("uid");

CREATE INDEX "idx_user_share_rid" ON "user_share" ("rid");