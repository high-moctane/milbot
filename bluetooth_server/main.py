import os
from dotenv import load_dotenv
import server

if __name__ == "__main__":
    # .env の読み込み
    load_dotenv()

    # サーバをたてる
    server.Server(server_name=os.getenv("SERVER_NAME"), port=os.getenv("PORT"))
