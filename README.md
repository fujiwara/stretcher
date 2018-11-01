Stretcher
=========

A deployment tool with [Consul](https://consul.io) / [Serf](https://www.serfdom.io/) event.

## Example manifest

```yml
src: s3://example.com/app.tar.gz
checksum: e0840daaa97cd2cf2175f9e5d133ffb3324a2b93
dest: /home/stretcher/app
commands:
  pre:
    - echo 'staring deploy'
  post:
    - echo 'deploy done'
  success:
    - echo 'deploy success'
  failure:
    - echo 'deploy failed!!'
    - cat >> /path/to/failure.log
excludes:
  - "*.pid"
  - "*.socket"
```

## Run

### stretcher agent

```
$ stretcher -h
Usage of stretcher:
  -max-bandwidth string
        max bandwidth for download src archives (Bytes/sec)
  -random-delay float
        sleep [0,random-delay) sec on start
  -retry int
        retry count for download src archives
  -retry-wait int
        wait for retry download src archives (sec) (default 3)
  -rsync-verbose string
        rsync verbose option (default -v)
  -timeout int
        timeout for download src archives (sec)
  -v    show version
  -version
        show version
```

#### with Consul

A stretcher agent is designed as running under "consul watch" and will be kicked by [Consul](https://consul.io) event.

```
$ consul watch -type event -name deploy /path/to/stretcher
```

* `-name`: your deployment identity name.

#### with Serf

A stretcher agent can be running as [Serf](https://www.serfdom.io/) event handler.

```
$ serf agent -event-handler="user:deploy=/path/to/stretcher >> /path/to/stretcher.log 2>&1"
```

#### Load AWS credentials

When you specify a S3 URL in manifest, requires a AWS credential setting one of below.

- ~/.aws/config and ~/.aws/credentials (overridden by `AWS_CONFIG_FILE` environment variable.)
  - `AWS_DEFAULT_PROFILE` is supported to select a profile from multiple credentials in file.
- `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_DEFAULT_REGION` environment variable.
- EC2 IAM role.
  - requires `AWS_DEFAULT_REGION` environment variable.

#### Load GCP credentials

When you specify a GS(Google Cloud Storage) URL in manifest, requires a GCP credential setting one of below.

- ServiceAccount
  - requires `GOOGLE_APPLICATION_CREDENTIALS=[PATH]` environment variable.
  - Replace [PATH] with the file path of the JSON file that contains your service account key.
- DefaultAccount
  - If the environment variable isn't set, load the default service account that Compute Engine provide, for applications that run on those services.

### Deployment process

#### Preparing

This process is not included in a stretcher agent.

1. Create a tar(or tar.gz) archive for deployment.
2. Upload the archive file to remote server (S3 or HTTP(S)).
3. Create a manifest file (YAML) and upload it to remote server.

#### Executing with Consul

Create a consul event to kick stretcher agents.

```
$ consul event -name [event_name] [manifest_url]
```

```
$ consul event -name deploy s3://example.com/deploy-20141117-112233.yml
```

  * `-name`: consul event name (specified by consul watch `-name`)

#### Executing with Serf

Create a serf user event to kick stretcher agents.

```
$ serf event [event_name] [manifest_url]
```

```
$ serf event deploy s3://example.com/deploy-20141117-112233.yml
```

  * event_name: user event name (specified by serf event handler).

#### Executing as command

Stretcher can read a manifest URL from stdin simply.

```
$ echo s3://example.com/deploy-20141117-112233.yml | stretcher
```

You can execute stretcher via ssh or any other methods.


### Deployment process

A stretcher agent executes a following process.

1. Receive a manifest URL as Consul/Serf event's payload.
2. Get a manifest.
3. Get src URL and store it to a temporary file, and Check `checksum`.
4. Invoke `pre` commands.
5. Extract `src` archive to a temporary directory.
6. Sync files from extracted archive to `dest` directory.
  - use `rsync -a --delete` or `mv`
  - sync strategy is switched by `sync_strategy`
7. Invoke `post` commands.
8.
  - Invoke `success` commands when the deployment process succeeded.
  - Invoke `failure` commands when the deployment process failed.

## Manifest spec

### `src`

Source archive URL.

* URL schema: 's3', 'http', 'file'
* Format: 'tar', 'tar.gz'

```yml
src: http://example.com/src/archive.tar.gz
```

### `checksum`

Checksum of source archive.

* Type: 'md5', 'sha1', 'sha256', 'sha512'

```yml
checksum: e0840daaa97cd2cf2175f9e5d133ffb3324a2b93
```

### `dest`

Destination directory.

```yml
dest: /home/stretcher/app
```

### `dest_mode`

Destination directory mode. Default: 0755

```yml
dest_mode: 0711
```

Destination directory mode will be set as...

1. `src` archive includes `.` => same as `.` in the archive.
2. `src` archive does not include `.` => `dest_mode`

### `commands`

* `pre`: Commands which will be invoked at before `src` archive extracted.
* `post`: Commands which will be invoked at after `dest` directory synced.
* `success`: Commands which will be invoked at deployment process is succeeded.
* `failure`: Commands which will be invoked at deployment process is failed.

```yml
commands:
  pre:
    - echo 'staring deploy'
  post:
    - echo 'deploy done'
  success:
    - echo 'deploy success'
  failure:
    - echo 'deploy failed!!'
    - cat >> /path/to/failure.log
```

stretcher agent logs will be passed to STDIN of `success` and `failure` commands.

### `excludes`

Pass to `rsync --exclude` arguments.

```yml
excludes:
  - "*.pid"
  - "*.socket"
```

### `exclude_from`

Pass to `rsync --exclude-from` arguments.
The file must be included in `src` archive.

```yml
exclude_from: exclude.list
```

### `sync_strategy`

A strategy for syncing src extracted directory to dest directory.

- `rsync`: Default
  - Use rsync(1) command with option `-av --delete`
- `mv`
  - Use `os.Rename()` of Golang.
  - Deployment will be failed if `dest` directory is already exists.

## Requirements

* tar
* rsync

tar and rsync must be exist in PATH environment.

If you use stretcher under systemd, You can see unfinished stdout with journald.
You should add `RateLimitBurst=0` into `/etc/systemd/journald.conf` for getting stdout completely.

## Commands execution only mode

If `src` is not defined in a manifest, Stretcher runs `pre`/`post` and `success`/`failure` commands simply.

## LICENSE

The MIT License (MIT)

Copyright (c) 2014 FUJIWARA Shunichiro / (c) 2014 KAYAC Inc.
