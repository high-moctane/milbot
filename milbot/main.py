# dotenv 読み込み
import settings

import asyncio
import certifi
import os
import slack
import ssl as ssl_lib

from log import Log

# Plugins
import import_plugins

if __name__ == "__main__":
    token = os.getenv("SLACK_API_TOKEN")
    loop = asyncio.get_event_loop()
    asyncio.set_event_loop(loop)
    try:
        ssl_context = ssl_lib.create_default_context(cafile=certifi.where())
        rtm_client = slack.RTMClient(
            token=token,
            ssl=ssl_context,
            run_async=True,
            loop=loop
        )
        future = rtm_client.start()
        loop.run_until_complete(future)
    except Exception as e:
        Log.fatal(str(e))
        raise(e)
