# 概要  
dockerのcontainerの中で動いてる各プロセスの情報を出力するexporterです．  
CPU使用率，GPU memory使用サイズ，メモリ使用率，vsz,rssを出力します．  

GPU memory使用サイズを得るために，nvidia/dcgm-exporterのコンテナを利用してます．  
そのため，事前にdcgm-exporterのコンテナが`--pid="host"`オプション付きで起動している必要があります．

# 使い方  
1. nvidia/dcgm-exporterのコンテナを`--pid="host"`オプション付きで起動させる．
1. コンテナ作成  
`docker-compose up -d`
1. 起動確認  
`curl localhost:8088/metrics`
