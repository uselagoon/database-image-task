#  Create the `.my.cnf` that the lagoon mariadb images use

[client]
user=root
password=Lag00n

[mysql]
database=${BUILDER_BACKUP_IMAGE_DATABASE_NAME:-drupal}

[mysqld]
max_allowed_packet=100M