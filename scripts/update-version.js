#!/usr/bin/env node
/**
 * 用法（在项目根目录）：
 *  node scripts/update-version.js 0.8.11      # 将 version 设置为 0.8.11
 *  node scripts/update-version.js patch (或 c) # 将 patch 自增（如 0.8.10 -> 0.8.11）
 *  node scripts/update-version.js minor (或 b) # 将 minor 自增（如 0.8.10 -> 0.9.0）
 *  node scripts/update-version.js major (或 a) # 将 major 自增（如 0.8.10 -> 1.0.0）
 *  或者使用环境变量： NEW_VERSION=0.8.11 node scripts/update-version.js
 *
 * 优先级（目标版本来源）：
 * 1. 命令行参数（node scripts/update-version.js <version|major|minor|patch>）
 * 2. 环境变量 NEW_VERSION
 */

const fs = require('fs');
const path = require('path');

/**
 * 从 Go 版本文件中读取当前版本号
 * @param {string} filePath - version.go 文件路径
 * @returns {string} 当前版本号
 */
function readGoVersion(filePath) {
    const content = fs.readFileSync(filePath, 'utf8');
    // 匹配 var Version string = "x.y.z" 格式
    const match = content.match(/var\s+Version\s+string\s*=\s*"([^"]+)"/);
    if (!match) {
        throw new Error('无法从文件中解析版本号: ' + filePath);
    }
    return match[1];
}

/**
 * 更新 Go 版本文件中的版本号
 * @param {string} filePath - version.go 文件路径
 * @param {string} newVersion - 新版本号
 */
function writeGoVersion(filePath, newVersion) {
    let content = fs.readFileSync(filePath, 'utf8');
    // 替换 Version 变量的值
    content = content.replace(
        /var\s+Version\s+string\s*=\s*"[^"]+"/,
        `var Version string = "${newVersion}"`
    );
    fs.writeFileSync(filePath, content, 'utf8');
}

/**
 * 验证版本号格式是否为 x.y.z
 * @param {string} v - 版本号
 * @returns {boolean}
 */
function isValidSemver(v) {
    return /^\d+\.\d+\.\d+$/.test(v);
}

/**
 * 版本号递增
 * @param {string} current - 当前版本号
 * @param {string} part - 递增部分：major/minor/patch
 * @returns {string} 新版本号
 */
function bumpVersion(current, part) {
    if (!isValidSemver(current)) {
        throw new Error('当前版本不是 x.y.z 格式: ' + current);
    }
    const [maj, min, pat] = current.split('.').map(n => parseInt(n, 10));
    if (part === 'major') return `${maj + 1}.0.0`;
    if (part === 'minor') return `${maj}.${min + 1}.0`;
    if (part === 'patch') return `${maj}.${min}.${pat + 1}`;
    throw new Error('未知的增量类型: ' + part);
}

/**
 * 更新版本文件
 * @param {string} filePath - 文件路径
 * @param {string|null} targetVersion - 目标版本号
 * @param {string|null} bumpOption - 递增选项
 * @returns {object|null} 更新结果
 */
function updateFileVersion(filePath, targetVersion, bumpOption) {
    if (!fs.existsSync(filePath)) {
        console.warn('文件不存在，跳过:', filePath);
        return null;
    }

    const from = readGoVersion(filePath);
    let to = targetVersion;

    if (!to && bumpOption) {
        to = bumpVersion(from, bumpOption);
    }

    if (!to) {
        throw new Error('没有提供目标版本或增量选项');
    }

    if (!isValidSemver(to)) {
        throw new Error('目标版本格式不合法，应为 x.y.z: ' + to);
    }

    writeGoVersion(filePath, to);
    return { filePath, from, to };
}

// 主逻辑
(function main() {
    const rawArgs = process.argv.slice(2);

    const aliasMap = { 'a': 'major', 'b': 'minor', 'c': 'patch' };
    const resolve = (v) => aliasMap[v] || v;

    const arg = resolve(rawArgs[0]);
    const envVersion = resolve(process.env.NEW_VERSION || null);
    const bumpOptions = new Set(['major', 'minor', 'patch']);

    let newVersion = null;
    let bumpOption = null;

    if (arg) {
        if (bumpOptions.has(arg)) {
            bumpOption = arg;
        } else if (isValidSemver(arg)) {
            newVersion = arg;
        } else {
            console.error('参数无效，应为 x.y.z 或 major/minor/patch');
            process.exit(1);
        }
    } else if (envVersion) {
        if (bumpOptions.has(envVersion)) {
            bumpOption = envVersion;
        } else if (isValidSemver(envVersion)) {
            newVersion = envVersion;
        } else {
            console.error('环境变量 NEW_VERSION 格式无效，应为 x.y.z 或 major/minor/patch');
            process.exit(1);
        }
    } else {
        console.error('未提供版本参数：使用 node scripts/update-version.js <version|major|minor|patch> 或 NEW_VERSION 环境变量');
        process.exit(1);
    }

    const cwd = process.cwd();
    const versionFile = path.join(cwd, 'global', 'version.go');

    try {
        const result = updateFileVersion(versionFile, newVersion, bumpOption);

        if (!result) {
            console.warn('没有更新任何文件。');
            process.exit(0);
        }

        console.log(`✓ ${path.relative(cwd, result.filePath)}: ${result.from} -> ${result.to}`);
    } catch (err) {
        console.error('错误：', err.message);
        process.exit(1);
    }
})();
