const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const inputFile = path.join(__dirname, '..', 'docs', 'Support.csv');

// i18n 配置
const i18n = {
    'zh-CN': {
        filename: 'Support.zh-CN.md',
        title: '支持者名单 (Thanks to Supporters)',
        quote: '非常感谢大家对本项目的支持！每一份打赏都是我持续维护和迭代的动力。 ❤️',
        listTitle: '致谢列表',
        headers: ['收款时间', '收款项', '金额', '昵称', '留言', '备注'],
        footer: '本数据最后更新于：',
        langName: '简体中文',
        target: 'zh-CN'
    },
    'zh-TW': {
        filename: 'Support.zh-TW.md',
        title: '支持者名單 (Thanks to Supporters)',
        quote: '非常感謝大家對本項目的支持！每一份打賞都是我持續維護和迭代的動力。 ❤️',
        listTitle: '致謝列表',
        headers: ['收款時間', '收款項', '金額', '昵稱', '留言', '備註'],
        footer: '本數據最後更新於：',
        langName: '繁體中文',
        target: 'zh-TW'
    },
    'en': {
        filename: 'Support.en.md',
        title: 'Supporters List',
        quote: 'Thank you very much for supporting this project! Every donation is the driving force for my continuous maintenance and iteration. ❤️',
        listTitle: 'Acknowledgement List',
        headers: ['Time', 'Item', 'Amount', 'Name', 'Message', 'Note'],
        footer: 'Last updated on: ',
        langName: 'English',
        target: 'en'
    },
    'ja': {
        filename: 'Support.ja.md',
        title: 'サポーターリスト',
        quote: 'このプロジェクトを応援していただき、誠にありがとうございます！皆様からのご支援は、継続的なメンテナンスと開発の原動力となっています。 ❤️',
        listTitle: '謝辞リスト',
        headers: ['受領時間', '項目', '金額', 'ニックネーム', 'メッセージ', '備考'],
        footer: '最終更新日：',
        langName: '日本語',
        target: 'ja'
    },
    'ko': {
        filename: 'Support.ko.md',
        title: '후원자 명단',
        quote: '이 프로젝트를 지원해 주셔서 정말 감사합니다! 여러분의 모든 후원은 지속적인 유지보수와 개발의 원동력이 됩니다. ❤️',
        listTitle: '감사 명단',
        headers: ['수령 시간', '항목', '금액', '닉네임', '메시지', '비고'],
        footer: '마지막 업데이트:',
        langName: '한국어',
        target: 'ko'
    }
};

/**
 * 批量翻译函数 - 调用独立的 Python 翻译辅助脚本
 */
function translateBatch(texts, targetLang) {
    if (texts.length === 0 || targetLang === 'zh-CN') {
        const result = {};
        texts.forEach(t => result[t] = t);
        return result;
    }

    const tmpIn = path.join(__dirname, '..', `tmp_translate_in_${targetLang}.json`);
    const helperScript = path.join(__dirname, 'translate_support.py');
    const { spawnSync } = require('child_process');
    try {
        // 1. 将待翻译文本写入临时文件 (避免 Shell 转义问题)
        fs.writeFileSync(tmpIn, JSON.stringify(texts), 'utf8');

        // 2. 调用 Python 脚本
        const target = targetLang === 'zh-TW' ? 'zh-TW' : targetLang;

        // 使用 spawnSync 避免 shell 转义和路径空格问题
        const result = spawnSync('python', [helperScript, tmpIn, target], {
            encoding: 'utf8',
            maxBuffer: 10 * 1024 * 1024 // 增加缓冲区到 10MB
        });

        if (result.status !== 0) {
            throw new Error(`Python script exited with status ${result.status}: ${result.stderr}`);
        }

        const translatedArray = JSON.parse(result.stdout.trim());

        const resultMap = {};
        texts.forEach((original, index) => {
            resultMap[original] = translatedArray[index] || original;
        });

        // 3. 清理临时文件
        if (fs.existsSync(tmpIn)) fs.unlinkSync(tmpIn);

        return resultMap;
    } catch (err) {
        console.warn(`Translation to ${targetLang} failed: ${err.message}. Fallback to original.`);
        if (fs.existsSync(tmpIn)) fs.unlinkSync(tmpIn);
        const resultMap = {};
        texts.forEach(t => resultMap[t] = t);
        return resultMap;
    }
}

async function genMarkdown() {
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

    const csvHeaders = parseCsvLine(lines[0]);
    const dataRows = lines.slice(1).map(line => {
        const fields = parseCsvLine(line);
        const obj = {};
        csvHeaders.forEach((h, i) => {
            obj[h] = fields[i] || '';
        });
        return obj;
    });

    // 收集所有需要翻译的文本 (收款项和留言)
    const uniqueTexts = new Set();
    dataRows.forEach(row => {
        if (row['收款项']) uniqueTexts.add(row['收款项']);
        if (row['留言'] && row['留言'] !== '-') uniqueTexts.add(row['留言']);
        if (row['备注'] && row['备注'] !== '-') uniqueTexts.add(row['备注']);
    });
    const textsToTranslate = Array.from(uniqueTexts);

    // 为每种语言生成文档
    for (const lang of Object.keys(i18n)) {
        const config = i18n[lang];
        const outputFilePath = path.join(__dirname, '..', 'docs', config.filename);

        console.log(`[${config.langName}] Translating...`);
        const translationMap = translateBatch(textsToTranslate, lang);
        console.log(`[${config.langName}] Translation complete. Sample: "${textsToTranslate[0]}" -> "${translationMap[textsToTranslate[0]] || 'N/A'}"`);

        let md = `# ${config.title}\n\n`;
        md += `> ${config.quote}\n\n`;

        md += `### 📜 ${config.listTitle}\n\n`;
        md += `| ${config.headers.join(' | ')} |\n`;
        md += `| ${config.headers.map(() => ':---').join(' | ')} |\n`;

        dataRows.forEach((row, index) => {
            const displayTime = row['收款时间'] || '';
            const rawItem = (row['收款项'] || '').trim();
            const displayItem = translationMap[rawItem] || rawItem;

            const displayAmount = `**${row['单位'] || ''}${row['金额'] || ''}**`;
            const displayName = row['昵称'] || '';

            const rawMessage = (row['留言'] || '').trim();
            const displayMessage = (rawMessage === '-' || !rawMessage) ? '-' : (translationMap[rawMessage] || rawMessage);

            const rawNote = (row['备注'] || '').trim();
            const displayNote = (rawNote === '-' || !rawNote) ? '-' : (translationMap[rawNote] || rawNote);

            md += `| ${displayTime} | ${displayItem} | ${displayAmount} | ${displayName} | ${displayMessage} | ${displayNote} |\n`;
        });

        const now = new Date();
        const timestamp = lang.startsWith('zh') || lang === 'ja' || lang === 'ko'
            ? now.toLocaleString('zh-CN', { hour12: false })
            : now.toUTCString();

        md += `\n\n--- \n*${config.footer}${timestamp}*`;

        fs.writeFileSync(outputFilePath, md, 'utf8');
        console.log(`[${config.langName}] Generated: ${config.filename}`);
    }
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
