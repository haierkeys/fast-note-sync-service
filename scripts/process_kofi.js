const fs = require('fs');
const path = require('path');
const https = require('https');

const ENV_FILE = path.join(__dirname, '..', '.env');
const OUTPUT_FILE = path.join(__dirname, '..', 'docs', 'Kofi_Support.csv');

const KOFI_URL = "https://ko-fi.com/api/transactions/download-csv?selectedMonth=All&transactionType=all&purchaseSource=undefined&searchKey=";

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

function formatCsvField(field) {
    if (field === null || field === undefined) return '';
    if (typeof field !== 'string') field = String(field);
    field = field.trim();
    if (field.includes(',') || field.includes('"') || field.includes('\n')) {
        return `"${field.replace(/"/g, '""')}"`;
    }
    return field;
}

async function fetchKofiCsv(cookie) {
    console.log("Fetching Ko-fi CSV data...");
    
    return new Promise((resolve, reject) => {
        const options = {
            headers: {
                'Cookie': cookie,
                'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
            }
        };

        https.get(KOFI_URL, options, (res) => {
            if (res.statusCode !== 200) {
                return reject(new Error(`Failed to fetch, status code: ${res.statusCode}`));
            }

            let data = '';
            res.on('data', chunk => data += chunk);
            res.on('end', () => resolve(data));
        }).on('error', err => reject(err));
    });
}

async function main() {
    const cookie = await getKofiCookie();
    if (!cookie) {
        console.error("Error: Ko-fi cookie not found.");
        console.error(`Please provide the cookie by setting the KOFI_COOKIE environment variable or adding it to: ${ENV_FILE}`);
        process.exit(1);
    }

    try {
        const csvData = await fetchKofiCsv(cookie);
        const lines = csvData.split(/\r?\n/).filter(line => line.trim() !== '');
        
        if (lines.length < 2) {
            console.warn("Downloaded CSV is empty or only contains headers.");
            return;
        }

        // The first line should be the header
        const headers = parseCsvLine(lines[0]).map(h => h.trim().toLowerCase());
        
        // Find column indexes
        const idxDate = headers.findIndex(h => h.includes('date'));
        const idxName = headers.findIndex(h => h === 'name' || h.includes('supporter'));
        const idxMessage = headers.findIndex(h => h === 'message');
        const idxAmount = headers.findIndex(h => h === 'amount');
        const idxCurrency = headers.findIndex(h => h === 'currency');
        const idxType = headers.findIndex(h => h === 'type' || h.includes('transaction type'));
        
        // Process data
        const dataRows = [];
        
        for (let i = 1; i < lines.length; i++) {
            const fields = parseCsvLine(lines[i]);
            if (fields.length < headers.length) continue;
            
            const time = idxDate >= 0 ? fields[idxDate].trim() : '';
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
