def is_bot_message(data):
    """送られてきた message が bot によるものかどうかを判定する。"""

    if "bot_message" in data.values():
        return True
    else:
        return False
