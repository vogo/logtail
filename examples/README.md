# Examples

## 1. install logtail

install logtail binary to current directory:
```bash
GOBIN=$(dir) go install github.com/vogo/logtail@master
```

or download the binary from [release page](https://github.com/vogo/logtail/releases/).

## 2. start logtail

```bash
./logtail -file <config-file>
```

## 3. config file examples
- tail a file and send error log to console: [logtail-tail-to-console.json](logtail-tail-to-console.json)
- tail a file and write error log to directory: [logtail-tail-to-dir.json](logtail-tail-to-dir.json)
- tail all files under a given directory and send error log to [lark webhook](https://open.feishu.cn/document/ukTMukTMukTM/ucTM5YjL3ETO24yNxkjN): [logtail-dir-to-lark.json](logtail-dir-to-lark.json)
- tail all files under a given directory and send error log to [dingtalk webhook](https://open.dingtalk.com/document/robots/custom-robot-access) : [logtail-dir-to-ding.json](logtail-dir-to-ding.json)
