#!/usr/bin/env bash

set -u

force=false
while getopts ":f" option; do
	case "$option" in
	f) force=true ;;
	*)
		printf 'Usage: %s [-f]\n' "$0" >&2
		exit 2
		;;
	esac
done

repo_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
source_dir="$repo_dir/ai_llm/skills"
target_dir="$HOME/.agents/skills"

mkdir -p "$target_dir"

for skill in "$source_dir"/*; do
	[ -e "$skill" ] || [ -L "$skill" ] || continue

	name="${skill##*/}"
	target="$target_dir/$name"

	if [ "$force" = true ]; then
		rm -rf -- "$target"
	elif [ -e "$target" ] || [ -L "$target" ]; then
		continue
	fi

	ln -s "$skill" "$target"
done
