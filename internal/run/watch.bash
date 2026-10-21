# shellcheck shell=bash
# offload_watch CMD [ARG…] — run CMD in a new session with stdin /dev/null and no fd 3-9, then end
# every process of that session (TERM, KILL after 10 s) when our stdin (offload's pipe) hits EOF or
# CMD's leader exits. Returns CMD's status. Fds 3-9 the caller opened (locks) stay held by the
# watcher until teardown ends.
offload_watch() {
  local p c who rc deadline
  # Explicit <&0: without it a non-interactive bash gives & jobs /dev/null as stdin (instant EOF).
  cat <&0 >/dev/null &
  c=$!
  # A non-interactive bash runs `&` jobs in its own process group ⇒ the job is never a group
  # leader ⇒ setsid(1) calls setsid() without forking ⇒ $! = the new session id. The same bash
  # makes `&` jobs ignore INT + QUIT; `trap -` hands CMD the default dispositions back.
  (
    trap - INT QUIT
    exec setsid "$@"
  ) </dev/null 3>&- 4>&- 5>&- 6>&- 7>&- 8>&- 9>&- &
  p=$!
  wait -n -p who "$p" "$c" 2>/dev/null
  rc=$?
  [ "$who" = "$c" ] || kill "$c" 2>/dev/null
  # Leader gone (stragglers) or offload gone (everything): end the session. bash reaps the leader
  # on SIGCHLD, so a dead leader never lingers as a zombie pgrep would count.
  pkill -TERM -s "$p" 2>/dev/null
  deadline=$((${EPOCHREALTIME/./} + 10000000)) # µs: KILL 10 s after TERM, however slow pgrep runs
  while pgrep -s "$p" >/dev/null && ((${EPOCHREALTIME/./} < deadline)); do
    sleep 0.1
  done
  if pgrep -s "$p" >/dev/null; then
    pkill -KILL -s "$p" 2>/dev/null
    while pgrep -s "$p" >/dev/null; do sleep 0.1; done
  fi
  if [ "$who" = "$c" ]; then
    wait "$p" 2>/dev/null
    rc=$?
  fi
  return "$rc"
}
