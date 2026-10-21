# shellcheck shell=bash
# offload_watch CMD [ARG…] — run CMD in a new session (and, where the user's systemd manager starts
# one, its own transient scope) with stdin /dev/null and no fd 3-9, then end every process of that
# session + scope (TERM, KILL after 10 s) when our stdin (offload's pipe) hits EOF or CMD's leader
# exits. Returns CMD's status. Fds 3-9 the caller opened (locks) stay held by the watcher until
# teardown ends.
# Every variable here is named `__…`: a local shadowing a variable the caller exported is exported
# with the local's value, so a plain name (`unit`, `p`) would reach CMD's environment.
# alive SID EOF MARK: a process of session SID remains, or of the scope whose cgroup MARK names (its
# cgroup.events says populated; systemd removes the dir once empty), or (EOF path) the leader is
# still our child — EOF can arrive before it has called setsid(), when no session SID exists yet.
# The parent check keeps a reaped leader's reused pid out.
alive() {
  local __cg=''
  pgrep -s "$1" >/dev/null && return 0
  [ -n "$3" ] && IFS= read -r __cg <"$3" 2>/dev/null
  [ "${__cg#0::/}" != "$__cg" ] && grep -qx 'populated 1' "/sys/fs/cgroup${__cg#0::}/cgroup.events" 2>/dev/null &&
    return 0
  [ -n "$2" ] && [ "$(ps -o ppid= -p "$1" 2>/dev/null | tr -d ' ')" = "$BASHPID" ]
}

# offload_launch UNIT RT MARK CMD…: start CMD in the background as a session leader; with UNIT,
# inside that scope of the manager under RT, CMD exec'd only once its cgroup line is in MARK (an
# empty MARK ⇒ CMD never ran). Sets __p = the leader's pid.
offload_launch() {
  local __unit=$1 __rt=$2 __mark=$3
  local __hx=${XDG_RUNTIME_DIR+1} __vx=${XDG_RUNTIME_DIR-} __hi=${INVOCATION_ID+1} __vi=${INVOCATION_ID-}
  shift 3
  # A non-interactive bash runs `&` jobs in its own process group ⇒ the job is never a group
  # leader ⇒ setsid(1) calls setsid() without forking ⇒ $! = the new session id. The same bash
  # makes `&` jobs ignore INT + QUIT; `trap -` hands CMD the default dispositions back.
  # systemd-run --scope moves its own pid into the scope, then execs in place (pid kept). It finds
  # the manager through XDG_RUNTIME_DIR alone (so it gets RT) and adds INVOCATION_ID; the wrapper
  # restores both as CMD had them. Its errors (a refused unit, a flag older systemd lacks) go to
  # /dev/null while fd 8 carries stderr through to CMD. --expand-environment=no: by default it
  # rewrites argv (`$$` → `$`, `${X}` → its value).
  (
    trap - INT QUIT
    if [ -n "$__unit" ]; then
      exec 8>&2 2>/dev/null
      XDG_RUNTIME_DIR=$__rt exec setsid systemd-run --user --scope --quiet --collect \
        --expand-environment=no --unit="$__unit" -- bash -c 'IFS= read -r __cg </proc/self/cgroup &&
          printf "%s\n" "$__cg" >"$1" || exit 1
          if [ -n "$2" ]; then export XDG_RUNTIME_DIR="$3"; else unset XDG_RUNTIME_DIR; fi
          if [ -n "$4" ]; then export INVOCATION_ID="$5"; else unset INVOCATION_ID; fi
          shift 5
          exec "$@" 2>&8 8>&-' offload "$__mark" "$__hx" "$__vx" "$__hi" "$__vi" "$@"
    fi
    exec setsid "$@"
  ) </dev/null 3>&- 4>&- 5>&- 6>&- 7>&- 8>&- 9>&- &
  __p=$!
}

offload_watch() {
  local __p __c __who __rc __deadline __eof __d __rt='' __unit='' __mark=''
  # Explicit <&0: without it a non-interactive bash gives & jobs /dev/null as stdin (instant EOF).
  cat <&0 >/dev/null &
  __c=$!
  # The user manager's socket: under our XDG_RUNTIME_DIR, else systemd's default place (offload may
  # run with another runtime dir, or none). cgroup v2 alone: alive() reads the scope's cgroup.events.
  if [ -f /sys/fs/cgroup/cgroup.controllers ]; then
    for __d in "${XDG_RUNTIME_DIR:-}" "/run/user/$UID"; do
      if [ -n "$__d" ] && [ -S "$__d/systemd/private" ]; then
        __rt=$__d
        break
      fi
    done
  fi
  if [ -n "$__rt" ] && __mark=$(mktemp 2>/dev/null); then __unit=offload-$BASHPID-$SRANDOM.scope; fi
  offload_launch "$__unit" "$__rt" "$__mark" "$@"
  wait -n -p __who "$__p" "$__c" 2>/dev/null
  __rc=$?
  # Leader gone with an empty mark: the scope never started, so CMD never ran ⇒ start it plainly.
  if [ -n "$__unit" ] && [ "$__who" = "$__p" ] && [ ! -s "$__mark" ]; then
    printf 'offload: %s: no systemd scope; a process that starts its own session can outlive this run\n' \
      "$(uname -n)" >&2
    __unit=
    offload_launch '' '' '' "$@"
    wait -n -p __who "$__p" "$__c" 2>/dev/null
    __rc=$?
  fi
  __eof=
  if [ "$__who" = "$__c" ]; then __eof=1; else kill "$__c" 2>/dev/null; fi
  # Leader gone (stragglers) or offload gone (everything): end the session + scope. bash reaps the
  # leader on SIGCHLD, so a dead leader never lingers as a zombie.
  pkill -TERM -s "$__p" 2>/dev/null
  [ -z "$__unit" ] || XDG_RUNTIME_DIR=$__rt systemctl --user kill --signal=TERM "$__unit" 2>/dev/null
  [ -z "$__eof" ] || kill -TERM "$__p" 2>/dev/null
  __deadline=$((${EPOCHREALTIME/./} + 10000000)) # µs: KILL 10 s after TERM, however slow pgrep runs
  while alive "$__p" "$__eof" "${__unit:+$__mark}" && ((${EPOCHREALTIME/./} < __deadline)); do
    sleep 0.1
  done
  while alive "$__p" "$__eof" "${__unit:+$__mark}"; do
    pkill -KILL -s "$__p" 2>/dev/null
    [ -z "$__unit" ] || XDG_RUNTIME_DIR=$__rt systemctl --user kill --signal=KILL "$__unit" 2>/dev/null
    [ -z "$__eof" ] || kill -KILL "$__p" 2>/dev/null
    sleep 0.1
  done
  if [ -n "$__eof" ]; then
    wait "$__p" 2>/dev/null
    __rc=$?
  fi
  [ -z "$__mark" ] || rm -f "$__mark" # after teardown: a late wrapper can no longer recreate it
  return "$__rc"
}
