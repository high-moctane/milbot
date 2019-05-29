from concurrent.futures import ThreadPoolExecutor

import l2ping


class Search:
    """与えられた bd_addr のリストそれぞれについて存在するかサーチする。

    引数:
        bd_addr_list -- bd_addr のリスト
    """

    def __init__(self, bd_addr_list, max_workers=10):
        self.bd_addr_list = bd_addr_list
        self.max_workers = max_workers

    def run(self):
        """サーチを実行する。

        戻り値:
            存在する bd_addr のリスト
        """

        l2pings = [l2ping.L2ping(bd_addr) for bd_addr in self.bd_addr_list]

        with ThreadPoolExecutor(max_workers=self.max_workers) as executor:
            results = executor.map(Search._run_l2ping, l2pings)

        return filter(None, results)

    @classmethod
    def _run_l2ping(self, l2ping):
        return l2ping.run()
