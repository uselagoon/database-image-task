# database-image-task

This is an image for use in a Lagoon advanced task to create a database container image from a live database.

Currently it only supports mysql/mariadb, but could eventually be extended to support other database systems.

There are some requirements and other things that need to be called out for its usage. Some gotchas as well with variable names that may prevent multiple versions working properly that may need revisting.

## MTK config

The mtk config file should be provided as a base64 encoded string, this allows for easier adding to the api to retain data

```
$ cat example.mtk.yml | base64 | tr -d '\n\r'
cmV3cml0ZToKICAjIERydXBhbCA4CiAgdXNlcnNfZmllbGRfZGF0YToKICAgIG1haWw6IGNvbmNhdCh1aWQsICJAU0FOSVRJU0VEIikKICAgIHBhc3M6ICciU0FOSVRJWkVEX1BBU1NXT1JEIicKICAgICMgV2UgZG9uJ3QgYWxsb3cgbm8gZGlydHkgbmFtZXMgYXJvdW5kIGhlcmUhCiAgICBuYW1lOiBjb25jYXQodWlkLCAiLVNBTklUSVNFRCIpCiAgICBpbml0OiAnIlNBTklUSVNFRF9JTklUIicKICAjIERydXBhbCA3CiAgdXNlcnM6CiAgICBtYWlsOiBjb25jYXQodWlkLCAiQFNBTklUSVNFRCIpCiAgICBwYXNzOiAnIlNBTklUSVpFRF9QQVNTV09SRCInCgp3aGVyZToKICAjIE9ubHkgaW5jbHVkZSBib2R5IGZpZWxkIGRhdGEgZm9yIGN1cnJlbnQgcmV2aXNpb25zLgogIG5vZGVfcmV2aXNpb25fX2JvZHk6IHwtCiAgICAgIHJldmlzaW9uX2lkIElOIChTRUxFQ1QgdmlkIEZST00gbm9kZSkKCm5vZGF0YToKICAtIGNhY2hlKgogIC0gY2FwdGNoYV9zZXNzaW9ucwogIC0gaGlzdG9yeQogIC0gZmxvb2QKICAtIGJhdGNoCiAgLSBxdWV1ZQogIC0gc2Vzc2lvbnMKICAtIHNlbWFwaG9yZQogIC0gc2VhcmNoX2FwaV90YXNrCiAgLSBzZWFyY2hfZGF0YXNldAogIC0gc2VhcmNoX2luZGV4CiAgLSBzZWFyY2hfdG90YWwKICAtIHdhdGNoZG9nCiAgLSB3ZWJmb3JtX3N1Ym1pc3Npb25fZGF0YQoKaWdub3JlOgogIC0gX19BQ1FVSUFfTU9OSVRPUklOR19fCgo=
```

## Variables

An example of the GraphQL needed to create an advanced task is available in 
createDumpSanitisedDB.gql.  The variables which can be passed in are also 
documented in this file (in the "displayName" field of each entry).  

## Process

The process goes through X basic stages:

1. Set up all the initial variables
2. MTK creates a database dump that's basically a sanitised .sql file
3. Make docker-style container with sanitised DB (using podman)
4. Save new container to registry

## Files

### Example mtk.yml file

example.mtk.yml

### Dockerfiles

Dockerfile: Docker file for the builder image

builder/mariadb.Dockerfile: Docker file for the sanitised DB image

builder/final-my.cnf: Configuration file for the above

builder/loading-my.cnf: Configuration file for MySQL while building the sanitised database

### Scripts

image-builder-entry: Launches the next script with appropriate parameters

mariadb-image-builder: Main script that dumps the sanitised database and puts it into a new image

### Example GraphQL queries to use it
createDumpSanitisedDB.gql
createDumpSanitisedDB_noArgs.gql
createDumpSanitisedDB_setDBVariables.gql

# Configuration for renovate bot
renovate.json

