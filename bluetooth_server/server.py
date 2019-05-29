import asyncio
from http.server import BaseHTTPRequestHandler, HTTPServer

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
            s = search.Search(bd_addrs)
            loop = asyncio.get_event_loop()
            exist_bd_addrs = loop.run_until_complete(s.run())
            loop.close()

            # 返事する
            response_body = "\n".join(exist_bd_addrs)
            self.wfile.write(response_body.encode())

        except err.BluetoothServerError as e:
            # ヘッダを作る
            self.send_response(200)
            self.send_header("Content-type", "text/plain")
            self.end_headers()

            # 返事する
            response_body = "BluetoothServerError\n" + str(e)
            self.wfile.write(response_body.encode())

        except Exception as e:
            # ヘッダを作る
            self.send_response(200)
            self.send_header("Content-type", "text/plain")
            self.end_headers()

            # 返事する
            response_body = "Error\n" + str(e)
            self.wfile.write(response_body.encode())


class Server:
    """サーバ。"""

    def __init__(self, server_name="localhost", port="8080"):
        self.server_name = server_name
        self.port = port

    def run(self):
        server = HTTPServer((self.server_name, self.port), Handler)
        server.serve_forever()
