# shellcheck shell=bash
# local run: offload_watch "$@" (fd 3 = heavy slot lock, held through teardown).
offload_watch "$@"
