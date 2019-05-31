import slack


@slack.RTMClient.run_on(event="hello")
def hello(**payload):
    """つながったことを表示する"""
    print("Connected!")
