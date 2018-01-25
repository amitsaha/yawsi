#!/bin/bash
set -xe

BUILD_ARTIFACTS_DIR="artifacts"
version=`git rev-parse --short HEAD`
VERSION_STRING="$(cat VERSION)-${version}"


# check all the required environment variables are supplied
[ -z "$BINARY_NAME" ] && echo "Need to set BINARY_NAME" && exit 1;
[ -z "$DEB_PACKAGE_NAME" ] && echo "Need to set DEB_PACKAGE_NAME" && exit 1;
[ -z "$DEB_PACKAGE_DESCRIPTION" ] && echo "Need to set DEB_PACKAGE_DESCRIPTION" && exit 1;

make build BINARY_NAME=${BINARY_NAME}
echo "Binary built. Building DEB now."

mkdir -p $BUILD_ARTIFACTS_DIR && cp $BINARY_NAME $BUILD_ARTIFACTS_DIR
fpm --output-type deb \
  --input-type dir --chdir /$BUILD_ARTIFACTS_DIR \
  --prefix /usr/bin --name $BINARY_NAME \
  --version $VERSION_STRING \
  --description '${DEB_PACKAGE_DESCRIPTION}' \
  -p ${DEB_PACKAGE_NAME}-${VERSION_STRING}.deb \
  $BINARY_NAME && cp *.deb /$BUILD_ARTIFACTS_DIR/
