/**
 * Ko-fi 订单数据提取与处理脚本
 *
 * 脚本说明：
 * 该脚本通过 Ko-fi 的 API 自动下载所有交易历史记录的 CSV 数据，
 * 提取、清洗和格式化所需的核心字段（收款时间、收款项、金额、单位、留言、昵称），
 * 并最终按金额降序、时间降序的规则生成一个新的 CSV 文件 (docs/Support_kofi.csv)。
 *
 * 运行前提：
 * 1. 登录 Ko-fi 网页端 (https://ko-fi.com/)
 * 2. 按 F12 打开浏览器开发者工具，切换到 "网络" (Network) 面板
 * 3. 刷新页面，点击任意一个请求（如对 ko-fi.com 的文档或接口请求）
 * 4. 在 "请求标头" (Request Headers) 中找到 `cookie` 字段并复制其完整的字符串值
 *    完整的 Cookie 必须包含以下关键鉴权字段：
 *    - `.AspNet.ApplicationCookie`: 核心登录凭证（最重要）
 *    - `ASP.NET_SessionId`: 会话标识
 *    - `ARRAffinity`: 服务器节点关联
 * 5. 在项目根目录的 `.env` 文件中配置 `KOFI_COOKIE` 变量，或直接通过系统环境变量传入。
 *    示例：
 *    KOFI_COOKIE="ARRAffinity=xxxxxxxx; ASP.NET_SessionId=yyyyyyyy; .AspNet.ApplicationCookie=zzzzzzzz;"
 *
 * 使用方式：
 * node scripts/process_kofi.js
 *
 * 本地备用方案：
 * 如果由于 Cloudflare 403 限制导致自动下载失败，您可以手动从 Ko-fi 页面下载交易记录 CSV 命名为 `Support_kofi_Raw.csv` 并放置到 `docs/` 目录下，脚本会自动检测并进行本地处理。
 */
const fs = require('fs');
const path = require('path');
const https = require('https');

const ENV_FILE = path.join(__dirname, '..', '.env');
const OUTPUT_FILE = path.join(__dirname, '..', 'docs', 'Support_kofi.csv');
const RAW_FILE = path.join(__dirname, '..', 'docs', 'Support_kofi_Raw.csv');

const KOFI_URL = "https://ko-fi.com/api/transactions/download-csv?selectedMonth=All&transactionType=all&purchaseSource=undefined&searchKey=";

/**
 * 获取 Ko-fi 的身份验证 Cookie
 * 优先从系统环境变量获取，如果不存在，则尝试读取项目根目录的 .env 文件
 */
async function getKofiCookie() {
    if (process.env.KOFI_COOKIE) {
        return process.env.KOFI_COOKIE;
    }
    if (fs.existsSync(ENV_FILE)) {
        const content = fs.readFileSync(ENV_FILE, 'utf8');
        const match = content.match(/^KOFI_COOKIE=(.*)$/m);
        if (match) {
            let cookie = match[1].trim();
            // 去除可能存在的引号
            if ((cookie.startsWith('"') && cookie.endsWith('"')) || (cookie.startsWith("'") && cookie.endsWith("'"))) {
                cookie = cookie.slice(1, -1);
            }
            return cookie;
        }
    }
    return null;
}

/**
 * 解析单行 CSV 文本，能够正确处理包含在引号内部的逗号等特殊情况
 * @param {string} line - CSV 单行字符串
 * @returns {Array<string>} - 解析后的字段数组
 */
function parseCsvLine(line) {
    const fields = [];
    let currentField = '';
    let inQuotes = false;
    for (let i = 0; i < line.length; i++) {
        const char = line[i];
        if (char === '"') {
            inQuotes = !inQuotes;
        } else if (char === ',' && !inQuotes) {
            fields.push(currentField);
            currentField = '';
        } else {
            currentField += char;
        }
    }
    fields.push(currentField);
    return fields;
}

/**
 * 格式化 CSV 字段输出，处理字段中可能包含的逗号、双引号或换行符，
 * 使用双引号包裹以防破坏 CSV 结构（符合 RFC 4180 规范）
 * @param {any} field - 待格式化的字段值
 * @returns {string} - 格式化且安全转义的字符串
 */
function formatCsvField(field) {
    if (field === null || field === undefined) return '';
    if (typeof field !== 'string') field = String(field);
    field = field.trim();
    if (field.includes(',') || field.includes('"') || field.includes('\n')) {
        return `"${field.replace(/"/g, '""')}"`;
    }
    return field;
}

