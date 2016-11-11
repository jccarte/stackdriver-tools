#!/usr/bin/env bash

set -e

source stackdriver-tools/ci/tasks/utils.sh
source /etc/profile.d/chruby-with-ruby-2.1.2.sh

cpi_release_name="stackdriver-tools"
semver=`cat version-semver/number`

pushd stackdriver-tools
  echo "Using BOSH CLI version..."
  bosh version

  echo "Creating CPI BOSH Release..."
  bosh create release --name ${cpi_release_name} --version ${semver} --with-tarball
popd

image_path=stackdriver-tools/dev_releases/${cpi_release_name}/${cpi_release_name}-${semver}.tgz
echo -n $(sha1sum $image_path | awk '{print $1}') > $image_path.sha1

mv ${image_path} candidate/
mv ${image_path}.sha1 candidate/
