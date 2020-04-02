# Milbot

某研究室で動く Slack の bot です(｀･ω･´)

メインの機能は研究室にいるメンバーの在室確認です。

在室確認にはスマートフォンの Bluetooth を利用します。
必ず Bluetooth をオンにしておいてください。

## 使い方

まずは `milbot help` を実行しましょう。コマンドの使い方が表示されます。

多くの機能は `milbot` で始まります。

## Bot の起動

Raspberry Pi を電源に差し込むと自動で起動します。

## Bot の終了

終了方法はいくつかあります。

- `milbot exit` コマンドを送ると bot が終了します。

- Raspberry Pi 上で以下のコマンドを実行します。
    ```bash
    # systemctl stop milbot
    ```

## Bot の機能追加

この Bot はプラグイン形式で機能を追加する構成にしてあります。

`botplugin.Plugin` interface を実装すると，プラグインになることができます。
詳しい定義は [botplugin/botplugin.go](botplugin/botplugin.go) を見てみてください。
作り方は [botplugins/ping/ping.go](botplugins/ping/ping.go) を参考にすると良いです。

プラグインは `botplugins` 以下に配置してください。

プラグインが完成したら，[main.go](main.go) の `plugins` にプラグインのインスタンスを
登録してください。
Bot の起動時に自動的に読み込まれて有効になります。

おもしろい機能を作ってプルリクを送ってください。

## `libatnd` について

`libatnd` は Bluetooth を使って在室管理を行うときに使えるライブラリです。
在室管理機能を使っておもしろいプラグインを作ってください。


## Milbot のセットアップ

### Raspberry Pi の準備

Milbot は Raspberry Pi 上で動かすことを前提としています。

ディストリビューションは Raspbian Lite を使うと楽で良いと思います。

`tmux` を Raspbian にインストールしておくと楽です。

### ネットワークを起動させてから bot を起動するようにするための操作

以下のコマンドを実行します。

```
# systemctl enable systemd-networkd
# systemctl enable systemd-networkd-wait-online
```

### 実行ファイルのビルド

以下のコマンドを使って Raspberry Pi 用の実行ファイルをビルドします。

```bash
$ make raspi
```

### 実行ファイルのデプロイ

生成された `milbot-raspi` を `scp` などを使って Raspberry Pi の
`/home/pi/milbot` 以下に配置します。

### Milbot の自動起動の有効化

以下のコマンドを実行して Raspberry Pi が起動したときに bot も起動するように
設定します。

```bash
# wget -P /etc/systemd/system/ https://raw.githubusercontent.com/high-moctane/milbot/master/milbot.service
# systemctl enable milbot
```
