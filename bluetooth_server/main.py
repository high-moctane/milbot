import os
import server

if __name__ == "__main__":
    # サーバをたてる
    server.Server(server_name=os.getenv("SERVER_NAME"),
                  port=int(os.getenv("PORT"))).run()
