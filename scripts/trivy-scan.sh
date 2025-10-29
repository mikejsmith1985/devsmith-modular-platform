#!/usr/bin/env bash
# scripts/trivy-scan.sh
# Lightweight Trivy integration: scans docker images referenced in docker-compose.yml
# and/or the project filesystem using the official Trivy container image.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

usage() {
  cat <<EOF
Usage: $0 [mode] [target]

Modes:
  image        Scan docker images referenced in docker-compose.yml (default)
  image-manual Scan specific docker image: pass image name as TARGET
  fs           Scan project filesystem (entire repo) for vulnerabilities

Examples:
  $0 image
  $0 image-manual alpine:3.18
  $0 fs

Notes:
- Requires docker to run the trivy container. The script will pull the trivy image if needed.
- Scans are limited to severity HIGH,CRITICAL by default to focus on urgent findings.
EOF
}

MODE="image"
TARGET=""
if [[ ${#@} -ge 1 ]]; then
  MODE="$1"
  TARGET="${2-}" 
fi

TRIVY_IMAGE="aquasec/trivy:latest"
SEVERITY="HIGH,CRITICAL"
FORMAT_TABLE="--format table"

run_trivy() {
  docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v "$ROOT_DIR":/project -w /project "$TRIVY_IMAGE" "$@"
}

scan_images_from_compose() {
  # Extract image names from docker-compose.yml (services.image fields)
  local compose_file="$ROOT_DIR/docker-compose.yml"
  if [[ ! -f "$compose_file" ]]; then
    echo "docker-compose.yml not found at $compose_file" >&2
    return 1
  fi

  # Use yq if available for robust parsing, otherwise fall back to grep parsing
  local images=()
  if command -v yq >/dev/null 2>&1; then
    mapfile -t images < <(yq e '.services[] | .image // ""' "$compose_file" | sed '/^$/d' | sort -u)
  else
    # Grep for 'image:' lines under services
    mapfile -t images < <(awk '/services:/,/^[[:space:]]*[^:]+:/{ if ($1=="image:") print $2 }' "$compose_file" | sed 's/"//g' | sed '/^$/d' | sort -u)
  fi

  if [[ ${#images[@]} -eq 0 ]]; then
    echo "No images found in docker-compose.yml" >&2
    return 1
  fi

  echo "Found images to scan:" 
  for img in "${images[@]}"; do
    echo "  - $img"
  done

  for img in "${images[@]}"; do
    echo -e "\nScanning image: $img (severity: $SEVERITY)\n----------------------------------------"
    # Pull the image (if not present) to ensure trivy can scan it
    echo "Pulling $img (may be skipped if already present)..."
    docker pull "$img" || echo "Warning: docker pull failed for $img; trivy may still try to scan local image name"

    run_trivy image --severity "$SEVERITY" $FORMAT_TABLE "$img" || echo "Trivy scan returned non-zero for $img"
  done
}

scan_single_image() {
  if [[ -z "$TARGET" ]]; then
    echo "Target image required for image-manual mode" >&2
    usage
    exit 2
  fi
  echo "Scanning image: $TARGET (severity: $SEVERITY)"
  docker pull "$TARGET" || echo "Warning: pull failed for $TARGET"
  run_trivy image --severity "$SEVERITY" $FORMAT_TABLE "$TARGET" || echo "Trivy scan returned non-zero"
}

scan_fs() {
  echo "Scanning filesystem (project root) for vulnerabilities (severity: $SEVERITY)"
  run_trivy fs --severity "$SEVERITY" $FORMAT_TABLE /project || echo "Trivy filesystem scan returned non-zero"
}

case "$MODE" in
  image)
    scan_images_from_compose
    ;;
  image-manual)
    scan_single_image
    ;;
  fs)
    scan_fs
    ;;
  -h|--help)
    usage
    ;;
  *)
    echo "Unknown mode: $MODE" >&2
    usage
    exit 2
    ;;
esac

exit 0
