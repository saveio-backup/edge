### 电影DAPP 合约接口v0.1

------





#### 1. 获取电影
- 请求参数

  | 参数         | 类型   | 必填 | 说明                   |
  | ------------ | ------ | ---- | ---------------------- |
  | name         | string | 否   | 搜索的电影名称         |
  | type         | number | 否   | 0. 全部.  1. 动作片... |
  | year         | number | 否   | 上映年份               |
  | region       | string | 否   | 地区:  CHINA, USA      |
  | onlyAvaiable | bool   | 否   | 是否仅显示可用源         |
  | offset       | number | 否   |                        |
  | limit        | number | 否   |                        |

- 返回值

  ```json
  {
    "Action": "getcurrentaccount",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
      "list":[
        {
          "id": "",
          "cover": "",
          "url": "",
          "name": "",
          "desc": "",
          "region": "",
          "type": 1,
          "size": 1024,
          "price": 0,
          "source": 10,
          "wallet": "AYMnqA65pJFKAbbpD8hi5gdNDBmeFBy5hS"
        }
      ]
    }, 
    "Version": "1.0.0"
  }
  ```



#### 2. 发布电影

* 请求参数

  | 参数     | 类型   | 必填 | 说明                       |
  | -------- | ------ | ---- | -------------------------- |
  | cover    | string | 是   | 封面图片的base64编码字符串 |
  | url      | string | 是   | 下载链接                   |
  | name     | string | 是   | 电影名称                   |
  | desc     | string | 是   | 简述                       |
  | type     | number | 是   | 类型                       |
  | year     | number | 是   | 上映年份                   |
  | language | string | 是   | 语言. CHINESE, ENGLISH     |
  | region   | string | 否   | 地区                       |
  | price    | number | 否   | 下载支付价格               |

  

* 返回值

  ```json
  {
    "Action": "publish",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
      "tx": ""
    },
    "Version": "1.0.0"
  }
  ```

  

#### 3. 编辑电影

- 请求参数

  | 参数     | 类型   | 必填 | 说明                       |
  | -------- | ------ | ---- | -------------------------- |
  | id       | string | 是   | 需要编辑的电影的id |
  | cover    | string | 否   | 封面图片的base64编码字符串 |
  | url      | string | 否   | 下载链接                   |
  | name     | string | 否   | 电影名称                   |
  | desc     | string | 否   | 简述                       |
  | type     | number | 否   | 类型                       |
  | year     | number | 否   | 上映年份                   |
  | language | string | 否   | 语言. CHINESE, ENGLISH     |
  | region   | string | 否   | 地区                       |
  | price    | number | 否   | 下载支付价格               |

  

- 返回值

  ```json

  {
    "Action": "set",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
      "tx": ""
    },
    "Version": "1.0.0"
  }
  ```

  

#### 4. 获取发布电影列表

- 请求参数

  | 参数   | 类型   | 必填 | 说明 |
  | ------ | ------ | ---- | ---- |
  | offset | number | 否   |      |
  | limit  | number | 否   |      |

- 返回值

  ```json
  {
    "Action": "getPublishList",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
      "list":[
        {
          "id": "",
          "name": "",
          "createdAt": 1559065715,
          "paidCount": 10,
          "profit": "100"
        }
      ],
      "total": 123
    },
    "Version": "1.0.0"
  }
  ```

  

#### 5. 收益明细

- 请求参数

  | 参数  | 类型   | 必填 | 说明                         |
  | ----- | ------ | ---- | ------------------------- |
  | start | number | 是   | 筛选的开始的时间戳，单位到秒 |
  | end   | number | 是   | 筛选的结束的时间戳，单位到秒 |
  | offset | number | 否   |      |
  | limit  | number | 否   |      |

- 返回值

  ```json
  {
    "Action": "getEarningsDetailed",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
      "list": [
        {
          "id": "",
          "name": "",
          "downloadedAt": 1559065715,
          "profit": "10",
          "totalProfit": "100",
        }
      ],
      "total": 123
    },
    "Version": "1.0.0"
  }
  ```

  

#### 6. 下载记录

- 请求参数

  | 参数  | 类型   | 必填 | 说明                         |
  | ----- | ------ | ---- | ---------------------------- |
  | start | number | 是   | 筛选的开始的时间戳，单位到秒 |
  | end   | number | 是   | 筛选的结束的时间戳，单位到秒 |
  | offset | number | 否   |  |
  | limit  | number | 否   |  |

- 返回值

  ```json
  {
    "Action": "getDownloadRecord",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
      "list": [
        {
          "id": "",
          "name": "",
          "downloadedAt": 1559065715,
          "cost": "0",
          "channelCost": "0.002",
        }
      ],
      "total": 123
    },
    "Version": "1.0.0"
  }
  ```

  

#### 7. 删除发布的电影

- 请求参数

  | 参数 | 类型   | 必填 | 说明          |
  | ---- | ------ | ---- | ------------- |
  | id   | string | 是   | 文件的id |

- 返回值

  ```json
  {
    "Action": "deletePublishFile",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
      "tx": ""
    },
    "Version": "1.0.0"
  }

  ```

#### 8. 删除下载记录

- 请求参数

  | 参数 | 类型   | 必填 | 说明          |
  | ---- | ------ | ---- | ------------- |
  | id   | string | 是   | 记录的id |

- 返回值

  ```json
  {
    "Action": "deleteDownloadRecord",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
      "tx": ""
    },
    "Version": "1.0.0"
  }

  ```  

#### 

