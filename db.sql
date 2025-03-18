
# sqlite3

PRAGMA foreign_keys = false;

-- ----------------------------
-- Table structure for pre_user
-- ----------------------------
DROP TABLE IF EXISTS "pre_user";
CREATE TABLE "pre_user" (
    "uid" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, -- 用户id
    "email" TEXT DEFAULT '', -- 邮箱地址
    "username" TEXT DEFAULT '', -- 用户名
    "password" TEXT DEFAULT '', -- 密码
    "salt" TEXT DEFAULT '', -- 密码混淆码
    "token" TEXT NOT NULL DEFAULT '', -- 用户授权令牌
    "avatar" TEXT NOT NULL DEFAULT '', -- 用户头像路径
    "is_deleted" INTEGER NOT NULL DEFAULT 0, -- 是否删除
    "updated_at" TIMESTAMP DEFAULT NULL, -- 更新时间
    "created_at" TIMESTAMP DEFAULT NULL, -- 创建时间
    "deleted_at" TIMESTAMP DEFAULT NULL, -- 标记删除时间
    UNIQUE ("email" ASC) -- 唯一索引
);
CREATE UNIQUE INDEX "email" ON "pre_user" ("email" ASC);

DROP TABLE IF EXISTS "pre_cloud_config";
CREATE TABLE "pre_cloud_config" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, -- 配置id
    "uid" INTEGER NOT NULL DEFAULT 0, -- 用户id
    "type" TEXT DEFAULT '', -- 类型
    "endpoint" TEXT DEFAULT '',-- 账户id
    "region" TEXT DEFAULT '',-- 账户id
    "account_id" TEXT DEFAULT '', -- 账户id
    "bucket_name" TEXT DEFAULT '',-- 存储桶名称
    "access_key_id" TEXT DEFAULT '', -- 访问密钥id
    "access_key_secret" TEXT DEFAULT '', -- 访问密钥
    "custom_path" TEXT DEFAULT '', -- 自定义路径
    "access_url_prefix" TEXT DEFAULT '', -- 访问地址前缀
    "is_enabled" INTEGER NOT NULL DEFAULT 1, -- 是否启用
    "is_deleted" INTEGER NOT NULL DEFAULT 0, -- 是否删除
    "updated_at" TIMESTAMP DEFAULT NULL, -- 更新时间
    "created_at" TIMESTAMP DEFAULT NULL, -- 创建时间
    "deleted_at" TIMESTAMP DEFAULT NULL -- 标记删除时间
);
CREATE INDEX "uid" ON "pre_cloud_config" ("uid" ASC); -- 用户id索引

PRAGMA foreign_keys = true;


## mysql
DROP TABLE IF EXISTS `pre_user`;
CREATE TABLE `pre_user`  (
 `uid` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户id',
 `email` char(255)  NOT NULL DEFAULT '' COMMENT '邮箱地址',
 `username` char(255)  NOT NULL DEFAULT '' COMMENT '用户名',
 `password` char(32)  NOT NULL DEFAULT '' COMMENT '密码',
 `salt` char(24)  NOT NULL DEFAULT '' COMMENT '密码混淆码',
 `token` char(255)  NOT NULL DEFAULT '' COMMENT '用户授权令牌',
 `avatar` char(255)  NOT NULL DEFAULT '' COMMENT '用户头像路径',
 `is_deleted` tinyint(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否删除',
 `updated_at` datetime(0) NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '更新时间',
 `created_at` datetime(0) NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '创建时间',
 `deleted_at` datetime(0) NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '标记删除时间',
 PRIMARY KEY (`uid`) ,
 UNIQUE INDEX `email`(`email`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4  COMMENT = '用户表';

DROP TABLE IF EXISTS "pre_cloud_config";

CREATE TABLE `pre_cloud_config` (
    `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '配置id',
    `uid` bigint(20) UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户id',
    `type` char(255) NOT NULL DEFAULT '' COMMENT '类型',
    `bucket_name` char(255) NOT NULL DEFAULT '' COMMENT '存储桶名称',
    `account_id` char(255) NOT NULL DEFAULT '' COMMENT '账户id',
    `access_key_id` char(255) NOT NULL DEFAULT '' COMMENT '访问密钥id',
    `access_key_secret` char(255) NOT NULL DEFAULT '' COMMENT '访问密钥',
    `custom_path` char(255) NOT NULL DEFAULT '' COMMENT '自定义路径',
    `access_url_prefix` char(255) NOT NULL DEFAULT '' COMMENT '访问地址前缀',
    `is_enabled` tinyint(1) UNSIGNED NOT NULL DEFAULT 1 COMMENT '是否启用',
    `is_deleted` tinyint(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否删除',
    `updated_at` datetime(0) NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '更新时间',
    `created_at` datetime(0) NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '创建时间',
    `deleted_at` datetime(0) NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '标记删除时间',
    PRIMARY KEY (`id`),
    INDEX `uid`(`uid`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COMMENT = '云配置表';
