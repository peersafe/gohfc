# 智能合约脚本使用

## 1. 获取合约调用信息

首先需要从wischain平台获取合约调用的相关信息

### 接口信息

**Path：** http://192.168.0.154:8888/v3/fabric/chaincode/cc_invoke_info (换成自己wischain服务的ip和端口)

**Method：** GET

**接口描述：**


### 请求参数

**Headers**

| 参数名称     | 参数值           | 是否必须 | 示例 | 备注                          |
| ------------ | ---------------- | -------- | ---- | ----------------------------- |
| Content-Type | application/json | 是       |      |                               |
| user         |                  | 是       |      | 用户名                        |
| token        |                  | 是       |      | 用户登陆成功后获取的token文件 |

**Query**

| 参数名称 | 是否必须 | 示例 | 备注         |
| -------- | -------- | ---- | ------------ |
| netuuid  | 是       |      | 网络uuid     |
| orguuid  | 是       |      | 组织uuid     |
| cc_uuid  | 是       |      | 智能合约uuid |

### 返回信息

```json
{
  "errCode": 0,
  "errMsg": "",
  "data": {
    "local_mspid": "org1MSPcsdctzar",
    "channel_id": "channelmawmkgjm",
    "chaincode_name": "cc1csdctzar",
    "apply_cert": {
      "sign_certs": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNyakNDQWxTZ0F3SUJBZ0lVWmxmNmpnT0F4YWNvWUhtaWhRY1NxY3JtS2N3d0NnWUlLb1pJemowRUF3SXcKZWpFTE1Ba0dBMVVFQmhNQ1ZWTXhFekFSQmdOVkJBZ1RDa05oYkdsbWIzSnVhV0V4RmpBVUJnTlZCQWNURFZOaApiaUJHY21GdVkybHpZMjh4SGpBY0JnTlZCQW9URldOaE1TNXZjbWN4TG1OelpHTjBlbUZ5TG1OdmJURWVNQndHCkExVUVBeE1WWTJFeExtOXlaekV1WTNOa1kzUjZZWEl1WTI5dE1CNFhEVEU1TVRJd05UQTVOREV3TUZvWERUTTAKTVRJd01UQTNNRFl3TUZvd1hURUxNQWtHQTFVRUJoTUNWVk14RnpBVkJnTlZCQWdURGs1dmNuUm9JRU5oY205cwphVzVoTVJRd0VnWURWUVFLRXd0SWVYQmxjbXhsWkdkbGNqRVBNQTBHQTFVRUN4TUdZMnhwWlc1ME1RNHdEQVlEClZRUURFd1YxYzJWeU56QlpNQk1HQnlxR1NNNDlBZ0VHQ0NxR1NNNDlBd0VIQTBJQUJFQnM2VFVtVTVnQXlyTTcKeFJSZ1B5ajFCcUZWemRUOE5FQWZtbEVJVWYwdUNYSS9PZTB1a3laT2J4bm01Ykljc2tDUTM2NTdhZUlXMjJoRApjdWoxakZLamdkUXdnZEV3RGdZRFZSMFBBUUgvQkFRREFnZUFNQXdHQTFVZEV3RUIvd1FDTUFBd0hRWURWUjBPCkJCWUVGTGEySFowYmZmLzkzZzRXL3kvYkVTcGZCTVl6TUI4R0ExVWRJd1FZTUJhQUZJZkR4bUZ5U3RzV1hxU1AKZkQwZjYrK3FHZXd6TUJjR0ExVWRFUVFRTUE2Q0REZGlaRGxqTURGak1qVTRZVEJZQmdncUF3UUZCZ2NJQVFSTQpleUpoZEhSeWN5STZleUpvWmk1QlptWnBiR2xoZEdsdmJpSTZJaUlzSW1obUxrVnVjbTlzYkcxbGJuUkpSQ0k2CkluVnpaWEkzSWl3aWFHWXVWSGx3WlNJNkltTnNhV1Z1ZENKOWZUQUtCZ2dxaGtqT1BRUURBZ05JQURCRkFpRUEKaXcwZ0gxSGxRd0RRZHlVbkhsWGtDZ0tFd29KWVRKTUJvTENvWnBmb0pab0NJR0lFbkJoRS9LdFU4V3ZaWXltbwpVRGpwSmxseXJSUm04cFRocWVPbVpmMDMKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=",
      "private_key": "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JR0hBZ0VBTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEJHMHdhd0lCQVFRZ1BKMFFSUm1Pb1c3UzNVbjYKWkY4QWRNUDdyV2VmMVpJSlN6OUFLUW9rRTZDaFJBTkNBQVJBYk9rMUpsT1lBTXF6TzhVVVlEOG85UWFoVmMzVQovRFJBSDVwUkNGSDlMZ2x5UHpudExwTW1UbThaNXVXeUhMSkFrTit1ZTJuaUZ0dG9RM0xvOVl4UwotLS0tLUVORCBQUklWQVRFIEtFWS0tLS0tCg=="
    },
    "cc_org_other": {
      "ccuuid": "cc1csdctzar",
      "cc_name": "cc1",
      "peers": [
        {
          "id": 3,
          "cc_uuid": "cc1csdctzar",
          "type": "peer",
          "node_uuid": "peer1.org1.csdctzar.com",
          "domain": "peer1.org1.csdctzar.com",
          "ip": "192.168.0.16",
          "port": "7051",
          "tls_cert": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNPakNDQWVHZ0F3SUJBZ0lVQzM2L3ZJZ1JrSVp3UzhaYnZiSzBqdGVPeERVd0NnWUlLb1pJemowRUF3SXcKZWpFTE1Ba0dBMVVFQmhNQ1ZWTXhFekFSQmdOVkJBZ1RDa05oYkdsbWIzSnVhV0V4RmpBVUJnTlZCQWNURFZOaApiaUJHY21GdVkybHpZMjh4SGpBY0JnTlZCQW9URldOaE1TNXZjbWN4TG1OelpHTjBlbUZ5TG1OdmJURWVNQndHCkExVUVBeE1WWTJFeExtOXlaekV1WTNOa1kzUjZZWEl1WTI5dE1CNFhEVEU1TVRJd05UQTNNRFl3TUZvWERUTTAKTVRJd01UQTNNRFl3TUZvd2VqRUxNQWtHQTFVRUJoTUNWVk14RXpBUkJnTlZCQWdUQ2tOaGJHbG1iM0p1YVdFeApGakFVQmdOVkJBY1REVk5oYmlCR2NtRnVZMmx6WTI4eEhqQWNCZ05WQkFvVEZXTmhNUzV2Y21jeExtTnpaR04wCmVtRnlMbU52YlRFZU1Cd0dBMVVFQXhNVlkyRXhMbTl5WnpFdVkzTmtZM1I2WVhJdVkyOXRNRmt3RXdZSEtvWkkKemowQ0FRWUlLb1pJemowREFRY0RRZ0FFY2ZPaXRlSFV3WE1SVkVUUFUyL0xBZWRyMERGK3hXK3d1MHBocndOZwppQ0xCcXB1a1V1OWRkdUs2Q3hNUnJWTlB2aDA4eDFhelNrTXBKZElYdXZRR2pLTkZNRU13RGdZRFZSMFBBUUgvCkJBUURBZ0VHTUJJR0ExVWRFd0VCL3dRSU1BWUJBZjhDQVFFd0hRWURWUjBPQkJZRUZJZkR4bUZ5U3RzV1hxU1AKZkQwZjYrK3FHZXd6TUFvR0NDcUdTTTQ5QkFNQ0EwY0FNRVFDSUZlamVkTzB0TFUrOHZGOGlDS0ZPcGNGZUp0bwpuQVBVbkVvSEgzU2JpQUt5QWlBakFNWGR3VnVmbWZEQkNvSk81WFlxR2MxU3YzUFNLU1ZtSERuMkNOS1NpZz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
        }
      ],
      "orders": [
        {
          "id": 1,
          "cc_uuid": "cc1csdctzar",
          "type": "orderer",
          "node_uuid": "orderer1.ordererorg0.csdctzar",
          "domain": "orderer1.ordererorg0.csdctzar",
          "ip": "192.168.0.16",
          "port": "7050",
          "tls_cert": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNXRENDQWYrZ0F3SUJBZ0lVUkwzbXgzRFB0R3d5RkRvc3FFRnhVeWtKaXZRd0NnWUlLb1pJemowRUF3SXcKZ1lneEN6QUpCZ05WQkFZVEFsVlRNUk13RVFZRFZRUUlFd3BEWVd4cFptOXlibWxoTVJZd0ZBWURWUVFIRXcxVApZVzRnUm5KaGJtTnBjMk52TVNVd0l3WURWUVFLRXh4allURXViM0prWlhKbGNtOXlaekF1WTNOa1kzUjZZWEl1ClkyOXRNU1V3SXdZRFZRUURFeHhqWVRFdWIzSmtaWEpsY205eVp6QXVZM05rWTNSNllYSXVZMjl0TUI0WERURTUKTVRJd05UQTNNRFF3TUZvWERUTTBNVEl3TVRBM01EUXdNRm93Z1lneEN6QUpCZ05WQkFZVEFsVlRNUk13RVFZRApWUVFJRXdwRFlXeHBabTl5Ym1saE1SWXdGQVlEVlFRSEV3MVRZVzRnUm5KaGJtTnBjMk52TVNVd0l3WURWUVFLCkV4eGpZVEV1YjNKa1pYSmxjbTl5WnpBdVkzTmtZM1I2WVhJdVkyOXRNU1V3SXdZRFZRUURFeHhqWVRFdWIzSmsKWlhKbGNtOXlaekF1WTNOa1kzUjZZWEl1WTI5dE1Ga3dFd1lIS29aSXpqMENBUVlJS29aSXpqMERBUWNEUWdBRQpFZlNaU0tuallxRFdiRFJJZms0Q0VBd213NjloS2hjODZuclNOV1lhVnJnOE42czZSTFlUNnRWdHFEdXdyT2I3CjNXM0JSZEFudGFMT0xiQ1FtZzFxSEtORk1FTXdEZ1lEVlIwUEFRSC9CQVFEQWdFR01CSUdBMVVkRXdFQi93UUkKTUFZQkFmOENBUUV3SFFZRFZSME9CQllFRkQ5RENzUU44bElZQWUvcHJtQVVMMUdiSmJ6ME1Bb0dDQ3FHU000OQpCQU1DQTBjQU1FUUNJRkJ5akU3eWxscFE0dWFMRDk2RGt3Ym9kSGdNZE92Qm1vWnpaVnBDWUxOM0FpQkVRVmZKCnIxNTM1WUFoY2R5V0N0cHovcXppN3l1UVBZMlJJM0FoeFZKTVFRPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
        },
        {
          "id": 2,
          "cc_uuid": "cc1csdctzar",
          "type": "orderer",
          "node_uuid": "orderer2.ordererorg0.csdctzar",
          "domain": "orderer2.ordererorg0.csdctzar",
          "ip": "192.168.0.16",
          "port": "7150",
          "tls_cert": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNXRENDQWYrZ0F3SUJBZ0lVUkwzbXgzRFB0R3d5RkRvc3FFRnhVeWtKaXZRd0NnWUlLb1pJemowRUF3SXcKZ1lneEN6QUpCZ05WQkFZVEFsVlRNUk13RVFZRFZRUUlFd3BEWVd4cFptOXlibWxoTVJZd0ZBWURWUVFIRXcxVApZVzRnUm5KaGJtTnBjMk52TVNVd0l3WURWUVFLRXh4allURXViM0prWlhKbGNtOXlaekF1WTNOa1kzUjZZWEl1ClkyOXRNU1V3SXdZRFZRUURFeHhqWVRFdWIzSmtaWEpsY205eVp6QXVZM05rWTNSNllYSXVZMjl0TUI0WERURTUKTVRJd05UQTNNRFF3TUZvWERUTTBNVEl3TVRBM01EUXdNRm93Z1lneEN6QUpCZ05WQkFZVEFsVlRNUk13RVFZRApWUVFJRXdwRFlXeHBabTl5Ym1saE1SWXdGQVlEVlFRSEV3MVRZVzRnUm5KaGJtTnBjMk52TVNVd0l3WURWUVFLCkV4eGpZVEV1YjNKa1pYSmxjbTl5WnpBdVkzTmtZM1I2WVhJdVkyOXRNU1V3SXdZRFZRUURFeHhqWVRFdWIzSmsKWlhKbGNtOXlaekF1WTNOa1kzUjZZWEl1WTI5dE1Ga3dFd1lIS29aSXpqMENBUVlJS29aSXpqMERBUWNEUWdBRQpFZlNaU0tuallxRFdiRFJJZms0Q0VBd213NjloS2hjODZuclNOV1lhVnJnOE42czZSTFlUNnRWdHFEdXdyT2I3CjNXM0JSZEFudGFMT0xiQ1FtZzFxSEtORk1FTXdEZ1lEVlIwUEFRSC9CQVFEQWdFR01CSUdBMVVkRXdFQi93UUkKTUFZQkFmOENBUUV3SFFZRFZSME9CQllFRkQ5RENzUU44bElZQWUvcHJtQVVMMUdiSmJ6ME1Bb0dDQ3FHU000OQpCQU1DQTBjQU1FUUNJRkJ5akU3eWxscFE0dWFMRDk2RGt3Ym9kSGdNZE92Qm1vWnpaVnBDWUxOM0FpQkVRVmZKCnIxNTM1WUFoY2R5V0N0cHovcXppN3l1UVBZMlJJM0FoeFZKTVFRPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
        }
      ]
    }
  }
}
```



