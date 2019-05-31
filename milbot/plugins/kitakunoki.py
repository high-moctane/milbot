import datetime
from html.parser import HTMLParser
import os
import re
import random
import redis
import requests
import slack
import urllib

import utils


# ======================================================================
# 初期化
# ここに直書きするのは大変気持ち悪いが，仕方ないということにする(｀･ω･´)
# ======================================================================
class Parser(HTMLParser):
    """このクラスで HTML をパースする。

    引数:
        index_url -- 索引ページの URL
    """

    def __init__(self, index_url):
        HTMLParser.__init__(self)
        self.index_url = index_url

        # 木のリンクの class 属性は "b0_na1" or "b0_na2" なので，
        # これを使って木のリンクかどうか判別する
        self.class_attr = None

        # リンクを入れとく
        self.href = None

        # (木の名前, 木のリンク) のリスト
        self.data = []

    def handle_starttag(self, tag, attrs):
        """starttag に出会うと走る"""

        attrs = dict(attrs)
        if tag == "td":
            self.class_attr = attrs.get("class")
            self.href = None
        elif tag == "a":
            self.href = attrs.get("href")

    def handle_endtag(self, tag):
        """endtag に出会うと走る"""

        self.class_attr = None
        self.href = None

    def handle_data(self, data):
        """タグに囲まれた内容に出会うと走る"""

        if self.href is None:
            return
        if self.class_attr != "b0_na1" and self.class_attr != "b0_na2":
            return

        tree_name = data
        tree_url = urllib.parse.urljoin(self.index_url, self.href)
        self.data.append((tree_name, tree_url))


redis_cli = redis.Redis(
    host="localhost",
    port=int(os.getenv("REDIS_PORT")),
    db=os.getenv("REDIS_DB"),
    decode_responses=True
)

# redis に {name: url} の dict 形式で保存をする
tree_dict = redis_cli.hgetall("kitakunoki")

# redis データがない場合取りに行く
if tree_dict is None:
    index_url = "http://www.chiba-museum.jp/jyumoku2014/kensaku/namae.html"
    resp = requests.get(index_url)
    resp.encoding = resp.apparent_encoding
    parser = Parser(index_url)
    parser.feed(resp.text)
    parser.close()
    for name, url in parser.data:
        print(name, url)
        redis_cli.hset("kitakunoki", name, url)


# ======================================================================
# 初期化ここまで
# ======================================================================


@slack.RTMClient.run_on(event="message")
def kitakunoki(**payload):
    """ping に対して pong と返事する"""

    data = payload["data"]
    if utils.is_bot_message(data):
        return
    web_client = payload["web_client"]
    channel_id = data.get("channel")
    text = data.get("text")
    ts = data.get("ts")

    if re.match(r"^milbot kitakunoki help", text, re.IGNORECASE):
        web_client.chat_postMessage(
            channel=channel_id,
            text=help_message()
        )
    elif re.match(r"^(帰宅|きたく)の(木|き)$", text, re.IGNORECASE):
        # 日替わり帰宅の木をする
        web_client.chat_postMessage(
            channel=channel_id,
            text=todays_kitakunoki()
        )
    elif re.match(r"(帰宅|きたく)の(木|き)の(苗|なえ)", text, re.IGNORECASE):
        # 帰宅の木の苗の絵文字をつける
        web_client.reactions_add(
            channel=channel_id,
            name="seeding",
            timestamp=ts
        )
    elif re.match(r"(帰宅|きたく)の(木|き)", text, re.IGNORECASE):
        # ランダムな木のリアクションをする
        web_client.reactions_add(
            channel=channel_id,
            name=random.choice(tree_emojis()),
            timestamp=ts
        )


def help_message():
    return """`帰宅の木` に反応して今日の帰宅の木をお知らせします。
また `帰宅の木` が含まれるメッセージに絵文字をつけます。"""


def todays_kitakunoki():
    """今日の帰宅の木のメッセージを構築します"""

    redis_cli = redis.Redis(
        host="localhost",
        port=int(os.getenv("REDIS_PORT")),
        db=os.getenv("REDIS_DB"),
        decode_responses=True
    )

    tree_dict = redis_cli.hgetall("kitakunoki")
    # ここでソートすることにより日替わり性が保たれる
    name = new_rand().choice(sorted(list(tree_dict)))
    url = tree_dict[name]
    return f"今日の帰宅の木は {name} です(｀･ω･´)\n{url}"


def new_rand():
    """乱数生成器は日替わりになるようにする"""

    now = datetime.datetime.now()
    seed = int("".join(map(str, [now.year, now.month, now.day])))
    return random.Random(seed)


def tree_emojis():
    return ["palm_tree", "evergreen_tree", "deciduous_tree", "christmas_tree"]