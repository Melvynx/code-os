#!/bin/sh
set -eu

skills_dir=${STACKENV_SKILLS_DIR:-"$HOME/.agents"}
branch=${STACKENV_SKILLS_BRANCH:-main}
device_name=${STACKENV_DEVICE_NAME:-$(hostname -s)}

if ! git -C "$skills_dir" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "StackEnv skills sync: $skills_dir is not a Git repository" >&2
  exit 1
fi

if ! git -C "$skills_dir" remote get-url origin >/dev/null 2>&1; then
  echo "StackEnv skills sync: $skills_dir has no origin remote" >&2
  exit 1
fi

lock_dir="$skills_dir/.git/stackenv-sync.lock"
if ! mkdir "$lock_dir" 2>/dev/null; then
  echo "StackEnv skills sync: another sync is already running"
  exit 0
fi
trap 'rmdir "$lock_dir"' EXIT INT TERM

git -C "$skills_dir" add -A
if ! git -C "$skills_dir" diff --cached --quiet; then
  timestamp=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
  git -C "$skills_dir" commit -m "sync($device_name): $timestamp"
fi

git -C "$skills_dir" pull --rebase --autostash origin "$branch"
git -C "$skills_dir" push origin "$branch"

echo "StackEnv skills sync: $device_name is up to date"
