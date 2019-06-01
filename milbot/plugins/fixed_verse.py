import MeCab
import re
import slack

import utils


@slack.RTMClient.run_on(event="message")
async def fixed_verse(**payload):
    """575 とか検出する"""

    data = payload["data"]
    if utils.is_bot_message(data):
        return
    web_client = payload["web_client"]
    channel_id = data.get("channel")
    text = data.get("text")
    text = "".join(text.split())

    m = message(text)
    if m != "":
        await web_client.chat_postMessage(
            channel=channel_id,
            text=m
        )


jiritsugo = ["動詞", "形容詞", "形容動詞", "名詞", "連体詞", "副詞", "接続詞", "感動詞"]
omit_part_of_speech = ["記号"]


class Morpheme:
    """形態素1つ1つを表す。

    引数:
        node -- MeCab でパースしたノード
    """

    def __init__(self, node):
        self.surface = node.surface
        elems = node.feature.split(",")
        self.part_of_speech = elems[0]
        self.pronounce = elems[-1]

        # 拗音は 1 モーラであることに注意
        self.mora_len = len(re.sub(r"[\*ャィュェョ]", "", elems[-1]))


def parse_text(text):
    """文字列を Morpheme の列に変換。

    引数:
        text -- 文字列

    戻り値:
        Morpheme のリスト
    """

    ans = []
    t = MeCab.Tagger()
    node = t.parseToNode(text)

    while node:
        if node.surface == "":
            node = node.next
            continue
        ans.append(Morpheme(node))
        node = node.next

    return ans


def find_a_valid_phrase(morphemes, mora_len):
    """morpheme 列から mola_len を満たすフレーズを抽出する。"""

    global omit_part_of_speech

    if len(morphemes) == 0:
        return None

    ans = []
    total_mora_len = 0
    i = 0
    while i < len(morphemes) and total_mora_len < mora_len:
        if morphemes[i].mora_len == 0:
            return None
        if morphemes[i].part_of_speech in omit_part_of_speech:
            ans.append(morphemes[i])
            i += 1
            continue
        if re.match(r"[^ァ-ンー]", morphemes[i].pronounce):
            return None

        ans.append(morphemes[i])
        total_mora_len += morphemes[i].mora_len
        i += 1

    if total_mora_len != mora_len:
        return None
    return ans


def find_a_verse(morphemes, mora_len_list):
    """有効な詩を見つける"""

    if len(mora_len_list) == 0:
        return None

    morphs = morphemes[:]

    verse = []
    for mola_len in mora_len_list:
        if len(morphs) == 0:
            return None
        phrase = find_a_valid_phrase(morphs, mola_len)
        if phrase is None:
            return None
        verse.append(phrase)
        morphs = morphs[len(phrase):]

    return verse


def find_all_verses(morphemes, mora_len_list):
    """morpheme 列から該当するすべての詩を見つける"""

    if len(morphemes) == 0 or len(mora_len_list) == 0:
        return []

    verses = []
    for i in range(len(morphemes)):
        verse = find_a_verse(morphemes[i:], mora_len_list)
        if verse is None:
            continue
        verses.append(verse)

    return verses


def morphemes_to_str(morphemes):
    return "".join([m.surface for m in morphemes])


def verse_to_str(verse):
    separator = " ／ "
    ans = ""
    for phrase in verse:
        ans += morphemes_to_str(phrase) + separator
    return ans[:-len(separator)]


def find(morphemes, mora_len_list, jiritsugo_flags):
    """morphemes から mola_len_list に該当し，jiritsugo_flags を満たす詩を見つける"""
    global jiritsugo

    valid_verses = []
    for verse in find_all_verses(morphemes, mora_len_list):
        valid = True
        for i, f in enumerate(jiritsugo_flags):
            if f and verse[i][0].part_of_speech not in jiritsugo:
                valid = False
        if not valid:
            continue
        valid_verses.append(verse)

    return valid_verses


def senryu(morphemes):
    """575 を探してくる"""

    verses = find(morphemes, [5, 7, 5], [True, False, False])
    return list(map(verse_to_str, verses))


def tanka(morphemes):
    """57577 を探してくる"""

    verses = find(morphemes, [5, 7, 5, 7, 7], [
                  True, False, False, True, False])
    return list(map(verse_to_str, verses))


def dodoitsu(morphemes):
    """7775 を探してくる"""

    verses = find(morphemes, [7, 7, 7, 5], [True, False, False, False])
    return list(map(verse_to_str, verses))


def message(text):
    mes = []
    morphemes = parse_text(text)
    s = senryu(morphemes)
    t = tanka(morphemes)
    d = dodoitsu(morphemes)
    if len(s) != 0:
        mes.append("Found 575:cop:\n" + "\n".join(s))
    if len(t) != 0:
        mes.append("Found 57577:cop:\n" + "\n".join(t))
    if len(d) != 0:
        mes.append("Found 7775:cop:\n" + "\n".join(d))
    return "\n\n".join(mes)
