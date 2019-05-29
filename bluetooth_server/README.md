# Bluetooth Server

このサーバに Bluetooth アドレスのリストの POST リクエストを送ると，
サーバが Bluetooth を走らせて，送られてきた Bluetooth アドレスのうち
近くに存在するもののみを返します。

## 起動方法

環境変数に `SERVER_NAME` と `PORT` を設定して起動しましょう。

`l2ping` を使っているので管理者権限が必要です。

```
$ sudo SERVER_NAME=localhost PORT=8080 python3 main.py
```

## 例

POST リクエストの例

```
POST / HTTP/1.1
HOST localhost:8080
Contents-Length: xx
Content-Type: text/plain

AA:AA:AA:AA:AA:AA
BB:BB:BB:BB:BB:BB
CC:CC:CC:CC:CC:CC
DD:DD:DD:DD:DD:DD
```

帰ってくる返事の例

```
HTTP/1.1 200 OK
Content-Type: text/plain

BB:BB:BB:BB:BB:BB
DD:DD:DD:DD:DD:DD
```

ステータスコードが 200 以外のときはなんかしらのエラーが起きています。
