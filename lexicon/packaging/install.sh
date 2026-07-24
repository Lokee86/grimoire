#!/usr/bin/env sh
set -eu

source_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
install_dir=${1:-${LEXICON_INSTALL_DIR:-${XDG_DATA_HOME:-"$HOME/.local/share"}/lexicon}}
bin_dir=${LEXICON_BIN_DIR:-"$HOME/.local/bin"}

if [ ! -x "$source_dir/lexicon" ]; then
    echo "lexicon was not found beside install.sh. Run this script from an extracted Lexicon release package." >&2
    exit 1
fi
if [ ! -d "$source_dir/adapters" ]; then
    echo "The adapters directory was not found beside install.sh. The release package is incomplete." >&2
    exit 1
fi

mkdir -p "$(dirname -- "$install_dir")"
install_parent=$(CDPATH= cd -- "$(dirname -- "$install_dir")" && pwd)
install_dir="$install_parent/$(basename -- "$install_dir")"

if [ "$source_dir" != "$install_dir" ]; then
    case "$install_dir/" in
        "$source_dir"/*)
            echo "Install directory cannot be inside the extracted release package." >&2
            exit 1
            ;;
    esac
    mkdir -p "$install_dir"
    cp -R "$source_dir"/. "$install_dir"/
fi
chmod +x "$install_dir/lexicon"

mkdir -p "$bin_dir"
ln -sf "$install_dir/lexicon" "$bin_dir/lexicon"

echo "Lexicon installed to $install_dir"
echo "Command link created at $bin_dir/lexicon"
case ":${PATH:-}:" in
    *":$bin_dir:"*) ;;
    *) echo "Add $bin_dir to PATH before running lexicon." ;;
esac
