#!/usr/bin/env bash
set -euo pipefail

############################################
# Config
############################################
OUT_DIR=/tmp/oci-build
BIN_DIR=$OUT_DIR/bin
IMAGE_DIR=$OUT_DIR/rootfs-oci
LAYER_TAR=$OUT_DIR/rootfs.tar

SKOPEO_URL="https://github.com/117503445/skopeo/releases/download/latest/skopeo-linux-amd64"
UMOCI_URL="https://github.com/opencontainers/umoci/releases/download/v0.6.0/umoci.linux.amd64"

############################################
# Prepare directories
############################################
mkdir -p "$OUT_DIR"
mkdir -p "$BIN_DIR"

echo "[1/6] Downloading umoci and skopeo..."

if [ ! -f "$BIN_DIR/umoci" ]; then
    curl -L "$UMOCI_URL" -o "$BIN_DIR/umoci"
    chmod +x "$BIN_DIR/umoci"
fi

if [ ! -f "$BIN_DIR/skopeo" ]; then
    curl -L "$SKOPEO_URL" -o "$BIN_DIR/skopeo"
    chmod +x "$BIN_DIR/skopeo"
fi

echo "✓ Binaries ready at $BIN_DIR"

############################################
# Detect architecture
############################################
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) OCI_ARCH=amd64 ;;
  aarch64) OCI_ARCH=arm64 ;;
  *) OCI_ARCH="$ARCH" ;;
esac

echo "Detected arch: $OCI_ARCH"

############################################
# Pack rootfs
############################################
echo "[2/6] Packing current rootfs -> $LAYER_TAR"

tar --numeric-owner --xattrs --acls -cpf "$LAYER_TAR" \
  --one-file-system \
  --exclude="$OUT_DIR" \
  --exclude=/proc --exclude=/sys --exclude=/dev --exclude=/run \
  --exclude=/tmp --exclude=/mnt --exclude=/media --exclude=/lost+found \
  --exclude=/var/cache --exclude=/var/tmp \
  -C / .

echo "✓ rootfs tar created"

############################################
# Init OCI layout
############################################
echo "[3/6] Initializing OCI layout"

rm -rf "$IMAGE_DIR"
"$BIN_DIR/umoci" init --layout "$IMAGE_DIR"

############################################
# Create manifest
############################################
echo "[4/6] Creating image manifest"
"$BIN_DIR/umoci" new --image "$IMAGE_DIR:latest"

############################################
# Add layer
############################################
echo "[5/6] Adding rootfs layer"
"$BIN_DIR/umoci" raw add-layer --image "$IMAGE_DIR:latest" "$LAYER_TAR"

############################################
# Configure image
############################################
echo "[6/6] Configuring image"
"$BIN_DIR/umoci" config --image "$IMAGE_DIR:latest" \
  --os linux --arch "$OCI_ARCH" \
  --workingdir / \
  --cmd /bin/sh \
  --label org.opencontainers.image.title="rootfs-snapshot" \
  --created "$(date -u +%Y-%m-%dT%H:%M:%SZ)"

echo
echo "======================================="
echo "✅ OCI image built successfully"
echo "Location: $IMAGE_DIR"
echo
echo "Test unpack:"
echo "  $BIN_DIR/umoci unpack --image $IMAGE_DIR:latest $OUT_DIR/bundle"
echo
echo "Push to registry example:"
echo "  $BIN_DIR/skopeo copy oci:$IMAGE_DIR:latest docker://REGISTRY/NAME:latest"
echo "======================================="

$BIN_DIR/skopeo copy \
  --format v2s2 \
  oci:/tmp/oci-build/rootfs-oci:latest \
  docker://registry.example.com/rootfs:latest