
# iot信令网关(websocket).
## 前置说明
iot信令网关(websocket)开放给企业用户app或者业务后台使用,实现iot开放平台接入,采用websocket通信，地址为/api/v1/ws/conv

## 通信帧格式

| 字段       | 长度 | 类型       | 说明                                       |
| ---------- | ---- | ---------- | ------------------------------------------ |
| Version    | 1    | uint8_t    | 固定为1                                    |
| MsgType    | 3    | uint8_t [] | 消息类型                                   |
| Timestamp  | 8    | int64_t    | 时间戳                                     |
| Via        | 32   | uint8_t [] | 应用服务id                                 |
| MsgIdLen   | 1    | uint8_t    | 消息id长度                                 |
| MsgId      | 变长 | uint8_t [] | 消息id，长度等于MsgIdLen存储的值           |
| FromLen    | 1    | uint8_t    | FromId长度                                 |
| FromId     | 变长 | uint8_t [] | From id，长度等于FromLen存储的值           |
| ToLen      | 1    | uint8_t    | ToId长度                                   |
| ToId       | 变长 | uint8_t [] | To id，长度等于ToLen存储的值               |
| PayloadLen | 4    | uint32_t   | 负载长度                                   |
| Payload    | 变长 | uint8_t [] | 负载信息（目前为json字符串形式的消息报文） |

## 消息类型

| 编号 | 内容                |
| ---- | ------------------- |
| 0x41 | APP登录设备鉴权请求, App与AppSvc专用        |
| 0x42 | APP登录设备鉴权响应, App与AppSvc专用        |
| 0x81 | MQTT转Websocket消息, 透传 |
| 0x82 | Websocket转MQTT消息, 透传 |

## 负载描述Payload

各消息类型对应请求与响应报文, 先将结构体转换为JSON字符串, 而作为负载附加于定长头部后, 形成通信帧。不同类型对应不同结构体。

### websocket连接请求
请求资源: /api/v1/ws/conn
通过http头: App设置签名, 参考如下:
`````go

func getAppSign() {
    idAndRealm := "uid@enterprise_id" // 用户Id@企业Id
    prod := "prod0001"  // 产品Id
    bodySign := "d41d8cd98f00b204e9800998ecf8427e" // post请求的body md5签名
    date_ms := time.Now().UnixNano() / 1e6
    requestline := "/api/v1/ws/conn"
    algo := "1" 
    appkey := "test"
    appsecret := "test"

    // 1. 拼接签名字符串, 参与签名的字段 idAndRealm, prod, bodySign, date, appkey, requestline.
    fileds := []string{
        "from: " + idAndRealm,
        "prod: " + prod,
        "abstract: " + bodySign,
        "date: " + date_ms,
        "appkey: " + appkey,
        "GET " + requestline + " HTTP/1.1",
    }

    // 2. 签名结果: base64(hmac-sha1).
    sha := base64.StdEncoding.EncodeToString([]byte(a.hmac(appsecret, strings.Join(fileds, "\n"))))

    // 3. 将编码后的字符串url encode后添加到url后面.
    params := url.Values{}
    params.Add("ak", appkey)        // appkey.
    params.Add("dt", date)          // date.
    params.Add("fm", idAndRealm)    // from.
    params.Add("pd", prod)          // prod id.
    params.Add("st", bodySign)      // body sign.
    params.Add("lg", algo)          // algo 1 or 2.
    params.Add("sg", sha)           // sign.
    return base64.StdEncoding.EncodeToString([]byte(params.Encode()))
}

`````

### APP登录设备鉴权请求
请求资源: /api/v1/ws/login

#### - 请求体
|   名称    |  类型  | 必填 |                             说明                             |
| :-------: | :----: | :--: | :----------------------------------------------------------: |
| deviceId  | string |  是  |             设备Id             |
| token  | string |  是  | 登录Token, 由appkey, secret签名算法生成, 同上 |

#### - 响应体

| 名称 |  类型  | 必填 |            说明            |
| :--: | :----: | :--: | :------------------------: |
| code | string |  是  | 错误码, 0: 成功, 非0: 失败 |
| msg  | string |  是  |          错误描述          |

## 视频直播相关
1. 发起startSendStream请求到设备
2. 建立p2p连接p2pTask
3. p2p内部控制指令及媒体流
### 视频实时直播【发起】（ws到设备）
#### - 请求体

