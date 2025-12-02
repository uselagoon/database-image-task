#!/bin/bash

/usr/local/bin/docker-entrypoint.sh mysqld \
    --innodb-buffer-pool-size=2G \
    --innodb-sort-buffer-size=128M \
    --bulk-insert-buffer-size=256M \
    --innodb-buffer-pool-instances=4 \
    --innodb-read-io-threads=4 \
    --innodb-write-io-threads=4 \
    --max-allowed-packet=1G \
    --datadir /initialized-db \
    --aria-log-dir-path /initialized-db > /tmp/output.log 2>&1

if [ "$?" != "0" ]; then
    # print the last 3 lines of the log that shows the error
    tail -n 3 /tmp/output.log
    exit 1
fi
