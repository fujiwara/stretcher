## An example of prepare & execute deployment.

### Step 1: run consul agent

Run a consul agent(server) on your local machine.

e.g.

```
$ consul agent -server -data-dir /tmp/consul -bootstrap-expect 1
```

### Step 2: run `prepare.sh`

`prepare.sh` runs a following process.

1. go build in `./project` directory.
2. Create a deployment archive as `example.tar.gz`.
3. Create a manifest YAML as `example.yml`.
4. Create a consul event for deployment.

After this step, a deployment event is queued in the consul server.

### Step 3: run `exec.sh`

`exec.sh` runs deployment process notified from the consul server.

1. Receive a deployment event by `consul watch`.
2. stretcher runs deployment!