|   名称    |  类型  | 必填 |                             说明                             |
| :-------: | :----: | :--: | :----------------------------------------------------------: |
| msg_name      | string |  是  | startSendStream    |
| msg_index     | string |  是  | 128   |
| stream_type   | string |  是  | 0-主码流，1-子码流，2-最次码流   |

#### - 响应体

| 名称 |  类型  | 必填 |            说明            |
| :--: | :----: | :--: | :------------------------: |
| code | string |  是  | 错误码，0：成功，非0：失败 |
| msg  | string |  是  |          错误描述          |
| data  | json |  是  |          数据          |
| reverse_type | int |  是  |  图像翻转类型:0-正常、1-水平翻转、2-垂直翻转、3-水平垂直翻转          |
| sps | string |  是  | 视码流sps信息,base64编码          |
| pps | string |  是  | 视码流pps信息,base64编码          |
| sub_sps | string |  是  | 子码流sps信息,base64编码          |
| sub_pps | string |  是  | 子码流pps信息,base64编码          |
| thr_sps | string |  是  | 次码流sps信息,base64编码          |
| thr_pps | string |  是  | 次码流pps信息,base64编码          |


### 视频实时直播【播放】（p2p到设备）
#### - 请求体

|   名称    |  类型  | 必填 |                             说明                             |
| :-------: | :----: | :--: | :----------------------------------------------------------: |
| msg_name      | string |  是  | startVideoLive    |
| msg_index     | string |  是  | 129   |

#### - 响应体

| 名称 |  类型  | 必填 |            说明            |
| :--: | :----: | :--: | :------------------------: |
| code | string |  是  | 错误码，0：成功，非0：失败 |
| msg  | string |  是  |          错误描述          |
| data  | json |  是  |          数据          |
| audio_type | int |  是  |  音频类型          |
| sample_rate | int |  是  | 音频采样          |
| audio_ratio | int |  是  | 音频比特率,0为无效值,目前用于g726          |

### 视频实时直播【关闭】（p2p到设备）
#### - 请求体

|   名称    |  类型  | 必填 |                             说明                             |
| :-------: | :----: | :--: | :----------------------------------------------------------: |
| msg_name      | string |  是  | closeVideoLive    |
| msg_index     | string |  是  | 130   |

#### - 响应体

| 名称 |  类型  | 必填 |            说明            |
| :--: | :----: | :--: | :------------------------: |
| code | string |  是  | 错误码，0：成功，非0：失败 |
| msg  | string |  是  |          错误描述          |
| data  | json |  是  |          数据          |

### 视频实时直播【控制指令】（p2p到设备）
#### - 请求体

|   名称    |  类型  | 必填 |                             说明                             |
| :-------: | :----: | :--: | :----------------------------------------------------------: |
| msg_name      | string |  是  | controlVideoLive    |
| msg_index     | string |  是  | 131   |
| cmd_type     | string |  是  | pause/resume/mute/unmute   |

#### - 响应体

| 名称 |  类型  | 必填 |            说明            |
| :--: | :----: | :--: | :------------------------: |
| code | string |  是  | 错误码，0：成功，非0：失败 |
| msg  | string |  是  |          错误描述          |

### 视频实时直播【双向对讲】（p2p到设备）
#### - 请求体

|   名称    |  类型  | 必填 |                             说明                             |
| :-------: | :----: | :--: | :----------------------------------------------------------: |
| msg_name      | string |  是  | startTalkback    |
| msg_index     | string |  是  | 132   |

#### - 响应体

| 名称 |  类型  | 必填 |            说明            |
| :--: | :----: | :--: | :------------------------: |
| code | string |  是  | 错误码，0：成功，非0：失败 |
| msg  | string |  是  |          错误描述          |

### 录像回放【查询列表（ws/p2p到设备）
#### - 请求体

|   名称    |  类型  | 必填 |                             说明                             |
| :-------: | :----: | :--: | :----------------------------------------------------------: |
| msg_name      | string |  是  | getVideoPlaybackFiles    |
| msg_index     | string |  是  | 133   |
| offset     | int |  是  | 索引下表   |
| limit     | int |  是  | 返回最大数量   |

#### - 响应体

| 名称 |  类型  | 必填 |            说明            |
| :--: | :----: | :--: | :------------------------: |
| code | string |  是  | 错误码，0：成功，非0：失败 |
| msg  | string |  是  |          错误描述          |
| data  | json |  是  |          数据结构          |
| total  | int |  是  |   文件总数量          |
| file_paths  | []string |  是  | 文件列表          |

