#!/bin/bash

if [[ -z $HELM_DUMP_PLUGIN_DIR ]]; then
  HELM_DUMP_PLUGIN_DIR=$(mktemp -d /tmp/helm-dump-XXXXXX)
fi

DIST_DIR=${DIST_DIR:-./dist}
PLUGIN_BUILD_DIR=${PLUGIN_BUILD_DIR:-${DIST_DIR}/plugin}

echo "Creating ${HELM_DUMP_PLUGIN_DIR}..."
mkdir -p "${HELM_DUMP_PLUGIN_DIR}"

echo "Installing plugin.yaml into ${HELM_DUMP_PLUGIN_DIR}"
install "${PLUGIN_BUILD_DIR}/plugin.yaml" "${HELM_DUMP_PLUGIN_DIR}"

for source_dir in "${DIST_DIR}"/helm-dump_*; do
  plugin_binary_dir=${source_dir#$DIST_DIR/}
  mkdir -p "${HELM_DUMP_PLUGIN_DIR}/${plugin_binary_dir}"
  echo "Adding ${plugin_binary_dir} to plugin bundle..."
  install "${source_dir}"/* "${HELM_DUMP_PLUGIN_DIR}/${plugin_binary_dir}"
done

echo "Done!"