#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# from https://cloud.google.com/sdk/docs/downloads-apt-get
export CLOUD_SDK_REPO="cloud-sdk-$(lsb_release -c -s)"
echo "deb http://packages.cloud.google.com/apt $CLOUD_SDK_REPO main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
sudo apt-get update && sudo apt-get install google-cloud-sdk

gcloud config set disable_prompts True
gcloud auth activate-service-account --key-file <(echo ${GCLOUD_CLIENT_SECRET} | base64 --decode)
gcloud auth configure-docker

readonly root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." >/dev/null 2>&1 && pwd)
readonly version=$(cat ${root}/VERSION)
readonly git_branch=${GITHUB_REF:11} # drop 'refs/head/' prefix
readonly git_timestamp=$(TZ=UTC git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")
readonly slug=${version}-${git_timestamp}-${GITHUB_SHA:0:16}

echo "Building riff kafka provider"
(cd $root && KO_DOCKER_REPO="gcr.io/projectriff/kafka-provider" ko resolve -P -t "${version}" -t "${slug}" -f config/ | \
  sed -e "s|projectriff.io/release: devel|projectriff.io/release: \"${version}\"|" > ${root}/riff-kafka-provider.yaml)

echo "Publishing riff kafka provider"
gsutil cp -a public-read ${root}/riff-kafka-provider.yaml gs://projectriff/kafka-provider/snapshots/riff-kafka-provider-${slug}.yaml
gsutil cp -a public-read ${root}/riff-kafka-provider.yaml gs://projectriff/kafka-provider/riff-kafka-provider-${version}.yaml

echo "Publishing version references"
gsutil -h 'Content-Type: text/plain' -h 'Cache-Control: private' cp -a public-read <(echo "${slug}") gs://projectriff/kafka-provider/snapshots/versions/${git_branch}
gsutil -h 'Content-Type: text/plain' -h 'Cache-Control: private' cp -a public-read <(echo "${slug}") gs://projectriff/kafka-provider/snapshots/versions/${version}
if [[ ${version} != *"-snapshot" ]] ; then
  gsutil -h 'Content-Type: text/plain' -h 'Cache-Control: private' cp -a public-read <(echo "${version}") gs://projectriff/kafka-provider/versions/releases/${git_branch}
fi
