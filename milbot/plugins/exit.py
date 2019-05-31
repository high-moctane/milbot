import re
import slack
import sys

import utils


@slack.RTMClient.run_on(event="message")
def bot_exit(**payload):
    """bot を終了させる"""

    data = payload["data"]
    if utils.is_bot_message(data):
        return
    web_client = payload["web_client"]
    channel_id = data.get("channel")
    text = data.get("text")

    if re.match(r"^milbot exit help", text, re.IGNORECASE):
        web_client.chat_postMessage(
            channel=channel_id,
            text="milbot を終了するコマンドです。\n気軽に実行しないでください。"
        )
    elif re.match(r"^milbot exit", text, re.IGNORECASE):
        web_client.chat_postMessage(
            channel=channel_id,
            text="milbot を終了します(｀-ω-´)zzZ"
        )
        sys.exit()
