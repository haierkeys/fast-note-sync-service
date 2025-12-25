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
    "path" text DEFAULT '',
    "path_hash" text DEFAULT '',
    "content" text DEFAULT '',
    "content_hash" text DEFAULT '',
    "content_last_snapshot" text DEFAULT '',
    "version" integer DEFAULT 1,
    "size" integer DEFAULT 0,
    "ctime" integer DEFAULT 0,
    "mtime" integer DEFAULT 0,
    "updated_timestamp" integer DEFAULT 0,
    "created_at" datetime DEFAULT NULL,
    "updated_at" datetime DEFAULT NULL
);

CREATE INDEX "idx_vault_id_path_hash" ON "note" ("vault_id", "path_hash" DESC);

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
    "client" text DEFAULT '',
    "version" integer DEFAULT 1,
    "created_at" datetime DEFAULT NULL,
    "updated_at" datetime DEFAULT NULL
);

CREATE INDEX "idx_note_history_note_id" ON "note_history" ("note_id");

CREATE INDEX "idx_note_history_version" ON "note_history" ("note_id", "version");

DROP TABLE IF EXISTS "vault";

CREATE TABLE "vault" (
    "id" integer PRIMARY KEY AUTOINCREMENT,
    "vault" text DEFAULT '',
    "note_count" integer DEFAULT 0,
    "note_size" integer DEFAULT 0,
    "file_count" integer DEFAULT 0,
    "file_size" integer DEFAULT 0,
    "created_at" datetime DEFAULT NULL,
    "updated_at" datetime DEFAULT NULL
);

CREATE INDEX "idx_vault_uid" ON "vault" ("vault" ASC);

-- 笔记库索引
PRAGMA foreign_keys = true;

## mysql
DROP TABLE IF EXISTS `pre_user`;

CREATE TABLE `pre_user` (
    `uid` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户id',
    `email` char(255) NOT NULL DEFAULT '' COMMENT '邮箱地址',
    `username` char(255) NOT NULL DEFAULT '' COMMENT '用户名',
    `password` char(32) NOT NULL DEFAULT '' COMMENT '密码',
    `salt` char(24) NOT NULL DEFAULT '' COMMENT '密码混淆码',
    `token` char(255) NOT NULL DEFAULT '' COMMENT '用户授权令牌',
    `avatar` char(255) NOT NULL DEFAULT '' COMMENT '用户头像路径',
    `is_deleted` tinyint(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否删除',
    `updated_at` datetime(0) NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '更新时间',
    `created_at` datetime(0) NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '创建时间',
    `deleted_at` datetime(0) NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '标记删除时间',
    PRIMARY KEY (`uid`),
    UNIQUE INDEX `email`(`email`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COMMENT = '用户表';

DROP TABLE IF EXISTS "pre_cloud_config";

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