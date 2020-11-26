logtail is a web socket log tail tool.

## usage 

1. install logtail: `go get github.com/vogo/logtail/cmd/logtail`

2. start logtail server `./logtail -port=<port> -cmd="<cmd>"`

examples:

```bash
# tailf file logs
./logtail -port=54321 -cmd="tailf /home/my/logs/myapp.log"

# tailf kubectl logs
./logtail -port=54322 -cmd="kubectl logs --tail 10 -f \$(kubectl get pods --selector=app=myapp -o jsonpath='{.items[*].metadata.name}')"
```

3. browse `http://<ip>:<port>` to view tailing logs.
