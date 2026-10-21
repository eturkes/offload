# shellcheck shell=bash
# remote run (after watch.bash): lock the workdir, prune it (git mode), run CMD under
# offload_watch, list its outputs, record the status.
# Args: WORKDIR META SUBDIR MODE NONCE COUNT EXCLUDES CMD — paths relative to $HOME; MODE = git |
# plain | nosync. Git mode reads NONCE + COUNT paths (NUL-terminated) from stdin first; the rest of
# stdin is offload's pipe (EOF = offload gone).
set -u
fail() {
  printf 'offload: %s: %s\n' "$(uname -n)" "$*" >&2
  exit 125
}
# in_tree PATH: every ancestor of PATH below the workdir is a real directory (no symlink), so PATH
# names something inside the workdir, never a target a symlink points at.
in_tree() {
  local f=$1 d=. rest comp
  case "/$f/" in //* | */../* | */./*) return 1 ;; esac
  rest=${f%/*}
  [ "$rest" != "$f" ] || rest=
  while [ -n "$rest" ]; do
    comp=${rest%%/*}
    d=$d/$comp
    if [ -L "$d" ] || [ ! -d "$d" ]; then return 1; fi
    if [ "$rest" = "$comp" ]; then rest=; else rest=${rest#*/}; fi
  done
}
# Prune = delete a file or symlink inside the workdir; never a directory, never recursive.
prune_one() {
  if in_tree "$1" && { [ -L "$1" ] || [ -f "$1" ]; }; then rm -f -- "$1"; fi
}
git_list() { # non-ignored files of the workdir, as git sees them
  git --git-dir=.git --work-tree=. -c core.excludesFile="$meta/excludes" \
    ls-files -z -co --exclude-standard
}

cd || fail "no home directory"
wd=$HOME/$1 meta=$HOME/$2 sub=$3 mode=$4 nonce=$5 count=$6 excludes=$7 cmd=$8
mkdir -p "$wd" "$meta" || fail "cannot create workdir $1"
exec 9>"$meta/lock"
flock -w 300 9 || fail "workdir $1 still busy after 300 s"
cd "$wd" || fail "cannot enter workdir $1"
rm -f "$meta/status" "$meta/down"

if [ "$mode" = git ]; then
  [ -d .git ] || fail "workdir $1 has no .git (upload incomplete)"
  head -z -n "$((count + 1))" >"$meta/upload"
  n=$(tr -cd '\0' <"$meta/upload" | wc -c)
  first=$(head -z -n 1 "$meta/upload" | tr -d '\0')
  if [ "$n" != "$((count + 1))" ] || [ "$first" != "$nonce" ]; then
    fail "upload list framing mismatch ($n of $((count + 1)) records); nothing pruned"
  fi
  printf '%s' "$excludes" >"$meta/excludes"
  git_list >"$meta/remote" || fail "git ls-files failed in workdir $1; nothing pruned"
  tail -z -n +2 "$meta/upload" | LC_ALL=C sort -z >"$meta/upload.sorted"
  LC_ALL=C sort -z "$meta/remote" | LC_ALL=C comm -z -23 - "$meta/upload.sorted" >"$meta/prune"
  while IFS= read -r -d '' f; do prune_one "$f"; done <"$meta/prune"
fi

cd "./$sub" || fail "workdir $1 has no directory $sub"
offload_watch bash -lc "$cmd"
rc=$?

if [ "$mode" = git ] && cd "$wd"; then
  git_list | while IFS= read -r -d '' f; do
    if in_tree "$f" && { [ -e "$f" ] || [ -L "$f" ]; }; then printf '%s\0' "$f"; fi
  done >"$meta/down"
fi
printf '%s %s\n' "$nonce" "$rc" >"$meta/status"
exit "$rc"
