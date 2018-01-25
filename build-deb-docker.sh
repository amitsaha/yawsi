#!/bin/bash
set -xe

# Check if we have docker running and accessible
# as the current user
# If not bail out with the default error message
docker ps

BUILD_IMAGE='amitsaha/golang-binary-builder'
FPM_IMAGE='amitsaha/golang-deb-builder'
BUILD_ARTIFACTS_DIR="artifacts"

version=`git rev-parse --short HEAD`
VERSION_STRING="$(cat VERSION)-${version}"


# check all the required environment variables are supplied
[ -z "$BINARY_NAME" ] && echo "Need to set BINARY_NAME" && exit 1;
[ -z "$DEB_PACKAGE_NAME" ] && echo "Need to set DEB_PACKAGE_NAME" && exit 1;
[ -z "$DEB_PACKAGE_DESCRIPTION" ] && echo "Need to set DEB_PACKAGE_DESCRIPTION" && exit 1;


docker build --build-arg \
    version_string=$VERSION_STRING \
    --build-arg \
    binary_name=$BINARY_NAME \
    -t $BUILD_IMAGE -f Dockerfile-go1.8 .
containerID=$(docker run --detach $BUILD_IMAGE)
docker cp $containerID:/${BINARY_NAME} .
sleep 1
docker rm $containerID

echo "Binary built. Building DEB now."

docker build --build-arg \
    version_string=$VERSION_STRING \
    --build-arg \
    binary_name=$BINARY_NAME \
    --build-arg \
    deb_package_name=$DEB_PACKAGE_NAME  \
    --build-arg \
    deb_package_description="$DEB_PACKAGE_DESCRIPTION" \
    -t $FPM_IMAGE -f Dockerfile-fpm .
containerID=$(docker run -dt $FPM_IMAGE)
# docker cp does not support wildcard:
# https://github.com/moby/moby/issues/7710
mkdir -p $BUILD_ARTIFACTS_DIR
docker cp $containerID:/deb-package/${DEB_PACKAGE_NAME}-${VERSION_STRING}.deb $BUILD_ARTIFACTS_DIR/.
sleep 1
docker rm -f $containerID
rm $BINARY_NAME
