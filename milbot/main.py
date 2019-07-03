# dotenv 読み込み
import settings

import asyncio
import certifi
import os
import signal
import slack
import ssl as ssl_lib
import sys
import time
import threading

from log import Log

# Plugins
import import_plugins


def kill():
    sec = int(os.getenv("RESTART_HOUR")) * 3600
    time.sleep(sec)
    os.kill(os.getpid(), signal.SIGKILL)


if __name__ == "__main__":
    token = os.getenv("SLACK_API_TOKEN")
    loop = asyncio.get_event_loop()
    asyncio.set_event_loop(loop)

    # 自動再起動の設定
    thread = threading.Thread(target=kill)
    thread.start()

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
