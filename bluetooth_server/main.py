import argparse
import os
import server

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--address", help="an address of server", default="0.0.0.0")
    parser.add_argument("--port", help="port number", type=int, default=8080)
    args = parser.parse_args()

    # サーバをたてる
    server.Server(server_name=args.address, port=args.port).run()
