# gimer

gimerはGo言語で作られたCLIタイマーツールです。
指定した時間のタイマーを作成し、終了時に通知音がなります。

![image](https://github.com/user-attachments/assets/204edf33-8f29-414f-9426-32b9302bb8ab)
![image](https://github.com/user-attachments/assets/5a70576f-b668-41c6-b065-f2a748cb5d01)

## 特徴
- 時間の指定が可能（秒、分、時間）
- タイマーの一時停止および再開機能
- タイマー終了時に通知音を再生
- タイマーの保存と再利用
- タイマーに説明をつけられる

## インストール方法

以下2つの方法があります。

### 1つ目：GitHubからCloneする方法
```sh
# gimerをgithubからclone
git clone git@github.com:kyaoi/gimer.git
cd gimer

# 依存関係をインストール
go mod tidy

# gimerをビルド
go build -o gimer
```


### 2つ目：go installでインストールする方法
```sh
go install github.com/kyaoi/gimer@latest
```



## 使い方

### タイマーを作成する場合
```sh
# 例）1時間半のタイマーを作成する場合
gimer start -H 1 -M 30
```

### タイマーを終わらせる場合

別のターミナルで`gimer stop`をするか、タイマー画面でqキーを押してください。



### タイマーを停止・再開する場合

タイマー画面でspaceキーを押してください。

### その他
その他コマンドの使い方についてはヘルプコマンドを実行して確認してください。

```sh
gimer -h
gimer start -h
gimer stop -h
gimer status -h
```
