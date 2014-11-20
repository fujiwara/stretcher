stretcher
=========

A deployment tool with consul event.

## example manifest

```yml
src: s3://example.com/app.tar.gz
checksum: e0840daaa97cd2cf2175f9e5d133ffb3324a2b93
dest: /home/stretcher/app
commands:
  pre:
    - echo 'staring deploy'
  post:
    - echo 'deploy done'
excludes:
  - "*.pid"
  - "*.socket"
```

## run

### stretcher agent

A stretcher agent is designed as running under "consul watch" and will be kicked by consul event.

When you use S3 URL, export `AWS_CONFIG_FILE` environment variable.

```
$ export AWS_CONFIG_FILE=/path/to/.aws/config
$ consul watch -type event -name deploy /path/to/stretcher
```

* `-name`: your deployment identifiy name.

### deployment process

#### preparing

1. Create tar(or tar.gz) archive for deployment.
2. Upload the archive file to remote server (S3 or HTTP(S)).
3. Create a manifest file (YAML) and upload it to remote server.
  * `src`: archive URL.
  * `checksum`: archive file's checksum. (md5, sha1, sha256 or sha512 hex format)
  * `dest`: destination directory path.
  * `commands`:
    * `pre`: commands to be executed at before the archive extracted.
    * `post`: commands to be executed at after the archive extracted.

#### executing

1. Create a consul event.
```
$ consul event -name deploy s3://example.com/deploy-20141117-112233.yml
```
  * `-name`: same as stretcher agent event name.
  * payload: manifest URL.

2. (agent)
  1. Fetching event from `consul watch`.
  2. Get manifest URL.
  3. Get src URL and store to tmporary directory (by `os.TempDir()`) and Check `checksum`.
  4. Invoke `pre` commands.
  5. Extract src archive to temporary directory.
  6. Sync files from extracted archive to `dest` directory
    * by `rsync -a --delete`
  7. Invoke `post` commands.


## Manifest spec

### `src`

* Source archive URL.
* URL schema: 's3', 'http', 'file'
* Format: 'tar', 'tar.gz'

```yml
src: http://example.com/src/archive.tar.gz
```

### `checksum`

* Checksum of source archive.
* Type: 'md5', 'sha1', 'sha256', 'sha512'

```yml
checksum: e0840daaa97cd2cf2175f9e5d133ffb3324a2b93
```

### `dest`

* Destination directory.

```yml
dest: /home/stretcher/app
```

### `commands`

* `pre`: Commands which will be invoked at before `src` archive extracted.
* `post`: Commands which will be invoked at after `dest` directory synced.

```yml
commands:
  pre:
    - echo 'staring deploy'
  post:
    - echo 'deploy done'
```

### `excludes`

* Pass to `rsync --exclude` arguments.

```yml
excludes:
  - "*.pid"
  - "*.socket"
```

### `exclude_from`

* Pass to `rsync --exclude-from` arguments.
* The file must be included in `src` archive.

```yml
excludes_from: exclude.list
```
