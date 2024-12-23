
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
| 0x01 | 登录请求            |
| 0x02 | 登录响应            |
| 0x03 | 连接设备请求        |
| 0x04 | 连接设备响应        |
| 0x05 | 断开设备连接请求        |
| 0x06 | 断开设备连接响应        |
| 0x41 | APP登录设备鉴权请求，App与AppSvc专用        |
| 0x42 | APP登录设备鉴权响应，App与AppSvc专用        |
| 0x81 | MQTT转Websocket消息，透传 |
| 0x82 | Websocket转MQTT消息，透传 |

## 负载描述Payload

各消息类型对应请求与响应报文，先将结构体转换为JSON字符串，而作为负载附加于定长头部后，形成通信帧。不同类型对应不同结构体。

### websocket连接请求
请求资源: /api/v1/ws/conn
通过http头: App设置签名

|   名称    |   类型   | 必填 |         说明         |
| :-------: | :------: | :--: | :------------------: |

### websocket连接响应

|   名称   |   类型   | 必填 |            说明            |
| :------: | :------: | :--: | :------------------------: |


### APP登录设备鉴权请求
请求资源: /api/v1/ws/login

|   名称    |  类型  | 必填 |                             说明                             |
| :-------: | :----: | :--: | :----------------------------------------------------------: |
| deviceId  | string |  是  |             设备Id             |
| token  | string |  是  | 登录Token, 由appkey，secret签名算法生成 |

### APP登录设备鉴权响应

| 名称 |  类型  | 必填 |            说明            |
| :--: | :----: | :--: | :------------------------: |
| code | string |  是  | 错误码，0：成功，非0：失败 |
| msg  | string |  是  |          错误描述          |

## 视频直播相关
### 视频实时直播【发起】（到设备）
#### - 请求体

|   名称    |  类型  | 必填 |                             说明                             |
| :-------: | :----: | :--: | :----------------------------------------------------------: |
| msg_name      | string |  是  | startSendStream    |
| msg_index     | string |  是  | 128   |
| device_id     | string |  是  | 设备Id   |
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
| app_id     | int |  是  | AppId   |
| device_id     | string |  是  | 设备Id   |

#### - 响应体

| 名称 |  类型  | 必填 |            说明            |
| :--: | :----: | :--: | :------------------------: |
| code | string |  是  | 错误码，0：成功，非0：失败 |
| msg  | string |  是  |          错误描述          |
| data  | json |  是  |          数据          |
| audioType | int |  是  |  音频类型          |
| sampleRate | int |  是  | 音频采样          |
| audioRatio | int |  是  | 音频比特率,0为无效值,目前用于g726          |

### 视频实时直播【关闭】（p2p到设备）
#### - 请求体

|   名称    |  类型  | 必填 |                             说明                             |
| :-------: | :----: | :--: | :----------------------------------------------------------: |
| msg_name      | string |  是  | closeVideoLive    |
| msg_index     | string |  是  | 130   |
| app_id     | int |  是  | AppId   |
| device_id     | string |  是  | 设备Id   |

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
| app_id     | int |  是  | AppId   |
| device_id     | string |  是  | 设备Id   |
| control_type     | string |  是  | pause/resume/mute/unmute   |

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
| app_id     | int |  是  | AppId   |
| device_id     | string |  是  | 设备Id   |

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
| data.total  | json |  是  |   文件总数量          |
| data.file_paths  | []string |  是  | 文件列表          |

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
| data.sps  | json |  是  |   视频sps          |
| data.pps  | json |  是  |   视频pps          |
| data.audioType  | []string |  是  | 音频类型          |
| data.sampleRate  | []string |  是  | 音频采样          |
| data.audioRatio  | []string |  是  | 音频比特率          |

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
| control_type     | string |  是  | pause/resume/mute/unmute   |

#### - 响应体
| 名称 |  类型  | 必填 |            说明            |
| :--: | :----: | :--: | :------------------------: |
| code | string |  是  | 错误码，0：成功，非0：失败 |
| msg  | string |  是  |          错误描述          |
| data  | json |  是  |          数据结构          |
