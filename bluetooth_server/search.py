import time

import l2ping


class Search:
    """与えられた bd_addr のリストそれぞれについて存在するかサーチする。

    引数:
        bd_addr_list -- bd_addr のリスト
    """

    def __init__(self, bd_addr_list):
        self.bd_addr_list = bd_addr_list

    def run(self):
        """サーチを実行する。

        戻り値:
            存在する bd_addr のリスト
        """

        exists = []
        for bd_addr in self.bd_addr_list:
            if l2ping.L2ping(bd_addr).run() is not None:
                exists.append(bd_addr)
            # 連続して呼び出すともしかして電波拾わないかもしれない
            time.sleep(0.2)

        return exists
