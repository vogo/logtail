# logtail is a web socket log tailing tool.

## 1. Features
- tailing command output
- support multiple command tailing
- support webhook notice (include dingtalk)
- support websocket tailing
- support log matching filter

## 2. Usage

### 2.1. install logtail

`go get -u github.com/vogo/logtail/cmd/logtail`

### 2.2. start logtail server

usage: `logtail -port=<port> -cmd="<cmd>"`

examples:

```bash
# tailing file logs
logtail -port=54321 -cmd="tailf /home/my/logs/myapp.log"

# tailing kubectl logs
logtail -port=54322 -cmd="kubectl logs --tail 10 -f \$(kubectl get pods --selector=app=myapp -o jsonpath='{.items[*].metadata.name}')"

# using a config file
logtail -file=/Users/wongoo/logtail-config.json
```

config file sample:
```json
{
  "port": 54321,
  "servers": [
    {
      "id": "app1",
      "command": "tail -f /Users/wongoo/app/app1.log",
      "routers": [
        {
          "matchers": [
            {
              "match_contains": "ERROR"
            }
          ],
          "transfers": [
            {
              "ding_url": "https://oapi.dingtalk.com/robot/send?access_token=<token>"
            },
            {
              "webhook_url": "http://127.0.0.1:9000"
            }
          ]
        }
      ]
    },
    {
      "id": "app2",
      "command": "tail -f /Users/wongoo/app/app2.log"
    }
  ]
}
```

### 2.3. tailing logs

browse `http://<server-ip>:<port>` to list all tailing logs.


## 3. Architecture

![](https://github.com/vogo/vogo.github.io/raw/master/logtail/logtail-architecture.png)
