def is_bot_message(data):
    """送られてきた message が bot によるものかどうかを判定する。"""

    return "bot_message" in data.values()
