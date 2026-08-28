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

git_dir=$(git -C "$skills_dir" rev-parse --absolute-git-dir)
lock_dir="$git_dir/stackenv-sync.lock"
if ! mkdir "$lock_dir" 2>/dev/null; then
  echo "StackEnv skills sync: another sync is already running"
  exit 0
fi
trap 'rmdir "$lock_dir"' EXIT INT TERM

if [ -d "$git_dir/rebase-merge" ] || [ -d "$git_dir/rebase-apply" ] || [ -f "$git_dir/MERGE_HEAD" ] || ! git -C "$skills_dir" diff --quiet --diff-filter=U; then
  echo "StackEnv skills sync: resolve the existing Git conflict before syncing" >&2
  exit 1
fi

current_branch=$(git -C "$skills_dir" symbolic-ref --quiet --short HEAD || true)
if [ "$current_branch" != "$branch" ]; then
  echo "StackEnv skills sync: expected branch $branch, found ${current_branch:-detached HEAD}" >&2
  exit 1
fi

git -C "$skills_dir" add -A
if ! git -C "$skills_dir" diff --cached --quiet; then
  timestamp=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
  git -C "$skills_dir" commit -m "sync($device_name): $timestamp"
fi

git -C "$skills_dir" pull --rebase --autostash origin "$branch"
git -C "$skills_dir" push origin "$branch"

echo "StackEnv skills sync: $device_name is up to date"
