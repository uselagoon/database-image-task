[mysqld]
max_allowed_packet=1G
bulk_insert_buffer_size = 256M
innodb_buffer_pool_chunk_size = 128M
innodb_buffer_pool_size = 128M
innodb_buffer_pool_instances = 4
innodb_read_io_threads = 4
innodb_write_io_threads = 4