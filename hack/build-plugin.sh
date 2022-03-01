#!/bin/bash

DIST_DIR=${DIST_DIR:-./dist}
PLUGIN_BUILD_DIR=${PLUGIN_BUILD_DIR:-${DIST_DIR}/plugin}

echo "Building plugin in ${PLUGIN_BUILD_DIR}... "
mkdir -p "${PLUGIN_BUILD_DIR}"
cp -R "${DIST_DIR}"/helm-dump_*/ "${PLUGIN_BUILD_DIR}"
cp ./plugin.yaml "${PLUGIN_BUILD_DIR}"
yq ".version |= $(yq .version dist/metadata.json)" ./plugin.yaml > "${PLUGIN_BUILD_DIR}/plugin.yaml"
echo "Done!"