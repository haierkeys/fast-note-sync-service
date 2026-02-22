import json
import sys
import os
import traceback
from deep_translator import GoogleTranslator

def main():
    if len(sys.argv) < 3:
        print(json.dumps([]))
        return

    arg_input = sys.argv[1]
    target = sys.argv[2]

    texts = []
    try:
        # 优先尝试作为文件路径读取
        if os.path.isfile(arg_input):
            with open(arg_input, 'r', encoding='utf-8') as f:
                texts = json.load(f)
        else:
            # 否则尝试作为 JSON 字符串解析
            texts = json.loads(arg_input)
    except Exception as e:
        print(f"Input Error: {e}", file=sys.stderr)
        print(json.dumps([]))
        return

    if not texts:
        print(json.dumps([]))
        return

    try:
        if target == 'zh-TW':
            target = 'zh-TW'

        translator = GoogleTranslator(source='auto', target=target)

        # 预处理：提取需要翻译的非空项
        to_translate = []
        mapping_indices = []
        for i, text in enumerate(texts):
            clean_text = text.strip() if text else ""
            if clean_text and clean_text != '-':
                to_translate.append(text)
                mapping_indices.append(i)

        results = list(texts) # 默认使用原文

        if to_translate:
            try:
                # 使用 translate_batch (deep-translator 1.9.0+ 支持)
                translated_list = translator.translate_batch(to_translate)
                for idx, translated in zip(mapping_indices, translated_list):
                    results[idx] = translated if translated else texts[idx]
            except Exception as e:
                # 如果批量翻译失败，回退到逐条翻译（防止全盘皆墨）
                print(f"Batch failed ({target}): {e}, switching to sequential.", file=sys.stderr)
                for idx, text in zip(mapping_indices, to_translate):
                    try:
                        res = translator.translate(text)
                        results[idx] = res if res else text
                    except:
                        pass

        # 使用 ensure_ascii=True (默认) 确保输出纯 ASCII 字符，彻底解决 Windows 编码干扰
        print(json.dumps(results))

    except Exception as e:
        print(f"Execution Error: {e}", file=sys.stderr)
        print(json.dumps(texts))

if __name__ == "__main__":
    main()