## 2. 修改当前目录下的client.yaml配置文件

```shell
crypto:
#   family: gm                #国密版配置
#   algorithm: P256SM2
#   hash: SM3
   family: ecdsa              #非国密版配置
   algorithm: P256-SHA256
   hash: SHA2-256
orderers:   #将获取合约调用信息的orderer的相关信息对应填写
  orderer0:
    host: 192.168.0.16:7050  #orderer的ip
    domainName: orderer1.ordererorg0.csdctzar   #orderer的域名
    useTLS: true
    #tlsPath: 将获取合约调用信息的orderer的tls_cert内容保存到本地，然后将对应的绝对路径写在下面
    tlsPath:
/opt/gopath/src/github.com/peersafe/gohfc/testInterface/test/orderer/orderer1.ordererorg0.csdctzar/server.crt
peers:  #将获取合约调用信息的peer的相关信息对应填写
  peer01:
    host: 192.168.0.16:7051    #peer的ip
    domainName: peer1.org1.csdctzar.com    #peer的域名
    useTLS: true
    #tlsPath: 将获取合约调用信息的peer的tls_cert内容保存到本地，然后将对应的绝对路径写在下面
    tlsPath: /opt/gopath/src/github.com/peersafe/gohfc/testInterface/test/peer/ca.crt
eventPeers:  #将获取合约调用信息的peer的相关信息对应填写
  peer01:
    host: 192.168.0.16:7051  #peer的ip
    domainName: peer1.org1.csdctzar.com   #peer的域名
    useTLS: true
    #tlsPath: 将获取合约调用信息的peer的tls_cert内容保存到本地，然后将对应的绝对路径写在下面
    tlsPath: /opt/gopath/src/github.com/peersafe/gohfc/testInterface/test/peer/ca.crt
channel:
		#mspConfigPath: 
		#1. 将获取合约调用信息申请的证书sign_certs内容保存到自己本地创建的msp/signcerts/cert.pem
		#2. 将获取合约调用信息申请的证书private_key内容保存到自己本地创建的msp/keystore/private.key内容保存到本地的
		#3. 将msp的绝对路径作为mspConfigPath
    mspConfigPath: /opt/gopath/src/github.com/peersafe/gohfc/testInterface/test/msp
    localMspId: org1MSPcsdctzar   #组织的mspid
    channelId: channelmawmkgjm    #channel的uuid
    chaincodeName: cc1csdctzar    #chaincode的uuid
log:
    logLevel: DEBUG
```

## 3. build.sh脚本使用

```shell
#编译api二进制
./buil.sh
#生成testInterface二进制
```

## 4. api二进制的使用

```shell
#监听区块
./testInterface -function=listenfull

#查询
./testInterface -function=query  a   #查询a的账户余额

#调用转账
./testInterface -function=invoke   #a用户给b用户每次转1

```



