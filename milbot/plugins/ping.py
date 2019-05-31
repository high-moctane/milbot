import re
import slack

import utils


@slack.RTMClient.run_on(event="message")
def pong(**payload):
    """ping に対して pong と返事する"""

    data = payload["data"]
    if utils.is_bot_message(data):
        return
    web_client = payload["web_client"]
    channel_id = data.get("channel")
    text = data.get("text")

    if re.match(r"^milbot ping", text, re.IGNORECASE):
        web_client.chat_postMessage(
            channel=channel_id,
            text="pong (｀･ω･´)"
        )
