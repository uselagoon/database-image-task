# https://serverfault.com/a/915845

ARG BUILDER_IMAGE
ARG CLEAN_IMAGE
FROM ${BUILDER_IMAGE} AS builder

# That file does the DB initialization but also runs mysql daemon, by removing the last line it will only init
RUN ["sed", "-i", "s/exec \"$@\"/echo \"not running $@\"/", "/usr/local/bin/docker-entrypoint.sh"]

# set the lagoon mysql defaults here
ENV MYSQL_ROOT_PASSWORD=Lag00n
ENV MYSQL_DATABASE=lagoon \
    MYSQL_USER=lagoon \
    MYSQL_PASSWORD=lagoon

COPY sanitised-dump.sql /docker-entrypoint-initdb.d/

# Need to change the datadir to something else that /var/lib/mysql because the parent docker file defines it as a volume.
# https://docs.docker.com/engine/reference/builder/#volume :
#       Changing the volume from within the Dockerfile: If any build steps change the data within the volume after
#       it has been declared, those changes will be discarded.
# capture only the last 3 lines of this output to help with debugging. capturing more than this has potential to leak data
# in the output
RUN /usr/local/bin/docker-entrypoint.sh mysqld \
    --max-allowed-packet=1G \
    --datadir /initialized-db | tail -n 3

#  create the `.my.cnf` that the lagoon mariadb images use
# apply the permissions in the builder image before transferring to the clean image
# this brings the `.my.cnf` file with it so that the clean image will start correctly
COPY my.cnf /initialized-db/.my.cnf
RUN chown -R 999:root /initialized-db
COPY import.my.cnf /etc/mysql/my.cnf
RUN chown -R 999:root /etc/mysql/my.cnf

FROM ${CLEAN_IMAGE}

COPY --from=builder /initialized-db /var/lib/mysql

RUN cp /var/lib/mysql/.my.cnf /etc/mysql/my.cnf

