import re
import slack

from log import Log
import utils


@slack.RTMClient.run_on(event="message")
async def help_func(**payload):
    """コマンド一覧をつくる"""

    data = payload["data"]
    if utils.is_bot_message(data):
        return
    web_client = payload["web_client"]
    channel_id = data.get("channel")
    text = data.get("text")

    try:
        if re.match(r"^milbot help hidden", text, re.IGNORECASE):
            await web_client.chat_postMessage(
                channel=channel_id,
                text=mes_hidden()
            )
        elif re.match(r"^milbot help", text, re.IGNORECASE):
            await web_client.chat_postMessage(
                channel=channel_id,
                text=mes()
            )
    except Exception as e:
        Log.error(e)
        raise(e)


def mes():
    return """以下のコマンドを受け付けます(｀･ω･´)
`milbot help`
`milbot help hidden`
`milbot ping`
`milbot peng`
`milbot peng help`
`milbot atnd`
`milbot atnd help`
`milbot atnd add`
`milbot atnd delete`
`milbot atnd list`
`帰宅の木`
`milbot kitakunoki help`

また以下の語句に反応します(｀･ω･´)
`帰宅の木`
`帰宅の木の苗`
575
57577
7775
"""


def mes_hidden():
    return """隠しコマンドです。絶対に実行しないでください(´･ω･｀)
`milbot exit`
`milbot exit help`
`milbot bash`
`milbot bash help`
`milbot python3`
`milbot python3 help`
"""
