#!/bin/sh

set -e

export TMPDIR_ROOT=$(mktemp -d /tmp/hg-tests.XXXXXX)

$(dirname $0)/image.sh

$(dirname $0)/test_check.sh

$(dirname $0)/test_in.sh

# $(dirname $0)/put.sh

echo -e '\e[32mall tests passed!\e[0m'

rm -rf $TMPDIR_ROOT
