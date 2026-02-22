const fs = require('fs');
const path = require('path');

const docsDir = path.join(__dirname, '..', 'docs');
const outputFile = path.join(__dirname, '..', 'docs', 'Support.csv');

function findLatestSourceFile() {
    const files = fs.readdirSync(docsDir);
    const pattern = /打赏.*链接小票单记录.*\.csv$/;

    const matchedFiles = files
        .filter(f => pattern.test(f))
        .map(f => {
            const filePath = path.join(docsDir, f);
            return {
                path: filePath,
                mtime: fs.statSync(filePath).mtime
            };
        })
        .sort((a, b) => b.mtime - a.mtime);

    return matchedFiles.length > 0 ? matchedFiles[0].path : null;
}

function processCsv() {
    const inputFile = findLatestSourceFile();

    if (!inputFile) {
        console.error(`No matching source file found in ${docsDir} (Keyword: "打赏" & "链接小票单记录")`);
        process.exit(1);
    }

    console.log(`Processing latest source file: ${inputFile}`);

    const content = fs.readFileSync(inputFile, 'utf8');
    const lines = content.split(/\r?\n/);

    // 真正的表头在第 8 行 (index 7)，数据从第 9 行 (index 8) 开始
    const dataLines = lines.slice(8).filter(line => line.trim() !== '');

    const dataRows = [];

    dataLines.forEach(line => {
        const fields = parseCsvLine(line);
        if (fields.length >= 9) {
            const time = fields[0].trim();
            const item = fields[3].trim();
            const amountStr = fields[4].trim();
            const message = fields[6].trim();
            const name = fields[7].trim();
            const note = fields[8].trim();

            // 处理金额：去掉 ¥，转为数字用于排序
            const amountValue = parseFloat(amountStr.replace(/[^\d.-]/g, '')) || 0;

            dataRows.push({
                time,
                item,
                amountVal: amountValue, // 用于排序
                amountStr: amountValue.toFixed(2), // 用于显示
                unit: '¥',
                message,
                name,
                note
            });
        }
    });

    // 排序逻辑：金额由大到小；金额一致则时间新在前（降序）
    dataRows.sort((a, b) => {
        if (b.amountVal !== a.amountVal) {
            return b.amountVal - a.amountVal;
        }
        return b.time.localeCompare(a.time);
    });

    // 构建输出内容
    const result = [];
    result.push('收款时间,收款项,金额,单位,留言,昵称,备注');

    dataRows.forEach(row => {
        const rowStr = [
            formatCsvField(row.time),
            formatCsvField(row.item),
            formatCsvField(row.amountStr),
            formatCsvField(row.unit),
            formatCsvField(row.message),
            formatCsvField(row.name),
            formatCsvField(row.note)
        ].join(',');
        result.push(rowStr);
    });

    fs.writeFileSync(outputFile, result.join('\n') + '\n', 'utf8');
    console.log(`Successfully processed and sorted. Saved to ${outputFile}`);
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
    if (!field) return '';
    if (typeof field !== 'string') field = String(field);
    field = field.trim();
    if (field.includes(',') || field.includes('"') || field.includes('\n')) {
        return `"${field.replace(/"/g, '""')}"`;
    }
    return field;
}

processCsv();
