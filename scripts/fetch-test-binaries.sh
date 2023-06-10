#!/usr/bin/env bash

set -e

k8s_version=1.14.1
arch=amd64

if [[ "$OSTYPE" == "linux-gnu" ]]; then
  os="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
  os="darwin"
else
  echo "OS '$OSTYPE' not supported." >&2
  exit 1
fi

root=$(cd "`dirname $0`"/..; pwd)
output_dir="$root"/_out
archive_name="kubebuilder-tools-$k8s_version-$os-$arch.tar.gz"
archive_file="$output_dir/$archive_name"
archive_url="https://storage.googleapis.com/kubebuilder-tools/$archive_name"

bin_dir="${output_dir}/kubebuilder/bin"
mkdir -p "$output_dir"

if [[ -f $bin_dir/etcd && -f $bin_dir/kube-apiserver && -f $bin_dir/kubectl && -f $bin_dir/kind ]]; then
  echo "binaries: skipping as already downloaded"
  exit 0
fi
echo "downloading binaries"
curl -sL "$archive_url" -o "$archive_file"
tar -zxf "$archive_file" -C "$output_dir/"

curl -sLo ./kind "https://kind.sigs.k8s.io/dl/v0.19.0/kind-${os}-amd64"
chmod +x ./kind
mv ./kind "$bin_dir"
