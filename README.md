# database-image-task

This is an image for use in a Lagoon advanced task to create a database 
container image from a live database, possibly sanitising the data on the way.  

Currently it only supports mysql/mariadb, but could eventually be extended to support other database systems.
It was originally designed for mariadb, hence the naming of some of the files in this project.  

There are some requirements and other things that need to be called out for its 
usage. Some gotchas as well with variable names that may prevent multiple 
versions working properly that may need revisting.

## MTK config

The mtk config file should be provided as a base64 encoded string, this allows for easier adding to the api to retain data

```
$ cat example.mtk.yml | base64 | tr -d '\n\r'
cmV3cml0ZToKICAjIERydXBhbCA4CiAgdXNlcnNfZmllbGRfZGF0YToKICAgIG1haWw6IGNvbmNhdCh1aWQsICJAU0FOSVRJU0VEIikKICAgIHBhc3M6ICciU0FOSVRJWkVEX1BBU1NXT1JEIicKICAgICMgV2UgZG9uJ3QgYWxsb3cgbm8gZGlydHkgbmFtZXMgYXJvdW5kIGhlcmUhCiAgICBuYW1lOiBjb25jYXQodWlkLCAiLVNBTklUSVNFRCIpCiAgICBpbml0OiAnIlNBTklUSVNFRF9JTklUIicKICAjIERydXBhbCA3CiAgdXNlcnM6CiAgICBtYWlsOiBjb25jYXQodWlkLCAiQFNBTklUSVNFRCIpCiAgICBwYXNzOiAnIlNBTklUSVpFRF9QQVNTV09SRCInCgp3aGVyZToKICAjIE9ubHkgaW5jbHVkZSBib2R5IGZpZWxkIGRhdGEgZm9yIGN1cnJlbnQgcmV2aXNpb25zLgogIG5vZGVfcmV2aXNpb25fX2JvZHk6IHwtCiAgICAgIHJldmlzaW9uX2lkIElOIChTRUxFQ1QgdmlkIEZST00gbm9kZSkKCm5vZGF0YToKICAtIGNhY2hlKgogIC0gY2FwdGNoYV9zZXNzaW9ucwogIC0gaGlzdG9yeQogIC0gZmxvb2QKICAtIGJhdGNoCiAgLSBxdWV1ZQogIC0gc2Vzc2lvbnMKICAtIHNlbWFwaG9yZQogIC0gc2VhcmNoX2FwaV90YXNrCiAgLSBzZWFyY2hfZGF0YXNldAogIC0gc2VhcmNoX2luZGV4CiAgLSBzZWFyY2hfdG90YWwKICAtIHdhdGNoZG9nCiAgLSB3ZWJmb3JtX3N1Ym1pc3Npb25fZGF0YQoKaWdub3JlOgogIC0gX19BQ1FVSUFfTU9OSVRPUklOR19fCgo=
```

An example can be found in `example.mtk.yml`

## Variables

An example of the GraphQL needed to create an advanced task is available in 
createDumpSanitisedDB.gql.  The variables which can be passed in are also 
documented in this file (in the "displayName" field of each entry).  

The GraphQL examples are:

* **createDumpSanitisedDB.gql**: Allows the person running the task to change all the (normal) variables
* **createDumpSanitisedDB_noArgs.gql**: Variables must be set on the environment; the person running the task has no control
* **createDumpSanitisedDB_setDBVariables.gql**: Allows the person running the task to additionally choose the database to which they connect

Most of the variables are explained in the example GraphQL files, but one in 
particular requires a better writeup.

### BUILDER_BACKUP_IMAGE_NAME

This is the name of the resulting image to build, without the tag (eg, for 
dockerhub myproject/image, or for custom registry quay.io/myproject/image, can 
also be a lpattern like 
`${registry}/${organization}/${project}/${environment}/${service}-data`)",

The default value is `${registry}/${organization}/${project}/${environment}/${service}-data`

The variables available are as follows:
* `${registry}`: `BUILDER_REGISTRY_HOST` (variable defined in Lagoon organisation/project/environment)
* `${organization}`: `BUILDER_REGISTRY_ORGANIZATION` (variable defined in Lagoon organisation/project/environment)
* `${project}`: The name of the Lagoon project
* `${environment}`: The name of the Lagoon environment
* `${service}`: The name of the Lagoon service in the docker-compose file
* `${database}`: The name of the database.  `${database}` is munged so that:
    1. Any special character not allowed in DockerHub repo names is removed (replaced with nothing), and
    2. If there are two special characters in a row, the first is retained, and later ones are removed (also as per DockerHub repo name requirements)

## The Images

There are functionally three images we have to worry about:
1. The Sanitiser Image: This is the image that dumps the sanitised database and builds the other images
2. The Sanitised Builder Image: This builds the sanitised image
3. The Sanitised Clean Image: This is the clean MariaDB image that the sanitised database is put in

## Building the Sanitiser Image

The image is built using the `Dockerfile`.  Some parts are written in Go.  

### The Go Parts

This uses the `Makefile` to build the parts of the image that are written in Go.  
This is the variable calculation part of the process, and was rewritten in Go 
so that tests could be applied.  Other files associated with this part of the 
process are:
* `cmd/main.go`
* `go.mod`
* `go.sum`
* `main.go`
* `internal/builder/builder.go`
* `internal/builder/builder_test.go`: Tests for `internal/builder/builder.go`
* `internal/builder/variables.go`
* `internal/builder/variables_test.go`: Tests for `internal/builder/variables.go`

## The Sanitiser Image in Use

The entry point is `image-builder-entry`.  This is just a wrapper around 
`mariadb-image-builder`.  

### Overall Process

The process goes through the following basic stages:

1. Set up all the initial variables
2. MTK creates a database dump that's basically a sanitised .sql file
3. Make docker-style container with sanitised DB (using podman); this uses a builder image, and copies the results into a clean image
4. Save new container to registry

### Files for the Sanitised Builder Process

The files used in this live in the `builder` directory (and you could arguably 
include the files in the internal/builder directory, which I've listed under 
"The Go Parts", above).  

These are:
* `builder/mariadb.Dockerfile`: The dockerfile that's the script for both the builder and clean images mentioned in step 3, above
* `builder/import.my.cnf.tpl`: The my.cnf used in the builder image

### Files forr the Sanitised Clean Image

* `builder/my.cnf.tpl`: The my.cnf used in the final sanitised image

## Renovate

`renovate.json` allows for the easy upgrading of various component pieces of 
software, especially mtk itself.
