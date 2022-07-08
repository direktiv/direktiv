#!/bin/bash

for arg; do
    MD5=`echo $arg | /usr/bin/md5sum | /bin/cut -f1 -d" "`
    echo "creating tar for $arg at $MD5"
    rm -Rf $MD5 && mkdir $MD5
    
    # cut of sha if provided
    NAME=`echo $arg | cut -d @ -f 1` 
    skopeo copy docker://$arg docker-archive:$MD5/$MD5.tar:$NAME
    ls -la $MD5/$MD5.tar
    cp $MD5/$MD5.tar /images
done
