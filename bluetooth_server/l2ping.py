import asyncio

import errors as err


class L2ping:
    """l2ping を飛ばすためのクラス。

    引数:
        bd_addr -- Bluetoothアドレスの文字列
    """

    def __init__(self, bd_addr):
        if not self._is_valid_bd_addr(bd_addr):
            raise(err.InvalidBDAddrError(bd_addr))
        self.bd_addr = bd_addr

    async def run(self):
        """l2ping を実行する。

        戻り値:
            bd_addr が見つかった場合 -> bd_addr
            見つからなかった場合     -> None
        """

        proc = await asyncio.create_subprocess_shell(
            "l2ping -c 1 {}".format(self.bd_addr)
        )
        if proc.returncode != 0:
            return None
        else:
            return self.bd_addr

    def _is_valid_bd_addr(self, bd_addr):
        if len(bd_addr) != 17:
            return False

        for i in range(2, 15, 3):
            if bd_addr[i] != ":":
                return False

        for num in bd_addr.split(":"):
            try:
                int(num, 16)
            except:
                return False

        return True
