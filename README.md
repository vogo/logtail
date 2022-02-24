# logtail is a log tailing utility.

[![codecov](https://codecov.io/gh/vogo/logtail/branch/master/graph/badge.svg)](https://codecov.io/gh/vogo/logtail)
[![GoDoc](https://godoc.org/github.com/vogo/logtail?status.svg)](https://godoc.org/github.com/vogo/logtail)
![license](https://img.shields.io/badge/license-Apache--2.0-green.svg)

## 1. Features
- tailing command output
- support watch files under directory/sub-directories
- support (dynamically) multiple commands tailing
- support websocket tailing
- support log matching filter
- support transfer log to console/file/webhook (include dingtalk)

## 2. Architecture

![](https://github.com/vogo/vogo.github.io/raw/master/logtail/logtail-architecture.png)

## 3. Usage

## 3.2. install logtail

`go install github.com/vogo/logtail/cmd/logtail@master`

### 3.2. Start logtail server using config file

```bash
logtail -file <config-file>
```

config file [examples](examples/README.md).

### 3.3. Start logtail server and config using webapi

usage: `logtail -port=<port>`
```bash
# start at port 54321
logtail -port=54321
```
Then config using [web api](webapi/README.md) from web page `http://<server-ip>:<port>/manage`.

And you can browse `http://<server-ip>:<port>` to list all tailing logs.

## 4. log format

You can config log format globally, or config it for a server.

The config `prefix` of the format is the wildcard of the prefix of a new log record,
`logtail` will check whether a new line is the Start of a new log record, or one of the following lines.

The wildcard does NOT support '*' for none or many chars, it supports the following tag:
- '?' as one byte char;
- '~' as one alphabet char;
- '!' as one number char;
- other chars must exact match.

example:
```bash
{
    "default_format":{
      "prefix": "!!!!-!!-!!"  # global format config, matches 2020-12-12
    },
    "servers": [
        {
          "name": "app1",
          "command": "tail -f /logs/app/app1.log",
          "format":{
            "prefix": "!!!!-!!-!!" # server format config, matches 2020-12-12
          }
        }
    ]
}
```

## 4. command examples

The following are some useful commands which can be used in logtail.

```bash
# tail log file
tail -f /usr/local/myapp/myapp.log

# k8s: find and tail logs for the single pod of myapp
kubectl logs --tail 10 -f $(kubectl get pods --selector=app=myapp -o jsonpath='{.items[*].metadata.name}')

# k8s: find and tail logs for the myapp deployment (multiple pods)
kubectl logs --tail 10 -f deployment/$(kubectl get deployments --selector=project-name=myapp -o jsonpath='{.items[*].metadata.name}')

# k8s: find and tail logs for the latest version of the myapp deployment (single pod)
s=$(kubectl get deployments --selector=project-name=myapp -o jsonpath='{.items[*].metadata.name}');s=${s##* };kubectl logs --tail 10 -f deployment/$s

# k8s: find and tail logs for the latest version of the myapp deployment (multiple pods)
app=$(kubectl get deployments --selector=project-name=myapp -o jsonpath='{.items[*].metadata.name}');app=${app##* };pods=$(kubectl get pods --selector=app=$app -o jsonpath='{.items[*].metadata.name}');cmd='';for pod in $pods; do cmd=$cmd'kubectl logs --tail 2 -f pod/'$pod$'\n'; done;cmd=${cmd::-1}; echo "$cmd"
```