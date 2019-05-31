import html
import os
import re
import slack
import subprocess
import tempfile

import utils


@slack.RTMClient.run_on(event="message")
def program_runner(**payload):
    """任意コード実行（脆弱性）"""

    data = payload["data"]
    # 謎のコメントアウト
    # if utils.is_bot_message(data):
    #     return
    web_client = payload["web_client"]
    channel_id = data.get("channel")
    text = data.get("text")

    if re.match(r"^milbot bash help", text, re.IGNORECASE):
        mes = bash_help()
    elif re.match(r"^milbot bash", text, re.IGNORECASE):
        mes = message(*bash(extract_code(text)))
    elif re.match(r"^milbot python3 help", text, re.IGNORECASE):
        mes = python3_help()
    elif re.match(r"^milbot python3", text, re.IGNORECASE):
        mes = message(*python3(extract_code(text)))
    else:
        return

    web_client.chat_postMessage(
        channel=channel_id,
        text=mes
    )


def extract_code(text):
    """text からコードの部分を取り出す"""

    elems = text.split()
    command = elems[0] + " " + elems[1]
    code = text[len(command)+1:]
    if code[:3] == "```":
        code = code[3:]
    if code[-3:] == "```":
        code = code[:-4]
    return html.unescape(code)


def message(stdout, stderr, return_code):
    """結果からメッセージを構築する"""

    message = ""
    if stdout != "":
        message += "stdout:\n```\n" + stdout + "```\n"
    if stderr != "":
        message += "stderr:\n```\n" + stderr + "```\n"
    message += f"return code: {return_code}"
    return message


def run(code, pre_filename, post_filename):
    """任意のコードを実行する。

    引数:
        pre_filename -- subprocess.run に渡すコマンドのファイル名の前までの部分
        post_filename -- subprocess.run に渡すコマンドのファイル名の後の部分
    """

    with tempfile.TemporaryDirectory() as d:
        with open(os.path.join(d, "main.sh"), "w") as f:
            f.write(code)
            f.flush()
            command = pre_filename + [f.name] + post_filename
            proc = subprocess.run(
                command,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE
            )
    try:
        stdout = proc.stdout.decode()
        stderr = proc.stderr.decode()
    except:
        return "", "バイナリを吐かないでください(´･ω･｀)", proc.returncode
    return stdout, stderr, proc.returncode


def bash(code):
    return run(code, ["bash"], [])


def bash_help():
    return """任意の bash スクリプトを実行する脆弱性です。
`milbot bash` に続けてスクリプト本文を書いてください。
スクリプト本文を Markdown のコードのように ``` で囲って書くのがおすすめです。
"""


def python3(code):
    return run(code, ["python3", "-B"], [])


def python3_help():
    return """任意の python3 スクリプトを実行する脆弱性です。
`milbot python3` に続けてスクリプト本文を書いてください。
スクリプト本文を Markdown のコードのように ``` で囲って書くのがおすすめです。
"""
