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

DROP TABLE IF EXISTS "note";

CREATE TABLE "note" (
    "id" integer PRIMARY KEY AUTOINCREMENT,
    "vault_id" integer NOT NULL DEFAULT 0,
    "action" text DEFAULT '',
    "rename" integer DEFAULT 0,
    "fid" integer DEFAULT 0,
    -- note table : parent folder id, 0 : root
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

CREATE INDEX "idx_vault_id_action_fid" ON "note" ("vault_id", "action", "fid" DESC);

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

DROP TABLE IF EXISTS "file";

CREATE TABLE "file" (
    "id" integer PRIMARY KEY AUTOINCREMENT,
    "vault_id" integer NOT NULL DEFAULT 0,
    "action" text DEFAULT '',
    "fid" integer DEFAULT 0,
    -- folder table : parent folder id, 0 : root
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

CREATE INDEX "idx_file_vault_id_action_fid" ON "file" ("vault_id", "action", "fid" DESC);

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

-- ----------------------------
-- Table structure for note_link
-- ----------------------------
DROP TABLE IF EXISTS "note_link";

CREATE TABLE "note_link" (
    "id" integer PRIMARY KEY AUTOINCREMENT,
    "source_note_id" integer NOT NULL,
    "target_path" text NOT NULL,
    "target_path_hash" text NOT NULL,
    "link_text" text,
    "is_embed" integer DEFAULT 0,
    "vault_id" integer NOT NULL,
    "uid" integer NOT NULL,
    "created_at" datetime DEFAULT NULL
);

CREATE INDEX "idx_source_note" ON "note_link" ("source_note_id");

CREATE INDEX "idx_target_path_hash" ON "note_link" ("target_path_hash", "vault_id", "uid");

DROP TABLE IF EXISTS "folder";

CREATE TABLE "folder" (
    "id" integer PRIMARY KEY AUTOINCREMENT,
    "vault_id" integer NOT NULL DEFAULT 0,
    "action" text DEFAULT '',
    "path" text DEFAULT '',
    "path_hash" text DEFAULT '',
    "level" integer DEFAULT 0,
    -- 文件夹层级
    "fid" integer DEFAULT 0,
    -- 父级文件夹ID,0 为根目录
    "ctime" integer NOT NULL DEFAULT 0,
    "mtime" integer NOT NULL DEFAULT 0,
    "updated_timestamp" integer NOT NULL DEFAULT 0,
    "created_at" datetime DEFAULT NULL,
    "updated_at" datetime DEFAULT NULL
);

CREATE INDEX "idx_folder_vault_id_path_hash" ON "folder" ("vault_id", "path_hash");

CREATE INDEX `idx_folder_vault_id_path` ON `folder`(`vault_id`, `path`);

CREATE INDEX "idx_folder_vault_id_fid_path" ON "folder" ("vault_id", "fid", "path");

CREATE INDEX "idx_folder_vault_id_level_path" ON "folder" ("vault_id", "level", "path");

CREATE INDEX "idx_folder_vault_id_updated_timestamp" ON "folder" ("vault_id", "updated_timestamp" DESC);

DROP TABLE IF EXISTS "storage";

CREATE TABLE "storage" (
    "id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "uid" integer NOT NULL DEFAULT 0,
    "type" text DEFAULT '',
    "endpoint" text DEFAULT '',
    "region" text DEFAULT '',
    "account_id" text DEFAULT '',
    "bucket_name" text DEFAULT '',
    "access_key_id" text DEFAULT '',
    "access_key_secret" text DEFAULT '',
    "custom_path" text DEFAULT '',
    "access_url_prefix" text DEFAULT '',
    "user" text DEFAULT '',
    "password" text DEFAULT '',
    "is_enabled" integer NOT NULL DEFAULT 0,
    "is_deleted" integer NOT NULL DEFAULT 0,
    "created_at" datetime DEFAULT NULL,
    "updated_at" datetime DEFAULT NULL,
    "deleted_at" datetime DEFAULT NULL
);

CREATE INDEX "idx_storage_uid" ON "storage" ("uid" DESC);

-- ----------------------------
-- Table structure for backup_config
-- ----------------------------
DROP TABLE IF EXISTS "backup_config";

CREATE TABLE "backup_config" (
    "id" integer PRIMARY KEY AUTOINCREMENT,
    "uid" integer NOT NULL DEFAULT 0,
    "vault_id" integer NOT NULL DEFAULT 0,
    "type" text DEFAULT '',
    -- full, incremental, sync
    "storage_ids" text DEFAULT '',
    -- JSON array of storage ids: [1, 2]
    "is_enabled" integer DEFAULT 0,
    "cron_strategy" text DEFAULT '',
    -- daily, weekly, monthly, custom
    "cron_expression" text DEFAULT '',
    "retention_days" integer DEFAULT 10,
    -- Retention policy (days)
    "last_run_time" datetime DEFAULT NULL,
    "next_run_time" datetime DEFAULT NULL,
    "last_status" integer DEFAULT 0,
    -- 0: Idle, 1: Running, 2: Success, 3: Failed, 4: Stopped, 5: Success but no update
    "last_message" text DEFAULT '',
    "created_at" datetime DEFAULT NULL,
    "updated_at" datetime DEFAULT NULL
);

CREATE INDEX "idx_backup_config_uid" ON "backup_config" ("uid");

CREATE INDEX "idx_backup_config_next_run_time" ON "backup_config" ("next_run_time");

-- ----------------------------
-- Table structure for backup_history
-- ----------------------------
DROP TABLE IF EXISTS "backup_history";

CREATE TABLE "backup_history" (
    "id" integer PRIMARY KEY AUTOINCREMENT,
    "uid" integer NOT NULL DEFAULT 0,
    "config_id" integer NOT NULL DEFAULT 0,
    "storage_id" integer NOT NULL DEFAULT 0,
    "type" text DEFAULT '',
    -- full, incremental, sync
    "start_time" datetime DEFAULT NULL,
    "end_time" datetime DEFAULT NULL,
    "status" integer DEFAULT 0,
    -- 0: Idle, 1: Running, 2: Success, 3: Failed, 4: Stopped, 5: Success but no update
    "file_size" integer DEFAULT 0,
    -- bytes
    "file_count" integer DEFAULT 0,
    "message" text DEFAULT '',
    "file_path" text DEFAULT '',
    -- remote path/key
    "created_at" datetime DEFAULT NULL,
    "updated_at" datetime DEFAULT NULL
);

CREATE INDEX "idx_backup_history_uid" ON "backup_history" ("uid", "created_at" DESC);

CREATE INDEX "idx_backup_history_config_id" ON "backup_history" ("config_id");

-- ----------------------------
-- Table structure for git_sync_config
-- ----------------------------
DROP TABLE IF EXISTS "git_sync_config";

CREATE TABLE "git_sync_config" (
    "id" integer PRIMARY KEY AUTOINCREMENT,
    "uid" integer NOT NULL DEFAULT 0,
    "vault_id" integer NOT NULL DEFAULT 0,
    "repo_url" text DEFAULT '',
    -- Git 仓库地址, 例如 https://github.com/user/repo.git
    "username" text DEFAULT '',
    -- 认证用户名
    "password" text DEFAULT '',
    -- 认证密码或 Personal Access Token
    "branch" text DEFAULT 'main',
    -- 分支名
    "is_enabled" integer DEFAULT 0,
    -- 是否启用自动同步
    "delay" integer DEFAULT 0,
    -- 延迟时间（例如同步前等待的时间，单位可以是秒或分钟）
    "retention_days" integer DEFAULT 0,
    -- 历史记录保留天数, 0: 不清理, -1: 仅保留最新, >0: 保留天数
    "last_sync_time" datetime DEFAULT NULL,
    -- 上次同步时间
    "last_status" integer DEFAULT 0,
    -- 0: 闲置, 1: 运行中, 2: 成功, 3: 失败
    "last_message" text DEFAULT '',
    -- 同步结果或错误信息
    "created_at" datetime DEFAULT NULL,
    "updated_at" datetime DEFAULT NULL
);

CREATE INDEX "idx_git_sync_config_uid" ON "git_sync_config" ("uid");

-- ----------------------------
-- Table structure for git_sync_history
-- ----------------------------
DROP TABLE IF EXISTS "git_sync_history";

CREATE TABLE "git_sync_history" (
    "id" integer PRIMARY KEY AUTOINCREMENT,
    "uid" integer NOT NULL DEFAULT 0,
    "config_id" integer NOT NULL DEFAULT 0,
    "start_time" datetime DEFAULT NULL,
    "end_time" datetime DEFAULT NULL,
    "status" integer DEFAULT 0,
    -- 0: Idle, 1: Running, 2: Success, 3: Failed, 4: Shutdown
    "message" text DEFAULT '',
    "created_at" datetime DEFAULT NULL,
    "updated_at" datetime DEFAULT NULL
);

CREATE INDEX "idx_git_sync_history_uid" ON "git_sync_history" ("uid", "created_at" DESC);

CREATE INDEX "idx_git_sync_history_config_id" ON "git_sync_history" ("config_id");