#!/bin/bash

echo "prep postresql"

mkdir -p /run/postgresql/
mkdir -p /var/lib/postgresql/data

find / | grep createuser

/usr/local/bin/initdb /var/lib/postgresql/data
echo "host all  all    0.0.0.0/0  md5" >> /var/lib/postgresql/data/pg_hba.conf
echo "listen_addresses='*'" >> /var/lib/postgresql/data/postgresql.conf
/usr/local/bin/pg_ctl start -D /var/lib/postgresql/data

/usr/local/bin/createuser -s postgres
/usr/local/bin/psql -l
#/usr/local/bin/psql --help

/usr/local/bin/createdb direktiv
/usr/local/bin/psql -c "ALTER USER postgres WITH ENCRYPTED PASSWORD 'postgres';"
