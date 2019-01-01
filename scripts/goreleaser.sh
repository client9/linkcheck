#!/bin/sh -e
# autorelease based on tag
export GO111MODULE=on
if test -z "$TRAVIS_TAG"; then
	echo "no tag found, not goreleasing"
	exit 0
fi
echo "found tag ${TRAVIS_TAG}"
curl -sL https://git.io/goreleaser | bash
