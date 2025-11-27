#!/bin/bash

/usr/local/bin/docker-entrypoint.sh mysqld \
    --max-allowed-packet=1G \
    --datadir /initialized-db \
    --aria-log-dir-path /initialized-db > /tmp/output.log 2>&1

if [ "$?" != "0" ]; then
    # print the last 3 lines of the log that shows the error
    tail -n 3 /tmp/output.log
    exit 1
fi
