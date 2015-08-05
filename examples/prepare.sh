#!/bin/bash

curl -s localhost:8500
if [[ $? != 0 ]]
then
    echo "Please run consul agent on localhost:8500"
    exit 1
fi

set -e

echo "build project..."
cd project
make

echo "create arhchive..."
tar czvf ../example.tar.gz ./

cd ..

echo "create manifest..."
SUM=$(openssl dgst -sha256 example.tar.gz | awk '{print $2}')
CWD=$(pwd)
cat > example.yml <<EOF
src: file://${CWD}/example.tar.gz
checksum: $SUM
dest: ${CWD}/deployed
commands:
  pre:
    - echo "starting deploy" | post-dashboard example 3
    - sleep 3
  post:
    - echo "deploy done"
    - ${CWD}/deployed/hello_world
  success:
    - post-dashboard example 0
  failure:
    - post-dashboard example 2
excludes:
  - "*.go"
  - Makefile
EOF

cat example.yml

echo "create deploy event"
consul event -name "example_deploy" "file://${CWD}/example.yml"

echo
echo "Please run ./exec.sh for deployment."
