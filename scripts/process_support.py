import csv
import json
import os
import sys
from datetime import datetime
from deep_translator import GoogleTranslator

# 配置路径
base_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
input_csv = os.path.join(base_dir, 'docs', 'Support.csv')

# 键名映射 (CN -> EN)
KEY_MAP = {
    '收款时间': 'time',
    '收款项': 'item',
    '金额': 'amount',
    '单位': 'unit',
    '留言': 'message',
    '昵称': 'name',
    '备注': 'note'
}

# 语言配置
LANG_CONFIG = {
    'en': {'name': 'English', 'md_title': 'Supporters List', 'md_headers': ['Time', 'Item', 'Amount', 'Name', 'Message', 'Note'], 'md_footer': 'Last updated: '},
    'zh-CN': {'name': '简体中文', 'md_title': '支持者名单', 'md_headers': ['收款时间', '收款项', '金额', '昵称', '留言', '备注'], 'md_footer': '最后更新于：'},
    'zh-TW': {'name': '繁體中文', 'md_title': '支持者名單', 'md_headers': ['收款時間', '收款項', '金額', '昵稱', '留言', '備註'], 'md_footer': '最後更新於：'},
    'ja': {'name': '日本語', 'md_title': 'サポーターリスト', 'md_headers': ['受領時間', '項目', '金額', '名前', 'メッセージ', '備考'], 'md_footer': '最終更新：'},
    'ko': {'name': '한국어', 'md_title': '후원자 명단', 'md_headers': ['시간', '항목', '금액', '이름', '메시지', '비고'], 'md_footer': '마지막 업데이트：'}
}

def translate_texts(texts, target_lang):
    """批量翻译文本"""
    if not texts or target_lang == 'zh-CN':
        return {text: text for text in texts}

    translator = GoogleTranslator(source='auto', target=target_lang)
    to_translate = [t for t in texts if t and t.strip() and t.strip() != '-']
    if not to_translate:
        return {text: text for text in texts}

    try:
        translated_list = translator.translate_batch(to_translate)
        t_map = dict(zip(to_translate, translated_list))
        return {text: t_map.get(text, text) for text in texts}
    except Exception as e:
        print(f"Translation to {target_lang} failed: {e}. Fallback to original.", file=sys.stderr)
        return {text: text for text in texts}

def generate_json(data, translation_map, lang_code):
    """生成 JSON 数据 (键名统一为英文)"""
    output_data = []
    for row in data:
        new_row = {}
        for cn_key, en_key in KEY_MAP.items():
            val = row.get(cn_key, '')
            # 翻译内容字段
            if cn_key in ['收款项', '留言', '备注'] and val and val != '-':
                val = translation_map.get(val, val)
            new_row[en_key] = val if val else ('-' if cn_key in ['留言', '备注'] else val)
        output_data.append(new_row)

    file_path = os.path.join(base_dir, 'docs', f'Support.{lang_code}.json')
    with open(file_path, 'w', encoding='utf-8') as f:
        json.dump(output_data, f, ensure_ascii=False, indent=2)
    print(f"Saved JSON: {file_path}")

def generate_md(data, lang_code, translation_map):
    """生成 Markdown 文件"""
    conf = LANG_CONFIG.get(lang_code)
    md = f"# {conf['md_title']}\n\n"
    md += f"| {' | '.join(conf['md_headers'])} |\n"
    md += f"| {' | '.join([':---']*len(conf['md_headers']))} |\n"

    for row in data:
        time = row.get('收款时间', '')
        item = translation_map.get(row.get('收款项', ''), row.get('收款项', ''))
        amount = f"**{row.get('单位', '')}{row.get('金额', '')}**"
        name = row.get('昵称', '')
        msg = row.get('留言', '')
        msg = translation_map.get(msg, msg) if msg and msg != '-' else '-'
        note = row.get('备注', '')
        note = translation_map.get(note, note) if note and note != '-' else '-'

        md += f"| {time} | {item} | {amount} | {name} | {msg} | {note} |\n"

    timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
    md += f"\n\n--- \n*{conf['md_footer']}{timestamp}*"

    file_path = os.path.join(base_dir, 'docs', f'Support.{lang_code}.md')
    with open(file_path, 'w', encoding='utf-8') as f:
        f.write(md)
    # print(f"Saved MD: {file_path}")

def main():
    if not os.path.exists(input_csv):
        print(f"Error: {input_csv} not found.")
        return

    data = []
    unique_texts = set()
    try:
        with open(input_csv, 'r', encoding='utf-8') as f:
            reader = csv.DictReader(f)
            for row in reader:
                data.append(row)
                for k in ['收款项', '留言', '备注']:
                    if row.get(k) and row[k] != '-':
                        unique_texts.add(row[k])
    except Exception as e:
        print(f"Error reading CSV: {e}")
        return

    texts_list = list(unique_texts)

    for lang in LANG_CONFIG.keys():
        print(f"Processing {lang}...")
        t_map = translate_texts(texts_list, lang)
        generate_json(data, t_map, lang)
        # 为特定语言生成 MD
        if lang in ['zh-CN', 'zh-TW', 'en']:
            generate_md(data, lang, t_map)

if __name__ == "__main__":
    main()
