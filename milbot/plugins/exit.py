import re
import slack
import sys

from log import Log
import utils


@slack.RTMClient.run_on(event="message")
async def bot_exit(**payload):
    """bot を終了させる"""

    data = payload["data"]
    if utils.is_bot_message(data):
        return
    web_client = payload["web_client"]
    channel_id = data.get("channel")
    text = data.get("text")

    if re.match(r"^milbot exit help", text, re.IGNORECASE):
        try:
            await web_client.chat_postMessage(
                channel=channel_id,
                text="milbot を終了するコマンドです(｀･ω･´)"
            )
        except Exception as e:
            Log.error(e)
            raise(e)

    elif re.match(r"^milbot exit", text, re.IGNORECASE):
        try:
            web_client.chat_postMessage(
                channel=channel_id,
                text="milbot を終了します(｀-ω-´)zzZ"
            )
        finally:
            sys.exit()
