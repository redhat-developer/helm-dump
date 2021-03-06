#!/bin/bash
# shellcheck disable=SC2120

pushd () {
    command pushd "$@" > /dev/null
}

popd () {
    command popd "$@" > /dev/null
}

export pushd popd

DIST_DIR=${DIST_DIR:-$PWD/dist}
PLUGIN_ALIAS=dump
PLUGIN_BUILD_WORKDIR=${PLUGIN_BUILD_WORKDIR:-${DIST_DIR}/plugin}
PLUGIN_SOURCE_DIR=${PLUGIN_SOURCE_DIR:-${PLUGIN_BUILD_WORKDIR}/${PLUGIN_ALIAS}}
PLUGIN_OUTPUT_DIR=${PLUGIN_OUTPUT_DIR:-${PLUGIN_BUILD_WORKDIR}/artifacts}

mkdir -p "$PLUGIN_OUTPUT_DIR"

PLUGIN_VERSION=$(yq .version "${DIST_DIR}"/metadata.json | xargs)
PLUGIN_BUNDLE_NAME="helm-dump_${PLUGIN_VERSION}"

echo -n "Building plugin in ${PLUGIN_SOURCE_DIR}... "
mkdir -p "${PLUGIN_SOURCE_DIR}"
cp -R "${DIST_DIR}"/helm-plugin_*/* "${DIST_DIR}"/crane-plugin_*/* "${PLUGIN_SOURCE_DIR}"
cp ./plugin.yaml "${PLUGIN_SOURCE_DIR}"
yq ".version |= \"$PLUGIN_VERSION\"" ./plugin.yaml > "${PLUGIN_SOURCE_DIR}/plugin.yaml"
echo "Done!"

echo -n "Creating ${PLUGIN_BUNDLE_NAME}.tar.gz... "
tar -czf "${PLUGIN_OUTPUT_DIR}/${PLUGIN_BUNDLE_NAME}.tar.gz" -C "${PLUGIN_BUILD_WORKDIR}" "${PLUGIN_ALIAS}"
echo "Done!"

echo -n "Creating ${PLUGIN_BUNDLE_NAME}.zip... "
(cd "${PLUGIN_BUILD_WORKDIR}" && zip -q -r "${PLUGIN_OUTPUT_DIR}/${PLUGIN_BUNDLE_NAME}.zip" "${PLUGIN_ALIAS}")
echo "Done!"

pushd "${PLUGIN_OUTPUT_DIR}" || exit
echo -n "Calculating checksum for plugin bundles... "
CHECKSUM_FILE="${PLUGIN_BUNDLE_NAME}.checksums.txt"
shasum -a 256 "${PLUGIN_BUNDLE_NAME}".tar.gz >> "${CHECKSUM_FILE}"
shasum -a 256 "${PLUGIN_BUNDLE_NAME}".zip >> "${CHECKSUM_FILE}"
echo "Done!"
popd || exit