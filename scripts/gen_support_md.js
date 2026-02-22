const fs = require('fs');
const path = require('path');

const inputFile = path.join(__dirname, '..', 'docs', 'Support.csv');
const outputFile = path.join(__dirname, '..', 'docs', 'Support.zh-CN.md');

function genMarkdown() {
    if (!fs.existsSync(inputFile)) {
        console.error(`Input file not found: ${inputFile}`);
        process.exit(1);
    }

    const content = fs.readFileSync(inputFile, 'utf8');
    const lines = content.split(/\r?\n/).filter(line => line.trim() !== '');

    if (lines.length < 2) {
        console.error("CSV file is empty or has no data.");
        return;
    }

    const headers = parseCsvLine(lines[0]);
    const dataRows = lines.slice(1).map(line => {
        const fields = parseCsvLine(line);
        const obj = {};
        headers.forEach((h, i) => {
            obj[h] = fields[i] || '';
        });
        return obj;
    });

    let md = `# 支持者名单 (Thanks to Supporters)\n\n`;
    md += `> 非常感谢大家对本项目的支持！每一份打赏都是我持续维护和迭代的动力。 ❤️\n\n`;

    md += `### 📜 致谢列表\n\n`;
    md += `| 收款时间 | 收款项 | 金额 | 昵称 | 留言 | 备注 |\n`;
    md += `| :--- | :--- | :--- | :--- | :--- | :--- |\n`;

    dataRows.forEach(row => {
        // 金额带上符号显示
        const displayAmount = `${row['单位']}${row['金额']}`;
        md += `| ${row['收款时间']} | ${row['收款项']} | **${displayAmount}** | ${row['昵称']} | ${row['留言'] || '-'} | ${row['备注'] || '-'} |\n`;
    });

    md += `\n\n--- \n*本数据最后更新于：${new Date().toLocaleString('zh-CN', { hour12: false })}*`;

    fs.writeFileSync(outputFile, md, 'utf8');
    console.log(`Successfully generated Markdown doc at ${outputFile}`);
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
    return fields.map(f => f.replace(/^"|"$/g, '').replace(/""/g, '"').trim());
}

genMarkdown();