/**
 * 将 Ko-fi 原始时间格式 (如 MM/DD/YYYY HH:mm) 标准化为 YYYY/MM/DD HH:mm:ss
 * @param {string} dateStr - 原始时间字符串
 * @returns {string} - 标准化后的时间字符串
 */
function formatKofiDate(dateStr) {
    if (!dateStr) return '';
    const parts = dateStr.split(/[\/\s-:]/);
    if (parts.length >= 5) {
        let year, month, day, hour, minute, second = '00';
        if (parts[2] && parts[2].length === 4) {
            // MM/DD/YYYY
            year = parts[2];
            month = parts[0].padStart(2, '0');
            day = parts[1].padStart(2, '0');
        } else if (parts[0] && parts[0].length === 4) {
            // YYYY-MM-DD or YYYY/MM/DD
            year = parts[0];
            month = parts[1].padStart(2, '0');
            day = parts[2].padStart(2, '0');
        } else {
            const d = new Date(dateStr);
            if (!isNaN(d.getTime())) {
                const y = d.getFullYear();
                const m = String(d.getMonth() + 1).padStart(2, '0');
                const dayStr = String(d.getDate()).padStart(2, '0');
                const h = String(d.getHours()).padStart(2, '0');
                const min = String(d.getMinutes()).padStart(2, '0');
                const s = String(d.getSeconds()).padStart(2, '0');
                return `${y}/${m}/${dayStr} ${h}:${min}:${s}`;
            }
            return dateStr;
        }

        hour = parts[3].padStart(2, '0');
        minute = parts[4].padStart(2, '0');
        if (parts[5]) {
            second = parts[5].padStart(2, '0');
        }
        return `${year}/${month}/${day} ${hour}:${minute}:${second}`;
    }
    
    const d = new Date(dateStr);
    if (!isNaN(d.getTime())) {
        const y = d.getFullYear();
        const m = String(d.getMonth() + 1).padStart(2, '0');
        const dayStr = String(d.getDate()).padStart(2, '0');
        const h = String(d.getHours()).padStart(2, '0');
        const min = String(d.getMinutes()).padStart(2, '0');
        const s = String(d.getSeconds()).padStart(2, '0');
        return `${y}/${m}/${dayStr} ${h}:${min}:${s}`;
    }
    return dateStr;
}


/**
 * 发起 HTTPS 请求获取 Ko-fi 历史订单的原始 CSV 数据
 * 采用 got-scraping 库自动模拟真实浏览器 (Chrome/macOS) 的 TLS/HTTP2 指纹，绕过 Cloudflare WAF 拦截
 * @param {string} cookie - 身份验证 Cookie
 * @returns {Promise<string>} - 返回包含所有交易记录的 CSV 文本数据
 */
async function fetchKofiCsv(cookie) {
    console.log("Fetching Ko-fi CSV data via got-scraping...");

    const { gotScraping } = await import('got-scraping');

    const response = await gotScraping({
        url: KOFI_URL,
        method: 'GET',
        headers: {
            'cookie': cookie,
            'accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7',
            'accept-language': 'zh-CN,zh;q=0.9,en;q=0.8',
            'cache-control': 'no-cache',
            'pragma': 'no-cache',
            'upgrade-insecure-requests': '1'
        },
        headerGeneratorOptions: {
            browsers: [
                {
                    name: 'chrome',
                    minVersion: 120
                }
            ],
            devices: ['desktop'],
            operatingSystems: ['macos']
        }
    });

    if (response.statusCode !== 200) {
        throw new Error(`Failed to fetch, status code: ${response.statusCode}`);
    }

    return response.body;
}

/**
 * 主执行流程
 * 1. 检测是否存在本地原始 CSV 文件 docs/Support_kofi_Raw.csv
 * 2. 如果存在本地文件，直接进行解析处理；否则，尝试通过 API 自动下载
 * 3. 动态识别表头索引，提取目标字段并进行数据清洗与标准化处理
 * 4. 排序数据：优先按赞助金额降序，金额相同时按时间降序
 * 5. 将处理后的结果覆盖写入目标文件 docs/Support_kofi.csv
 */
