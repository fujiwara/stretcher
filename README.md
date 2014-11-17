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

1. Create tar(or tar.gz) archive for deployment.
2. Upload the archive file to remote server (S3 or HTTP(S)).
3. Create a manifest file (YAML) and upload it to remote server.
  * `src`: archive URL.
  * `checksum`: archive file's checksum. (md5, sha1, sha256 or sha512 hex format)
  * `dest`: destination directory path.
  * `commands`:
    * `pre`: commands to be executed at before the archive extracted.
    * `post`: commands to be executed at after the archive extracted.
4. Create a consul event.
```
$ consul event -name deploy s3://example.com/deploy-20141117-112233.yml
```
  * `-name`: same as stretcher agent event name.
  * payload: manifest URL.
