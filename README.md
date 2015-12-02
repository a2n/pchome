# pchome
指令操作 ![PChome 買網址](http://myname.pchome.com.tw/img/pchomelogo_domain.gif)


## 用途
使用指令管理 PChome DNS 記錄。


## 功能
*  自管 DNS
*  DNSSEC 設定
*  域名 regex 比對，例如 ```.*tw``` 搜出所有 ```.tw``` 結尾域名。


## 指令
*  [組態](#config)
*  [NS](#ns)
*  [DNSSEC](#dnssec)

## config
```config``` 這個指令集裡面是操作組態相關動作。

### init
首次使用需要紀錄帳密。

    ./pchome config -init

詢問你的帳密後，開始從 PChome 取出所有域名、NS 和 DNSSEC 記錄。最後把結果存在當下的 ```.pchome``` JSON 檔案裡。

### remove
要移除組態，輸入

    ./pchome config -remove
    
或是也可以自行輸入下列指令移除組態檔

    rm -f .pchome

### update
此指令是和 PChome 網站同步資料

    ./pchome config -update

## ns
```ns``` 是給自管 DNS 用戶使用，用來操作 NS 記錄。

### add
添加 NS 記錄。

    ./pchome ns -add -zone example.com -name ns0.example.com -ip 10.0.0.0
    
為 ```example.com``` 這個域名添加一筆 NS 記錄，名稱為 ```ns0.example.com```，IP 為 ```10.0.0.0```。

### delete
移除 NS 記錄。

    ./pchome ns -delete -zone example.com -name ns0.example.com -ip 10.0.0.0
    
為 ```example.com``` 這個域名移除一筆 NS 記錄，名稱為 ```ns0.example.com```，IP 為 ```10.0.0.0```。


### list
列舉所有 NS 記錄。

    ./pchome ns -list -zone example.com
    
列舉出 ```example.com``` 這個域名所有的 NS 記錄。


## dnssec
### add
添加 DNSSEC 記錄。

    ./pchome dnssec -add -zone example.com -keyTag 1234 -algorithm 13 -digest 4355a46b19d348dc2f57c046f8ef63d4538ebb936000f3c9ee954a27460dd865
    
為 ```example.com``` 這個域名添加一筆 DNSSEC 記錄，key tag 為 ```1234```，algorithm 為 ```113```，digest 為 ```4355a46b19d348dc2f57c046f8ef63d4538ebb936000f3c9ee954a27460dd865```。

### delete
移除 DNSSEC 記錄。

    ./pchome dnssec -delete -zone example.com -keyTag 1234 -algorithm 13 -digest 4355a46b19d348dc2f57c046f8ef63d4538ebb936000f3c9ee954a27460dd865
    
為 ```example.com``` 這個域名移除一筆 DNSSEC 記錄，key tag 為 ```1234```，algorithm 為 ```113```，digest 為 ```4355a46b19d348dc2f57c046f8ef63d4538ebb936000f3c9ee954a27460dd865```。

### list
列舉 DNSSEC 記錄。

    ./pchome dnssec -list -zone example.com
    
列舉出 ```example.com``` 這個域名所有的 DNSSEC 記錄。

# 連結
-   [Google Groups](https://groups.google.com/forum/?fromgroups=#!forum/pchome-dns)

# 授權
[GNU General Public License v3.0](https://opensource.org/licenses/GPL-3.0)