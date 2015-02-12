stretcher
=========

A deployment tool with consul event.

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

A stretcher agent is designed as running under "consul watch" and will be kicked by consul event.

When you specify S3 URL, requires `AWS_CONFIG_FILE` environment variable.

```
$ export AWS_CONFIG_FILE=/path/to/.aws/config
$ consul watch -type event -name deploy /path/to/stretcher
```

* `-name`: your deployment identifiy name.

### Deployment process

#### Preparing

This process is not included in a stretcher agent.

1. Create a tar(or tar.gz) archive for deployment.
2. Upload the archive file to remote server (S3 or HTTP(S)).
3. Create a manifest file (YAML) and upload it to remote server.

#### Executing

Create a consul event to kick stretcher agents.

```
$ consul event -name deploy s3://example.com/deploy-20141117-112233.yml
```
  * `-name`: Same as stretcher agent event name.
  * payload: Manifest URL.

stretcher agent executes a following process.

1. Recieve a consul event from `consul watch`.
2. Get manifest URL.
3. Get src URL and store it to a temporary file, and Check `checksum`.
4. Invoke `pre` commands.
5. Extract `src` archive to a temporary directory.
6. Sync files from extracted archive to `dest` directory
  * using `rsync -a --delete`
7. Invoke `post` commands.

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

Destination directory's mode will be set as...

1. `src` archive includes `.` => same of `.` in the archive.
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

## Requirements

* [Consul](http://consul.io) version 0.4.1 or later.
* tar
* rsync

tar and rsync must be exist in PATH environment.

## LICENSE

The MIT License (MIT)

Copyright (c) 2014 FUJIWARA Shunichiro
