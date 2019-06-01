import datetime
import os
import requests
import sys


class Log:
    """自前のロガーを作ってしまった……
    milbot は docker で動くことを想定しているので，ロギングは出力で OK
    """
    _url = os.getenv("SLACK_LOG_WEBHOOK_URL")

    @classmethod
    def info(self, message):
        """うまく行った処理を吐き出す。表示のみする。

        引数:
            message -- 内容
        """
        mes = self._format("INFO", message)
        self._print(mes)

    @classmethod
    def error(self, message):
        """うまく行かなかった処理を吐き出す。表示と Slack 投稿する。

        引数:
            message -- 内容
        """
        mes = self._format("ERROR", message)
        self._print(mes)
        self._send_slack_message(mes)

    @classmethod
    def fatal(self, message):
        """プログラム続行不可能なときに使う。表示と Slack 投稿し，bot を終了する。

        引数:
            message -- 内容
        """
        mes = self._format("FATAL", message)
        self._print(mes)
        self._send_slack_message(mes)

    @classmethod
    def _format(self, level, message):
        return f"[{self._timestamp()}] {level}: {message}"

    @classmethod
    def _timestamp(self):
        now = datetime.datetime.now()
        return now.strftime("%Y/%m/%d %H:%M:%S.%f")

    @classmethod
    def _print(self, message):
        print(message, file=sys.stderr)

    @classmethod
    def _send_slack_message(self, message):
        try:
            requests.post(self._url,
                          json={"text": message})
        except Exception as e:
            self._print(self._format("ERROR", str(e)))
