import slack

from log import Log


@slack.RTMClient.run_on(event="hello")
def hello(**payload):
    """つながったことを表示する"""
    Log.info("RTM connected")
