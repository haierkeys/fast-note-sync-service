/**
 * 打赏支持者名单处理与翻译脚本 (Support List Processor & Translator)
 *
 * 核心功能:
 * 1. 读取 清洗过数据的 `docs/Support.csv`，提取所有的“收款项”和“留言”。
 * 2. 采用 MD5 字典缓存机制 (`docs/.support-translate-dict.json`)，仅翻译新增的条目，大幅节省 API 费用和时间。
 * 3. 采用 OpenAI 兼容的大模型 API 进行批次翻译，支持流式响应。
 * 4. 自动生成多语言版本的 `Support.{lang}.json` 以及 `Support.{lang}.md` 文件。
 *
 * 用法 (Usage):
 *   # 默认运行 (读取 .env 配置)
 *   node scripts/process_support.mjs
 *
 *   # 临时指定运行模型
 *   node scripts/process_support.mjs --model Qwen/Qwen3.6-35B-A3B
 *
 * 环境变量 (配置在项目根目录的 .env 文件中):
 *   OPENAI_API_KEY      (必填) 你的大模型 API 密钥
 *   OPENAI_BASE_URL     (可选) 代理地址，默认: https://www.dmxapi.cn/v1
 *   OPENAI_MODEL        (可选) 默认使用的模型，默认: qwen3.5-27b
 */
import { fileURLToPath } from "node:url";
import crypto from "node:crypto";
import path from "node:path";
import fs from "node:fs";


const __dirname = path.dirname(fileURLToPath(import.meta.url));
const base_dir = path.resolve(__dirname, "..");
const input_csv = path.join(base_dir, "docs", "Support.csv");
const DICT_FILE = path.join(base_dir, "docs", ".support-translate-dict.json");
const ENV_FILE = path.join(base_dir, ".env");