### 录像回放【播放】（ws/p2p到设备）
#### - 请求体

|   名称    |  类型  | 必填 |                             说明                             |
| :-------: | :----: | :--: | :----------------------------------------------------------: |
| msg_name      | string |  是  | startPlayVideoPlayback    |
| msg_index     | string |  是  | 134   |
| file_path     | string |  是  | 录像文件   |
| play_sec     | int |  是  |  从第几秒开始播放   |

#### - 响应体

| 名称 |  类型  | 必填 |            说明            |
| :--: | :----: | :--: | :------------------------: |
| code | string |  是  | 错误码，0：成功，非0：失败 |
| msg  | string |  是  |          错误描述          |
| data  | json |  是  |          数据结构          |
| sps  | string |  是  |   视频sps          |
| pps  | string |  是  |   视频pps          |
| audio_type | int |  是  |  音频类型          |
| sample_rate | int |  是  | 音频采样          |
| audio_ratio | int |  是  | 音频比特率,0为无效值,目前用于g726          |

### 录像回放【关闭】（ws/p2p到设备）
#### - 请求体

|   名称    |  类型  | 必填 |                             说明                             |
| :-------: | :----: | :--: | :----------------------------------------------------------: |
| msg_name      | string |  是  | closePlayVideoPlayback    |
| msg_index     | string |  是  | 135   |
| file_path     | string |  是  | 录像文件   |

#### - 响应体

| 名称 |  类型  | 必填 |            说明            |
| :--: | :----: | :--: | :------------------------: |
| code | string |  是  | 错误码，0：成功，非0：失败 |
| msg  | string |  是  |          错误描述          |


### 录像回放【控制】（ws/p2p到设备）
#### - 请求体
|   名称    |  类型  | 必填 |                             说明                             |
| :-------: | :----: | :--: | :----------------------------------------------------------: |
| msg_name      | string |  是  | controlPlayVideoPlayback    |
| msg_index     | string |  是  | 135   |
| file_path     | string |  是  | 录像文件   |
| cmd_type     | string |  是  | pause/resume/mute/unmute   |

#### - 响应体
| 名称 |  类型  | 必填 |            说明            |
| :--: | :----: | :--: | :------------------------: |
| code | string |  是  | 错误码，0：成功，非0：失败 |
| msg  | string |  是  |          错误描述          |
| data  | json |  是  |          数据结构          |


## p2p通道建立相关
### p2p任务【申请】主动端
#### - 请求体
|   名称    |  类型  | 必填 |                             说明                             |
| :-------: | :----: | :--: | :----------------------------------------------------------: |
| msg_name      | string |  是  | p2pRequest    |
| msg_index     | string |  是  | 136   |
| cmd_type      | string |  是  | rpass   |

#### - 响应体【无】

### p2p任务【返回passId】被动端
#### - 请求体
|   名称    |  类型  | 必填 |                             说明                             |
| :-------: | :----: | :--: | :----------------------------------------------------------: |
| msg_name      | string |  是  | p2pResponse    |
| msg_index     | string |  是  | 137   |
| cmd_type      | string |  是  | spass   |
| pass_id       | string |  是  | 随机生成PassId   |

#### - 响应体【无】

### p2p任务【发送offer/answer】主动端
#### - 请求体
|   名称    |  类型  | 必填 |                             说明                             |
| :-------: | :----: | :--: | :----------------------------------------------------------: |
| msg_name      | string |  是  | p2pOfferOrAnswer    |
| msg_index     | string |  是  | 138   |
| cmd_type      | string |  是  | offer/answer   |
| pass_id       | string |  是  | 返回的PassId   |
| description   | string |  是  | offer sdp信息   |

#### - 响应体【无】


### p2p任务【发送candiate】主动端
#### - 请求体
|   名称    |  类型  | 必填 |                             说明                             |
| :-------: | :----: | :--: | :----------------------------------------------------------: |
| msg_name      | string |  是  | p2pCandidate    |
| msg_index     | string |  是  | 139   |
| cmd_type      | string |  是  | candidate   |
| pass_id       | string |  是  | 返回的PassId   |
| candidate     | string |  是  | ice candidate信息   |
| mid           | string |  是  | ice mid信息   |

#### - 响应体【无】