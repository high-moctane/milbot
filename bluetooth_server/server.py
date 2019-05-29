from http.server import BaseHTTPRequestHandler, HTTPServer
import os
import sys
import traceback

import errors as err
import search


class Handler(BaseHTTPRequestHandler):
    """Body に改行（空白）区切りの POST リクエストを受け付け，
    そのうち存在する bd_addr を返すという http サーバのハンドラ。"""

    def do_POST(self):
        try:
            # bd_addrs の取得
            content_len = int(self.headers.get("content-length"))
            request_body = self.rfile.read(content_len).decode()
            bd_addrs = request_body.split()

            # ヘッダを作る
            self.send_response(200)
            self.send_header("Content-type", "text/plain")
            self.end_headers()

            # サーチする
            s = search.Search(bd_addrs, max_workers=int(
                os.getenv("MAX_WORKERS")))
            exist_bd_addrs = s.run()

            # 返事する
            response_body = "\n".join(exist_bd_addrs)
            self.wfile.write(response_body.encode())

        except Exception:
            t, v, tb = sys.exc_info()
            trace = traceback.format_exception(t, v, tb)
            print("\n".join(trace))

            # ヘッダを作る
            self.send_response(400)
            self.send_header("Content-type", "text/plain")
            self.end_headers()

            # 返事する
            response_body = "Error\n" + "\n".join(trace)
            self.wfile.write(response_body.encode())


class Server:
    """サーバ。"""

    def __init__(self, server_name="localhost", port=8080):
        self.server_name = server_name
        self.port = port

    def run(self):
        server = HTTPServer((self.server_name, self.port), Handler)
        server.serve_forever()
