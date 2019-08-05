# milbot

某研究室で動く Slack の bot です(｀･ω･´)

## 構成

```
├── bin
├── bluetooth_server
├── milbot
│   ├── Dockerfile
│   ├── botutils
│   ├── main.go
│   ├── plugin
│   ├── plugins
├── redis_docker
└── systemd
```

`bluetooth_server` は Raspberry Pi 上で動かすメンバー探索用スクリプトです。
管理者権限で実行する必要があります。

`milbot` は Slack bot 本体です。Docker 上で動きます。

`redis_docker` はデータ永続化のための Docker コンテナです。

`bin` は便利スクリプト集です。

`systemd` は bot の自動起動用の service ファイルです。


## 運用

### 起動と停止

`bin` の中のスクリプトを管理者権限で実行してください。

`systemd` の中の service ファイルを適当な場所において `systemctl` でうまいこと
自動起動できます。


### 更新

`milbot` 本体を更新する際は，コンパイルは Raspberry Pi 上でやらないほうがいいです。
とても時間がかかります。
手元の環境でクロスコンパイルし，`scp` などをつかって転送しましょう。
その後 `docker build -t milbot /path/to/milbot` を実行しましょう。

クロスコンパイルは以下のようにすればできます。バイナリ名は `milbot-rpi` にしてください。

```
$ GOOS=linux GOARCH=arm GOARM=6 go build -o milbot-rpi
```


## メンバー探索の仕組み

`bluetooth_server` は HTTP サーバになっています。`bluetooth_server` に
改行区切りの bd_addr を body にした POST リクエストを送ると，
部屋に存在する bd_addr を改行区切りで返します。


## 機能の追加

`milbot/plugin/plugin.go` にある `Plugin interface` を満たしたものはなんでも
プラグインになることができます。

`milbot/main.go` のグローバル変数 `plugins` スライスにプラグインのインスタンスを
格納してください。
自動的に実行されます。

プラグインは `milbot/plugins/` 以下にまとめてあります。
作り方は `milbot/plugins` 以下のプラグインを参考にしてみてください。

`milbot/botutils` に bot に便利な関数をまとめておきました。使ってみてください。

動作確認を行う際はユニットテストをするのがいいと思います。

プルリクどんどんください(｀･ω･´)