import os
import random
import re
import slack

import utils


@slack.RTMClient.run_on(event="message")
async def peng(**payload):
    """ペンギン燃やし"""

    data = payload["data"]
    if utils.is_bot_message(data):
        return
    web_client = payload["web_client"]
    channel_id = data.get("channel")
    text = data.get("text")

    if re.match(r"^milbot peng help", text, re.IGNORECASE):
        mes = help_message(jackpot_probability())
    elif re.match(r"^milbot peng", text, re.IGNORECASE):
        mes = fire_penguin(jackpot_probability())
    else:
        return

    await web_client.chat_postMessage(
        channel=channel_id,
        text=mes
    )


def jackpot_probability():
    """当たる確率"""
    return float(os.getenv("PENG_PROBABILITY"))


def fire_probability(jackpot_prob):
    """jackpot_prob を実現する各要素が炎になる確率"""
    return jackpot_probability() ** (1/8)


def emoji(fire_prob):
    """fire_prob の確率で :fire: になる"""
    if random.random() < fire_prob:
        return ":fire:"
    else:
        return ":snowflake:"


def fire_penguin(jackpot_prob):
    """jackpot_prob のもとで確率的ファイアペンギンを生成する"""
    fire_prob = fire_probability(jackpot_prob)
    fire_snow = [emoji(fire_prob) for _ in range(8)]

    atari_text = ""
    if ":snowflake:" not in fire_snow:
        atari_text = ":tada:" * 3

    mes = "\n".join([
        "".join(fire_snow[:3]),
        fire_snow[3] + ":penguin:"+fire_snow[4],
        "".join(fire_snow[5:]),
        "",
        atari_text
    ])
    return mes


def help_message(jackpot_prob):
    return f"ペンギン燃やしゲームです(｀･ω･´)\n当たりの確率は {jackpot_prob} です:fire::penguin::fire:"
