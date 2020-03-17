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

function failBadPGPSignature {
    echo "PGP Signature failed. retrying in $delay s" >&2
    return 1
}

function failBadConnection {
    echo "Couldn't download artifact $bn. Retrying in $delay s" >&2
    return 1
}

#usage: verifyAll <manifestFile>
function verifyAll {
    for artifact in `cat $1`
    do
        bn="$(basename $artifact)"
        echo "Downloading artifact $bn from endpoint https://$S3_ENDPOINT/$BUCKET"
        wget -q https://$S3_ENDPOINT/$BUCKET/$bn || failBadConnection || return 1
        echo "Downloading artifact signature $bn.asc from $S3_ENDPOINT_SIGNATURES"
        wget -q -T 300 https://$S3_ENDPOINT_SIGNATURES/$BUCKET/$bn.asc
        echo "verifying signature..."
        gpg --verify $bn.asc $bn || failBadPGPSignature || return 1
    done
}

function setupGPG {
    gpg --version
    mkdir ~/.gnupg
    echo "disable-ipv6" >> ~/.gnupg/dirmngr.conf
    gpg --keyserver hkp://pool.sks-keyservers.net --recv BCE9D9A42D51784F || exit 1
}

setupGPG
retry verifyAll $1
