import asyncio
import os
import re
import redis
import requests
import slack

import utils


@slack.RTMClient.run_on(event="message")
async def atnd(**payload):
    """出席管理"""

    data = payload["data"]
    if utils.is_bot_message(data):
        return

    web_client = payload["web_client"]
    channel_id = data.get("channel")
    text = data.get("text")

    redis_cli = redis.Redis(
        host="localhost",
        port=int(os.getenv("REDIS_PORT")),
        db=os.getenv("REDIS_DB"),
        decode_responses=True
    )

    if re.match(r"^milbot atnd add", text, re.IGNORECASE):
        # メンバー登録をする
        mes = add(text, redis_cli)
    elif re.match(r"^milbot atnd delete", text, re.IGNORECASE):
        # メンバー削除をする
        mes = delete(text, redis_cli)
    elif re.match(r"^milbot atnd list", text, re.IGNORECASE):
        # メンバーのリストを出す
        mes = show_list(redis_cli)
    elif re.match(r"^milbot atnd help", text, re.IGNORECASE):
        # ヘルプメッセージ
        mes = help_message()
    elif re.match(r"^milbot atnd", text, re.IGNORECASE):
        # 在室確認をする
        mes = atnd_default(redis_cli)
    else:
        return

    await web_client.chat_postMessage(
        channel=channel_id,
        text=mes
    )


def add(text, redis_cli):
    """メンバー登録をする"""

    elems = text.split()
    if len(elems) != 5:
        return "不正な入力です\nusage: milbot atnd add name address"

    name, bd_addr = elems[3], elems[4]
    if not is_valid_bd_addr(bd_addr):
        return f"不正な Bluetooth アドレスです: {bd_addr}"

    members = redis_cli.hkeys("atnd_members")
    if name in members:
        return f"{name} は既に登録されています"

    redis_cli.hset("atnd_members", name, bd_addr)

    return f"登録しました {name}: {bd_addr}"


def delete(text, redis_cli):
    """メンバー削除をする"""

    elems = text.split()
    if len(elems) != 4:
        return "不正な入力です\nusage: milbot atnd delete name"

    name = elems[3]
    members = redis_cli.hkeys("atnd_members")
    if name not in members:
        return f"{name} はもともと登録されていません"

    redis_cli.hdel("atnd_members", name)
    return f"{name} を削除しました"


def show_list(redis_cli):
    """メンバーリストを出す"""

    members = redis_cli.hkeys("atnd_members")
    if len(members) == 0:
        return "誰も登録されていません(´･ω･｀)"
    return "以下のメンバーが登録されています(｀･ω･´)\n" + "\n".join(members)


def atnd_default(redis_cli):
    """メンバーがいるかどうかを検索する"""

    member_addr = redis_cli.hgetall("atnd_members")
    addr_member = {addr: member for addr, member in member_addr.items()}
    addrs_str = "\n".join(list(addr_member)) + "\n"
    try:
        resp = requests.post(
            "http://" + os.getenv("ATND_SERVER_NAME"), addrs_str)
    except Exception as e:
        return f"エラーが発生しました(´･ω･｀)\n{e}"
    if resp.status_code != "200":
        return f"エラーが発生しました(´･ω･｀)\n{resp.text}"
    exist_addrs = resp.text.split()

    return atnd_default_message(addr_member, exist_addrs)


def atnd_default_message(addr_member, addrs):
    """メンバーがいるかどうかのメッセージ構築をする

    引数:
        addr_member -- {addr: member_name} の dict
        addrs -- 在室している bd_addr
    """

    if len(addrs) == 0:
        return "現在部屋には誰もいません(´･ω･｀)"

    mes = "現在部屋には\n"
    for addr in addrs:
        mes += "* " + addr_member[addr] + "\n"
    mes += "が在室しています(｀･ω･´)"
    return mes


def is_valid_bd_addr(bd_addr):
    """有効な形式の bd_addr かどうか検査する"""

    if len(bd_addr) != 17:
        return False

    for i in range(2, 15, 3):
        if bd_addr[i] != ":":
            return False

    for num in bd_addr.split(":"):
        try:
            int(num, 16)
        except:
            return False

    return True


def help_message():
    return """以下の機能があります。
`milbot atnd help`
    このメッセージを表示する。

`milbot atnd`
    出席しているメンバーを確認する。

`milbot atnd list`
    登録してあるメンバーを確認する。

`milbot atnd add name bd_addr`
    メンバー登録をする。
    `name` に自分の名前を入れてください。半角スペースは使えません。
    `bd_addr` に Bluetooth アドレスを `00:33:66:99:aa:dd` 形式で書いてください。

`milbot atnd delete name`
    メンバーを削除します。
    `name` に削除するメンバーの名前を正しく書いてください。
    `milbot atnd list` で正確な名前を確認してください。
"""
