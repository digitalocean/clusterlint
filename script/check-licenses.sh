#!/bin/bash

function list_go_files {
    srcdir=$(realpath $(dirname "${BASH_SOURCE[0]}")/..)
    for f in $(find "$srcdir" -name '*.go' -and -not -path "$srcdir/vendor/*") ; do
        echo "$f"
    done
}

function has_license {
    head -n2 "$1" | grep -q 'Copyright .... DigitalOcean'
}

ret=0
for f in $(list_go_files) ; do
    if ! has_license "$f" ; then
        echo "$f is missing license header"
        ret=1
    fi
done

exit $ret
