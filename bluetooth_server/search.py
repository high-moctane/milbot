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

        return filter(None, [l2ping.L2ping(bd_addr).run() for bd_addr in self.bd_addr_list])
