# shellcheck shell=bash
# remote run (after watch.bash; installed + exec'd by run.go's boot): lock the workdir, prune it (git
# mode), run CMD under offload_watch, list its outputs, record the status.
# Args: WORKDIR META SUBDIR MODE NONCE COUNT CMD — paths relative to $HOME; MODE = git | plain |
# nosync. Git mode reads NONCE, EXCLUDES (the controller's global git excludes) + COUNT paths
# (NUL-terminated) from stdin first; the rest of stdin is offload's pipe (EOF = offload gone).
# Variables are `__…`, as in watch.bash: CMD inherits this shell's exports.
set -u
fail() {
  printf 'offload: %s: %s\n' "$(uname -n)" "$*" >&2
  exit 125
}
# in_tree PATH: every ancestor of PATH below the workdir is a real directory (no symlink), so PATH
# names something inside the workdir, never a target a symlink points at.
in_tree() {
  local __f=$1 __d=. __rest __comp
  case "/$__f/" in //* | */../* | */./*) return 1 ;; esac
  __rest=${__f%/*}
  [ "$__rest" != "$__f" ] || __rest=
  while [ -n "$__rest" ]; do
    __comp=${__rest%%/*}
    __d=$__d/$__comp
    if [ -L "$__d" ] || [ ! -d "$__d" ]; then return 1; fi
    if [ "$__rest" = "$__comp" ]; then __rest=; else __rest=${__rest#*/}; fi
  done
}
# Prune = delete a file or symlink inside the workdir; never a directory, never recursive. Nonzero
# only when a deletion it attempted failed.
prune_one() {
  if in_tree "$1" && { [ -L "$1" ] || [ -f "$1" ]; }; then rm -f -- "$1" || return 1; fi
}
git_list() { # non-ignored files of the workdir, as git sees them
  git --git-dir=.git --work-tree=. -c core.excludesFile="$__meta/excludes" \
    ls-files -z -co --exclude-standard
}

cd || fail "no home directory"
__wd=$HOME/$1 __meta=$HOME/$2 __sub=$3 __mode=$4 __nonce=$5 __count=$6 __cmd=$7
mkdir -p "$__wd" "$__meta" || fail "cannot create workdir $1"
exec 9>"$__meta/lock"
flock -w 300 9 || fail "workdir $1 still busy after 300 s"
cd "$__wd" || fail "cannot enter workdir $1"
rm -f "$__meta/status" "$__meta/down"

if [ "$__mode" = git ]; then
  [ -d .git ] || fail "workdir $1 has no .git (upload incomplete)"
  head -z -n "$((__count + 2))" >"$__meta/upload"
  __n=$(tr -cd '\0' <"$__meta/upload" | wc -c)
  __first=$(head -z -n 1 "$__meta/upload" | tr -d '\0')
  if [ "$__n" != "$((__count + 2))" ] || [ "$__first" != "$__nonce" ]; then
    fail "upload list framing mismatch ($__n of $((__count + 2)) records); nothing pruned"
  fi
  head -z -n 2 "$__meta/upload" | tail -z -n 1 | tr -d '\0' >"$__meta/excludes"
  git_list >"$__meta/remote" || fail "git ls-files failed in workdir $1; nothing pruned"
  tail -z -n +3 "$__meta/upload" | LC_ALL=C sort -z >"$__meta/upload.sorted"
  LC_ALL=C sort -z "$__meta/remote" | LC_ALL=C comm -z -23 - "$__meta/upload.sorted" >"$__meta/prune"
  while IFS= read -r -d '' __f; do
    prune_one "$__f" || fail "cannot prune $__f in workdir $1"
  done <"$__meta/prune"
fi

cd "./$__sub" || fail "workdir $1 has no directory $__sub"
offload_watch bash -lc "$__cmd"
__rc=$?

# A failed listing leaves no down list ⇒ offload's download fails (125) instead of fetching nothing.
if [ "$__mode" = git ] && cd "$__wd" && git_list >"$__meta/down.all"; then
  while IFS= read -r -d '' __f; do
    if in_tree "$__f" && { [ -e "$__f" ] || [ -L "$__f" ]; }; then printf '%s\0' "$__f"; fi
  done <"$__meta/down.all" >"$__meta/down"
fi
printf '%s %s\n' "$__nonce" "$__rc" >"$__meta/status"
exit "$__rc"
