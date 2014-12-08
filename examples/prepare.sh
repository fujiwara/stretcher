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
go build -o hello_world

echo "create arhchive..."
tar czvf ../example.tar.gz ./

cd ..

echo "create manifest..."
SUM=$(sha1sum example.tar.gz | awk '{print $1}')
CWD=$(pwd)
cat > example.yml <<EOF
src: file://${CWD}/example.tar.gz
checksum: $SUM
dest: ${CWD}/deployed
commands:
  pre:
    - echo "starting deploy"
  post:
    - echo "deploy done"
    - ${CWD}/deployed/hello_world
EOF

cat example.yml

echo "create deploy event"
consul event -name "example_deploy" "file://${CWD}/example.yml"

echo "Please run ./exec.sh for deployment."