if (fs.existsSync(ENV_FILE)) {
  const envContent = fs.readFileSync(ENV_FILE, "utf-8");
  envContent.split(/\r?\n/).forEach(line => {
    const trimmed = line.trim();
    if (trimmed && !trimmed.startsWith("#")) {
      const match = trimmed.match(/^([^=]+)=(.*)$/);
      if (match) {
        const key = match[1].trim();
        const value = match[2].trim().replace(/^(['"])(.*)\1$/, "$2");
        if (!process.env[key]) process.env[key] = value;
      }
    }
  });
}

const args = process.argv.slice(2);
function getArg(name) {
  const idx = args.indexOf(`--${name}`);
  return idx !== -1 && args[idx + 1] ? args[idx + 1] : null;
}

const OPENAI_API_KEY = process.env.OPENAI_API_KEY;
const OPENAI_BASE_URL = process.env.OPENAI_BASE_URL || "https://www.dmxapi.cn/v1";
const MODEL = getArg("model") || process.env.OPENAI_MODEL || "qwen3.5-27b";

if (!OPENAI_API_KEY) {
  console.warn("⚠️  Warning: OPENAI_API_KEY is not set. Translation API will fail if there are new items to translate.");
}

const KEY_MAP = {
  "收款时间": "time",
  "收款项": "item",
  "金额": "amount",
  "单位": "unit",
  "留言": "message",
  "昵称": "name"
};

const LANG_CONFIG = {
  "en": {
    name: "English",
    title: "Supporters List",
    quote: "Thank you very much for supporting this project! Every donation is the driving force for my continuous maintenance and iteration. ❤️",
    listTitle: "Acknowledgement List",
    headers: ["Time", "Item", "Amount", "Name", "Message"],
    footer: "Last updated on: "
  },
  "zh-CN": {
    name: "简体中文",
    title: "支持者名单 (Thanks to Supporters)",
    quote: "非常感谢大家对本项目的支持！每一份打赏都是我持续维护和迭代的动力。 ❤️",
    listTitle: "致谢列表",
    headers: ["收款时间", "收款项", "金额", "昵称", "留言"],
    footer: "本数据最后更新于："
  },
  "zh-TW": {
    name: "繁體中文",
    title: "支持者名單 (Thanks to Supporters)",
    quote: "非常感謝大家對本項目的支持！每一份打賞都是我持續維護和迭代的動力。 ❤️",
    listTitle: "致謝列表",
    headers: ["收款時間", "收款項", "金額", "昵稱", "留言"],
    footer: "本數據最後更新於："
  },
  "ja": {
    name: "日本語",
    title: "サポーターリスト",
    quote: "このプロジェクトを応援していただき、誠にありがとうございます！皆様からのご支援は、継続的なメンテナンスと開発の原動力となっています。 ❤️",
    listTitle: "謝辞リスト",
    headers: ["受領时间", "项目", "金额", "昵称", "メッセージ"],
    footer: "最終更新日："
  },
  "ko": {
    name: "한국어",
    title: "후원자 명단",
    quote: "이 프로젝트를 지원해 주셔서 정말 감사합니다! 여러분의 모든 후원은 지속적인 유지보수와 개발의 원동력이 됩니다. ❤️",
    listTitle: "감사 명단",
    headers: ["수령 시간", "항목", "금액", "닉네임", "메시지"],
    footer: "마지막 업데이트:"
  }
};

const LANG_NAMES = {
  "en": "English",
  "zh-CN": "Simplified Chinese",
  "zh-TW": "Traditional Chinese",
  "ja": "Japanese",
  "ko": "Korean"
};

function md5(str) {
  return crypto.createHash("md5").update(str, "utf-8").digest("hex");
}

function loadDict() {
  if (fs.existsSync(DICT_FILE)) {
    try {
      return JSON.parse(fs.readFileSync(DICT_FILE, "utf-8"));
    } catch (e) {
      console.warn(`⚠️ Dict load failed: ${e.message}`);
    }
  }
  return {};
}

function saveDict(dict) {
  fs.writeFileSync(DICT_FILE, JSON.stringify(dict, null, 2), "utf-8");
}

function parseCSV(content) {
  const lines = content.split(/\r?\n/).filter(l => l.trim());
  if (lines.length === 0) return [];

  function splitLine(line) {
    const result = [];
    let cur = "";
    let inQuote = false;
    for (let i = 0; i < line.length; i++) {
      const c = line[i];
      if (c === '"') {
        if (inQuote && line[i + 1] === '"') {
          cur += '"';
          i++;
        } else {
          inQuote = !inQuote;
        }
      } else if (c === ',' && !inQuote) {
        result.push(cur);
        cur = "";
      } else {
        cur += c;
      }
    }
    result.push(cur);
    return result;
  }

  const headers = splitLine(lines[0]);
  const data = [];
  for (let i = 1; i < lines.length; i++) {
    const values = splitLine(lines[i]);
    const row = {};
    for (let j = 0; j < headers.length; j++) {
      row[headers[j]] = values[j] || "";
    }
    data.push(row);
  }
  return data;
}

async function callOpenAIStream(messages, onToken) {
  const res = await fetch(`${OPENAI_BASE_URL}/chat/completions`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${OPENAI_API_KEY}`,
    },
    body: JSON.stringify({
      model: MODEL,
      temperature: 0.3,
      stream: true,
      messages,
      thinking: { type: "disabled" },
    }),
  });

  if (!res.ok) {
    const body = await res.text();
    throw new Error(`API error: ${res.status} ${body}`);
  }

  let fullContent = "";
  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    const lines = buffer.split("\n");
    buffer = lines.pop() || "";
    for (const line of lines) {
      const trimmed = line.trim();
      if (!trimmed || !trimmed.startsWith("data: ")) continue;
      const data = trimmed.slice(6);
      if (data === "[DONE]") continue;
      try {
        const parsed = JSON.parse(data);
        const delta = parsed.choices?.[0]?.delta?.content;
        if (delta) {
          fullContent += delta;
          onToken?.(delta, fullContent);
        }
      } catch { }
    }
  }
  return fullContent;
}

function extractJSON(content) {
  let jsonStr = content.trim();
  const fenceMatch = jsonStr.match(/```(?:json)?\s*([\s\S]*?)```/);
  if (fenceMatch) {
    jsonStr = fenceMatch[1].trim();
  } else {
    const firstBrace = jsonStr.indexOf("{");
    const lastBrace = jsonStr.lastIndexOf("}");
    if (firstBrace !== -1 && lastBrace !== -1 && lastBrace > firstBrace) {
      jsonStr = jsonStr.substring(firstBrace, lastBrace + 1);
    }
  }
  return JSON.parse(jsonStr);
}

function formatDuration(ms) {
  if (ms < 1000) return `${ms}ms`;
  const s = Math.floor(ms / 1000);
  if (s < 60) return `${s}s`;
  const m = Math.floor(s / 60);
  return `${m}m${s % 60}s`;
}

function writeProgress(text) {
  if (process.stdout.clearLine) {
    process.stdout.clearLine(0);
    process.stdout.cursorTo(0);
  }
  process.stdout.write(text);
}

async function translateBatch(texts, sourceLangName, targetLangName) {
  if (texts.length === 0) return {};
  const entries = texts.map(t => [md5(t), t]);
  const jsonPayload = JSON.stringify(Object.fromEntries(entries), null, 2);
  const systemPrompt = `You are a professional translator. Translate the following JSON values from ${sourceLangName} to ${targetLangName}.
CRITICAL RULES:
1. Return ONLY a valid JSON object.
2. The keys MUST be exactly the same as the input JSON.
3. Keep technical terms unchanged if applicable.
4. For "Traditional Chinese" target, convert to Traditional Chinese accurately.
5. If a value is just a name, keep it natural or transliterate appropriately.`;

  const batchStart = Date.now();
  let charCount = 0;
  console.log(`     🔗 正在连接 API 发送请求 (请耐心等待响应)...`);
  const content = await callOpenAIStream(
    [
      { role: "system", content: systemPrompt },
      { role: "user", content: jsonPayload },
    ],
    (delta, full) => {
      charCount += delta.length;
      const elapsed = formatDuration(Date.now() - batchStart);
      const keysReceived = (full.match(/"[^"]+"\s*:/g) || []).length;
      writeProgress(`     ⏳ 流式接收中... ${keysReceived}/${texts.length} keys | ${charCount} chars | ${elapsed}`);
    }
  );

  process.stdout.write("\n");
  return extractJSON(content);
}

function generateJson(data, translationMap, langCode) {
  const outputData = [];
  for (const row of data) {
    const newRow = {};
    for (const [cnKey, enKey] of Object.entries(KEY_MAP)) {
      let val = row[cnKey] || "";
      if (['收款项', '留言'].includes(cnKey) && val && val !== '-') {
        val = translationMap[val] || val;
      }
      newRow[enKey] = val ? val : (cnKey === '留言' ? '-' : val);
    }
    outputData.push(newRow);
  }
  const filePath = path.join(base_dir, 'docs', `Support.${langCode}.json`);
  fs.writeFileSync(filePath, JSON.stringify(outputData, null, 2), 'utf-8');
  console.log(`  💾 Saved JSON: ${filePath}`);
}

function generateMd(data, langCode, translationMap) {
  const conf = LANG_CONFIG[langCode];
  let md = `# ${conf.title}\n\n`;
  md += `> ${conf.quote}\n\n`;
  md += `### 📜 ${conf.listTitle}\n\n`;
  md += `| ${conf.headers.join(' | ')} |\n`;
  md += `| ${conf.headers.map(() => ':---').join(' | ')} |\n`;

  for (const row of data) {
    const time = row['收款时间'] || '';
    const itemOrig = (row['收款项'] || '').trim();
    const item = translationMap[itemOrig] || itemOrig;
    const amountStr = row['金额'] || '';
    const unitStr = row['单位'] || '';
    let displayUnit = unitStr;
    if (displayUnit.toUpperCase() === 'USD') {
      displayUnit = '$';
    }
    const amount = amountStr ? `**${displayUnit}${amountStr}**` : '';
    const name = row['昵称'] || '';
    const msgOrig = (row['留言'] || '').trim();
    const msg = (msgOrig && msgOrig !== '-') ? (translationMap[msgOrig] || msgOrig) : '-';

    const rowValues = [];
    conf.headers.forEach(header => {
      if (['收款时间', '收款時間', '受領時間', '受領时间', 'Time', '수령 시간'].includes(header)) {
        rowValues.push(time);
      } else if (['收款项', '收款項', '項目', '项目', 'Item', '항목'].includes(header)) {
        rowValues.push(item);
      } else if (['金额', '金額', 'Amount', '금액'].includes(header)) {
        rowValues.push(amount);
      } else if (['昵称', '昵稱', 'ニックネーム', 'Name', '닉네임'].includes(header)) {
        rowValues.push(name);
      } else if (['留言', 'メッセージ', 'Message', '메시지'].includes(header)) {
        rowValues.push(msg);
      } else {
        rowValues.push('');
      }
    });

    md += `| ${rowValues.join(' | ')} |\n`;
  }

  const now = new Date();
  const timestamp = langCode.startsWith('zh') || langCode === 'ja' || langCode === 'ko'
      ? now.toLocaleString('zh-CN', { hour12: false })
      : now.toUTCString();

  md += `\n\n--- \n*${conf.footer}${timestamp}*`;

  const filePath = path.join(base_dir, 'docs', `Support.${langCode}.md`);
  fs.writeFileSync(filePath, md, 'utf-8');
  console.log(`  📝 Saved MD: ${filePath}`);
}

function parseDate(dateStr) {
  if (!dateStr) return 0;
  const parts = dateStr.split(' ');
  const datePart = parts[0];
  const timePart = parts[1] || '00:00:00';

  const dateSubparts = datePart.split(/[-/]/);
  if (dateSubparts.length === 3) {
    if (dateSubparts[0].length === 4) {
      // YYYY/MM/DD or YYYY-MM-DD
      return new Date(`${dateSubparts[0]}-${dateSubparts[1]}-${dateSubparts[2]}T${timePart}`).getTime() || 0;
    } else if (dateSubparts[2].length === 4) {
      // MM/DD/YYYY or DD/MM/YYYY
      return new Date(`${dateSubparts[2]}-${dateSubparts[0]}-${dateSubparts[1]}T${timePart}`).getTime() || 0;
    }
  }
  const parsed = new Date(dateStr).getTime();
  return isNaN(parsed) ? 0 : parsed;
}

function getCNYValue(row) {
  const amount = parseFloat(row['金额']) || 0;
  const unit = (row['单位'] || '').toUpperCase();
  if (unit === 'USD' || unit === '$') {
    return amount * 6.81; // 粗略以 6.81 汇率折合人民币进行排序
  }
  return amount;
}

async function main() {
  const kofi_csv = path.join(base_dir, "docs", "Support_kofi.csv");
  let cnData = [];
  let kofiData = [];

  if (fs.existsSync(input_csv)) {
    const csvContent = fs.readFileSync(input_csv, "utf-8");
    cnData = parseCSV(csvContent);
  } else {
    console.warn(`⚠️  Warning: ${input_csv} not found.`);
  }

  if (fs.existsSync(kofi_csv)) {
    console.log(`📊 Found Ko-fi support data: ${kofi_csv}, loading...`);
    const kofiContent = fs.readFileSync(kofi_csv, "utf-8");
    kofiData = parseCSV(kofiContent);
  }

  const data = cnData.concat(kofiData);

  if (data.length === 0) {
    console.error("❌ Error: No supporter data found in either Support.csv or Support_kofi.csv");
    return;
  }

  // 统一按赞助金额降序排序（外币折合人民币），金额相同时按时间降序排序
  data.sort((a, b) => {
    const valA = getCNYValue(a);
    const valB = getCNYValue(b);
    if (valB !== valA) {
      return valB - valA;
    }
    const timeA = parseDate(a['收款时间']);
    const timeB = parseDate(b['收款时间']);
    return timeB - timeA;
  });

  // 1. 提取 简体中文源 (Support.csv) 文本
  const cnTextsSet = new Set();
  for (const row of cnData) {
    for (const k of ['收款项', '留言']) {
      if (row[k] && row[k] !== '-') {
        cnTextsSet.add(row[k]);
      }
    }
  }
  const cnTextsList = Array.from(cnTextsSet);

  // 2. 提取 英文源 (Support_kofi.csv) 文本
  const kofiTextsSet = new Set();
  for (const row of kofiData) {
    for (const k of ['收款项', '留言']) {
      if (row[k] && row[k] !== '-') {
        kofiTextsSet.add(row[k]);
      }
    }
  }
  const kofiTextsList = Array.from(kofiTextsSet);

  console.log(`📊 Support.csv (中文源) 提取到 ${cnTextsList.length} 条需要处理的文本`);
  console.log(`📊 Ko-fi (英文源) 提取到 ${kofiTextsList.length} 条需要处理的文本`);

  const dict = loadDict();
  let dictUpdated = false;

  const targetLangs = Object.keys(LANG_CONFIG);
  const BATCH_SIZE = 20;

  // 🚩 第一阶段：翻译 Ko-fi 英文源记录 (目标语言: zh-CN, zh-TW, ja, ko)
  if (kofiTextsList.length > 0) {
    const kofiTargetLangs = ['zh-CN', 'zh-TW', 'ja', 'ko'];
    const totalKofiBatches = Math.ceil(kofiTextsList.length / BATCH_SIZE);

    let hasKofiTranslations = false;
    for (const lang of kofiTargetLangs) {
      kofiTextsList.forEach(t => {
        const hash = md5(t);
        if (!dict[hash] || !dict[hash][lang]) {
          hasKofiTranslations = true;
        }
      });
    }

    if (hasKofiTranslations) {
      console.log(`\n🚀 开始翻译 Ko-fi (源语言: 英文, 共 ${totalKofiBatches} 个批次)...`);
    }

    let batchCount = 0;
    for (let i = 0; i < kofiTextsList.length; i += BATCH_SIZE) {
      const batchTexts = kofiTextsList.slice(i, i + BATCH_SIZE);
      batchCount++;
      let batchNeedSave = false;

      console.log(`\n📦 --- Ko-fi (英文源) 批次 ${batchCount}/${totalKofiBatches} ---`);

      // 批次内翻译成所有目标语言 (Batch-Sequential)
      for (const lang of kofiTargetLangs) {
        const toTranslate = [];
        batchTexts.forEach(t => {
          const hash = md5(t);
          if (!dict[hash] || !dict[hash][lang]) {
            toTranslate.push(t);
          }
        });

        if (toTranslate.length > 0) {
          console.log(`   ⏳ 正在翻译到 [${LANG_CONFIG[lang].name}] (${toTranslate.length} 条增量)...`);
          try {
            const result = await translateBatch(toTranslate, "English", LANG_NAMES[lang]);
            toTranslate.forEach(t => {
              const hash = md5(t);
              const translated = result[hash];
              if (translated) {
                if (!dict[hash]) dict[hash] = {};
                dict[hash][lang] = translated;
                batchNeedSave = true;
                dictUpdated = true;
              }
            });
            await new Promise(resolve => setTimeout(resolve, 500));
          } catch (e) {
            console.error(`   ❌ 翻译到 [${LANG_CONFIG[lang].name}] 失败: ${e.message}`);
          }
        } else {
          console.log(`   ✅ [${LANG_CONFIG[lang].name}] 缓存全部命中`);
        }
      }

      if (batchNeedSave) {
        saveDict(dict);
        console.log(`💾 批次 ${batchCount} 翻译已实时保存`);
      }
    }
  }

  // 🚩 第二阶段：翻译 Support.csv 中文源记录 (目标语言: en, zh-TW, ja, ko)
  if (cnTextsList.length > 0) {
    const cnTargetLangs = ['en', 'zh-TW', 'ja', 'ko'];
    const totalCnBatches = Math.ceil(cnTextsList.length / BATCH_SIZE);

    let hasCnTranslations = false;
    for (const lang of cnTargetLangs) {
      cnTextsList.forEach(t => {
        const hash = md5(t);
        if (!dict[hash] || !dict[hash][lang]) {
          hasCnTranslations = true;
        }
      });
    }

    if (hasCnTranslations) {
      console.log(`\n🚀 开始翻译 Support.csv (源语言: 简体中文, 共 ${totalCnBatches} 个批次)...`);
    }

    let batchCount = 0;
    for (let i = 0; i < cnTextsList.length; i += BATCH_SIZE) {
      const batchTexts = cnTextsList.slice(i, i + BATCH_SIZE);
      batchCount++;
      let batchNeedSave = false;

      console.log(`\n📦 --- Support.csv (中文源) 批次 ${batchCount}/${totalCnBatches} ---`);

      // 批次内翻译成所有目标语言 (Batch-Sequential)
      for (const lang of cnTargetLangs) {
        const toTranslate = [];
        batchTexts.forEach(t => {
          const hash = md5(t);
          if (!dict[hash] || !dict[hash][lang]) {
            toTranslate.push(t);
          }
        });

        if (toTranslate.length > 0) {
          console.log(`   ⏳ 正在翻译到 [${LANG_CONFIG[lang].name}] (${toTranslate.length} 条增量)...`);
          try {
            const result = await translateBatch(toTranslate, "Simplified Chinese", LANG_NAMES[lang]);
            toTranslate.forEach(t => {
              const hash = md5(t);
              const translated = result[hash];
              if (translated) {
                if (!dict[hash]) dict[hash] = {};
                dict[hash][lang] = translated;
                batchNeedSave = true;
                dictUpdated = true;
              }
            });
            await new Promise(resolve => setTimeout(resolve, 500));
          } catch (e) {
            console.error(`   ❌ 翻译到 [${LANG_CONFIG[lang].name}] 失败: ${e.message}`);
          }
        } else {
          console.log(`   ✅ [${LANG_CONFIG[lang].name}] 缓存全部命中`);
        }
      }

      if (batchNeedSave) {
        saveDict(dict);
        console.log(`💾 批次 ${batchCount} 翻译已实时保存`);
      }
    }
  }

  if (dictUpdated) {
    console.log(`\n✨ 翻译环节结束，开始生成多语言文件...`);
  } else {
    console.log(`\n✨ 已全部命中字典缓存，开始生成多语言文件...`);
  }

  for (const lang of targetLangs) {
    console.log(`\n🌐 生成 ${lang} (${LANG_CONFIG[lang].name}) 文件...`);
    let tMap = {};
    const allTexts = cnTextsList.concat(kofiTextsList);
    
    allTexts.forEach(t => {
      const hash = md5(t);
      const isKofiText = kofiTextsSet.has(t);
      const isCnText = cnTextsSet.has(t);

      if (lang === 'zh-CN' && isCnText) {
        tMap[t] = t;
      } else if (lang === 'en' && isKofiText) {
        tMap[t] = t;
      } else {
        if (dict[hash] && dict[hash][lang]) {
          tMap[t] = dict[hash][lang];
        } else {
          tMap[t] = t;
        }
      }
    });

    generateJson(data, tMap, lang);
    generateMd(data, lang, tMap);
  }

  console.log("\n✅ 全部处理完成！");
}

main().catch(err => {
  console.error("❌ Fatal error:", err);
  process.exit(1);
});
