# dotenv 読み込み
import settings

import certifi
import os
import slack
import ssl as ssl_lib

# Plugins
import import_plugins

if __name__ == "__main__":
    token = os.getenv("SLACK_API_TOKEN")
    ssl_context = ssl_lib.create_default_context(cafile=certifi.where())
    rtm_client = slack.RTMClient(token=token, ssl=ssl_context)
    rtm_client.start()
