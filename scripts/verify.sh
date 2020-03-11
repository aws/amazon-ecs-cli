#!/bin/bash
# usage: source these functions and set the s3 endpoint to
# either s3.<region>.amazonaws.com or s3.<region>.amazonaws.com.cn

# usage: retry <command>
function retry {
    local delay=900 # 60 * 15 = 15 minutes
    while true; do
        "$@" && break
        sleep $delay
    done
}

#usage: verifyAll <manifestFile>
function verifyAll {
    for artifact in `cat $1`
    do
        bn="$(basename $artifact)"
        echo "Downloading artifact and verifying signature for $bn"
        wget -q -T 300 https://$S3_ENDPOINT/$BUCKET/$bn
        wget -q -T 300 https://$S3_ENDPOINT/$BUCKET/$bn.asc
        gpg --verify $bn.asc $bn
    done
}

function setupGPG {
    gpg --version
    mkdir ~/.gnupg
    echo "disable-ipv6" >> ~/.gnupg/dirmngr.conf
    gpg --keyserver hkp://keys.gnupg.net --recv BCE9D9A42D51784F
}
