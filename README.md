# Eth API Service

This is a homework for Blocto interview. 

## Usage

1. Rename `config.sample.json` as `config.json`.
2. Edit `config.json` and fill correct value. (endpoint need support websocket)
3. change workdir into this repo.
3. exec `go run .`

## Sync Workflow

1. 先把 config.Start 到目前鏈上最高的高度，中間的都同步完
2. 同步舊資料的同時，也使用 `eth_subscribe` 來訂閱所有新長出的區塊。
    - 不知為何 Eth 文件沒有提到 `eth_subscribe` 這個方法，我是在 ethclient source code 裡面翻到的。

其他的做法：

如果 endpoint 不支援 websocket 的話

1. 把 config.Start 到目前鏈上最高的高度，中間的都同步完
2. 使用 `eth_newBlockFilter` 取得 filter id
3. 每隔 N 秒，使用 `eth_getFilterChanges` 搭配 filter id 取得「新出現的 Block ID」
4. 使用 `eth_getBlockByNumber` 取得區塊與交易的詳細資料
5. 存回資料庫
