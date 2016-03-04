#!/bin/sh

set -e

. "$(dirname "$0")/helpers.sh"

it_has_installed_mercurial() {
	hg --version
}

run it_has_installed_mercurial