async function main() {
    let csvData = '';

    // 优先检测本地是否存在手动下载的原始数据文件，用于绕过 Cloudflare WAF 限制
    if (fs.existsSync(RAW_FILE)) {
        console.log(`Found local raw CSV file: ${RAW_FILE}. Processing locally...`);
        try {
            csvData = fs.readFileSync(RAW_FILE, 'utf8');
        } catch (err) {
            console.error(`Error reading local file ${RAW_FILE}:`, err);
            process.exit(1);
        }
    } else {
        const cookie = await getKofiCookie();
        if (!cookie) {
            console.error("Error: Ko-fi cookie not found.");
            console.error(`Please provide the cookie by setting the KOFI_COOKIE environment variable or adding it to: ${ENV_FILE}`);
            console.error(`Alternatively, you can manually download the CSV and save it as: ${RAW_FILE}`);
            process.exit(1);
        }

        console.log("Loaded KOFI_COOKIE:", cookie);

        try {
            csvData = await fetchKofiCsv(cookie);
        } catch (err) {
            console.error("Error processing Ko-fi data:", err);
            console.error("\n==========================================================================");
            console.error("💡 提示 (Cloudflare WAF 拦截 403 避坑指南):");
            console.error("由于 Ko-fi 使用了 Cloudflare WAF 安全系统，Node.js 客户端的 TLS/HTTP2 特征");
            console.error("与真实浏览器不完全匹配，即使 Cookie 最新，也可能会被拦截返回 403 Challenge。");
            console.error("\n🛠️ 本地备用解决方案 (推荐):");
            console.error(`1. 登录 Ko-fi，在浏览器中点击下载交易 CSV 文件`);
            console.error(`2. 将下载好的 CSV 文件重命名为: Support_kofi_Raw.csv`);
            console.error(`3. 移动或保存至 docs 目录下，路径为: ${RAW_FILE}`);
            console.error(`4. 再次执行脚本，脚本将自动检测并读取该本地文件进行转换处理。`);
            console.error("==========================================================================\n");
            process.exit(1);
        }
    }

    try {
        const lines = csvData.split(/\r?\n/).filter(line => line.trim() !== '');

        if (lines.length < 2) {
            console.warn("Downloaded or local CSV is empty or only contains headers.");
            return;
        }

        // The first line should be the header
        const headers = parseCsvLine(lines[0]).map(h => h.trim().toLowerCase());

        // Find column indexes
        const idxDate = headers.findIndex(h => h.includes('date'));
        const idxName = headers.findIndex(h => h === 'name' || h.includes('supporter') || h === 'from');
        const idxMessage = headers.findIndex(h => h === 'message');
        const idxAmount = headers.findIndex(h => h === 'amount' || h === 'received' || h === 'given');
        const idxCurrency = headers.findIndex(h => h === 'currency');
        const idxType = headers.findIndex(h => h === 'type' || h.includes('transaction type') || h.includes('transactiontype') || h === 'item');

        // Process data
        const dataRows = [];

        for (let i = 1; i < lines.length; i++) {
            const fields = parseCsvLine(lines[i]);
            if (fields.length < headers.length) continue;

            const rawTime = idxDate >= 0 ? fields[idxDate].trim() : '';
            const time = formatKofiDate(rawTime);
            const name = idxName >= 0 ? fields[idxName].trim() : '';
            const message = idxMessage >= 0 ? fields[idxMessage].trim() : '';
            const amountStr = idxAmount >= 0 ? fields[idxAmount].trim() : '0';
            const unit = idxCurrency >= 0 ? fields[idxCurrency].trim() : '';
            let item = idxType >= 0 ? fields[idxType].trim() : 'Coffee';

            // Format amount
            const amountValue = parseFloat(amountStr.replace(/[^\d.-]/g, '')) || 0;
            if (amountValue <= 0) continue; // Skip 0 amounts if any

            // Standardize item types to some extent if needed, or leave as is
            if (!item) item = 'Coffee';

            dataRows.push({
                time,
                item,
                amountVal: amountValue,
                amountStr: amountValue.toFixed(2),
                unit,
                message,
                name
            });
        }

        // Sort: Amount descending, then Date descending
        dataRows.sort((a, b) => {
            if (b.amountVal !== a.amountVal) {
                return b.amountVal - a.amountVal;
            }
            return b.time.localeCompare(a.time);
        });

        // Output format: 收款时间,收款项,金额,单位,留言,昵称
        const result = [];
        result.push('收款时间,收款项,金额,单位,留言,昵称');

        dataRows.forEach(row => {
            const rowStr = [
                formatCsvField(row.time),
                formatCsvField(row.item),
                formatCsvField(row.amountStr),
                formatCsvField(row.unit),
                formatCsvField(row.message),
                formatCsvField(row.name)
            ].join(',');
            result.push(rowStr);
        });

        fs.writeFileSync(OUTPUT_FILE, result.join('\n') + '\n', 'utf8');
        console.log(`Successfully processed ${dataRows.length} transactions.`);
        console.log(`Saved to ${OUTPUT_FILE}`);

    } catch (err) {
        console.error("Error processing Ko-fi data:", err);
    }
}

main();
