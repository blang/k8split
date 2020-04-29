# K8split

This tools splits a single file of concatenated kubernetes manifests into a temporary directory structure.
It's useful for diff'ing large deployments or helm charts.

## Build
```
go build ./cmd/k8split
```

## Usage
```
# Use with files
dir=$(k8split largemanifest.yml)
ls $dir

# Use with Stdin
dir=$(cat largemanifest.yml | k8split -)
ls $dir

# Define tmp dir suffix for readability in diff
export K8SPLIT_DIR_SUFFIX="current-release"
dir=$(cat largemanifest.yml |  k8split -)
ls $dir

# Export to specific directory
mkdir ./output
export K8SPLIT_TARGET_DIR="./output"
cat largemanifest.yml |  k8split -
```
