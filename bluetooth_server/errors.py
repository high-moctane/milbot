class BluetoothServerError(Exception):
    """Bluetooth server のエラーの基底クラス"""


class InvalidBDAddrError(BluetoothServerError):
    """Bluetooth アドレスが不正なときのエラー。

    引数:
        bd_addr -- エラーとなった bd_addr
    """

    def __init__(self, bd_addr):
        self.bd_addr = bd_addr
