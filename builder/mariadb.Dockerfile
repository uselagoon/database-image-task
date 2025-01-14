# https://serverfault.com/a/915845

ARG BUILDER_IMAGE
FROM ${BUILDER_IMAGE:-mariadb:10.6} as builder

# The entrypoint file does the DB initialization but also runs mysql daemon.  By removing the last line it will only init but not start
RUN ["sed", "-i", "s/exec \"$@\"/echo \"not running $@\"/", "/usr/local/bin/docker-entrypoint.sh"]

# set the lagoon mariadb-drupal defaults here
ENV MYSQL_ROOT_PASSWORD=Lag00n
ENV MARIADB_DATABASE=drupal \
    MARIADB_USER=drupal \
    MARIADB_PASSWORD=drupal

COPY sanitised-dump.sql /docker-entrypoint-initdb.d/

#  create the `.my.cnf` that the lagoon mariadb images use
COPY final-my.cnf /initialized-db/.my.cnf
RUN chown -R 100:root /initialized-db
COPY loading-my.cnf /etc/mysql/my.cnf
RUN chown -R 100:root /etc/mysql/my.cnf

RUN head -20 /docker-entrypoint-initdb.d/sanitised-dump.sql

# Need to change the datadir to something else that /var/lib/mysql because the parent docker file defines it as a volume.
# https://docs.docker.com/engine/reference/builder/#volume :
#       Changing the volume from within the Dockerfile: If any build steps change the data within the volume after
#       it has been declared, those changes will be discarded.
RUN ["/usr/local/bin/docker-entrypoint.sh", "mysqld", "--datadir", "/initialized-db", "--aria-log-dir-path", "/initialized-db"]

# apply the permissions in the builder image before transferring to the clean image
# this brings the `.my.cnf` file with it so that the clean image will start correctly
RUN chown -R 100:root /initialized-db

ARG CLEAN_IMAGE
FROM ${CLEAN_IMAGE:-uselagoon/mariadb-10.6-drupal:latest}

COPY --from=builder /initialized-db /var/lib/mysql

RUN cp /var/lib/mysql/.my.cnf /etc/mysql/my.cnf

