### 接口规范



#### 一、参数

#### HTTP Post Body 参数

首字母大写，驼峰

示例:

```json
{
    "Path": "./wallet.dat",
    "Desc": "wallet.dat",
    "Duration": 0,
    "Interval": 3600,
    "Times": 24,
    "Privilege": 1,
    "CopyNum": 0,
    "EncryptPassword": "",
    "Url": "oni://share/12nsdhu",
    "WhiteList": [],
    "Share": false,
    "StoreType": 0
}
```



#### HTTP Get Query 参数

首字母小写，驼峰

示例:

```
http://{{host}}/api/v1/dsp/file/uploadfee/77616C6C65742E646174?duration=0&storeType=1
```

