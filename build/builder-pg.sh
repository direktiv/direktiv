#!/bin/sh

install_postgresql()
{
  if [ "$1" = "true" ] ; then

    mkdir -p /run/postgresql/
    mkdir -p /var/lib/postgresql/data
    /usr/bin/initdb /var/lib/postgresql/data
    echo "host all  all    0.0.0.0/0  md5" >> /var/lib/postgresql/data/pg_hba.conf
    echo "listen_addresses='*'" >> /var/lib/postgresql/data/postgresql.conf
    pg_ctl start -D /var/lib/postgresql/data
    /usr/bin/createdb direktiv
    psql -c "ALTER USER direktiv WITH ENCRYPTED PASSWORD 'direktiv';"

  fi
}

apk update
install_postgresql $1

rm /done
rm /builder-pg.sh
