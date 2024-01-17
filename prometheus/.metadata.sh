# Each line must have an export clause.
# This file is parsed and sourced by the Makefile, Docker and Homebrew builds.
# Powered by Application Builder: https://github.com/golift/application-builder
# Keep in sync with circle-ci job names
declare -A annotate_map=(
    ["x86_64"]="amd64"
    ["armv7l"]="arm"
    ["armv6l"]="arm GOARM=6"
    ["aarch64"]="arm64"
    ["x86"]="386"
)

# Must match the repo name.
BINARY="wattpilot_exporter"
# Github repo containing homebrew formula repo.
HBREPO="mabunixda/wattpilot"
MAINT="Martin Buchleitner"
VENDOR=""
DESC=""
GOLANGCI_LINT_ARGS="--enable-all -D gochecknoglobals -D funlen -e G402 -D gochecknoinits"
# Example must exist at examples/$CONFIG_FILE.example
CONFIG_FILE="up.conf"
LICENSE="MIT"
# FORMULA is either 'service' or 'tool'. Services run as a daemon, tools do not.
# This affects the homebrew formula (launchd) and linux packages (systemd).
FORMULA="service"

OS=$(uname -s | awk '{print tolower($0)}')
U_ARCH=$(uname -m | awk '{print tolower($0)}')

ARCH="${annotate_map[$U_ARCH]}"

export OS ARCH
export BINARY HBREPO MAINT VENDOR DESC GOLANGCI_LINT_ARGS CONFIG_FILE LICENSE FORMULA

# The rest is mostly automatic.
# Fix the repo if it doesn't match the binary name.
# Provide a better URL if one exists.

# Used for source links and wiki links.
SOURCE_URL="https://github.com/${HBREPO}"
# Used for documentation links.
URL="${SOURCE_URL}"

# Dynamic. Recommend not changing.
VVERSION=$(git describe --abbrev=0 --tags $(git rev-list --tags --max-count=1) || echo "v0.0.0")
VERSION="$(echo $VVERSION | tr -d v | grep -E '^\S+$' || echo development)"
# This produces a 0 in some envirnoments (like Homebrew), but it's only used for packages.
ITERATION=$(git rev-list --count --all || echo 0)
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
COMMIT="$(git rev-parse --short HEAD || echo 0)"

GIT_BRANCH="$(git rev-parse --abbrev-ref HEAD || echo unknown)"
BRANCH="${TRAVIS_BRANCH:-${GIT_BRANCH}}"

# This is a custom download path for homebrew formula.
SOURCE_PATH=https://github.com/${HBREPO}/archive/v${VERSION}.tar.gz

export SOURCE_URL URL VVERSION VERSION ITERATION DATE BRANCH COMMIT SOURCE_PATH
