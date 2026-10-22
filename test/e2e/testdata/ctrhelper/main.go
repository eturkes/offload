// Command ctrhelper is the only program in the e2e test image (`FROM scratch`): each mode is one
// behavior a K predicate needs observed from inside the container.
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "ctrhelper MODE [ARG…]")
		os.Exit(64)
	}
	arg := func(i int) string {
		if len(os.Args) > i {
			return os.Args[i]
		}
		return ""
	}
	switch os.Args[1] {
	case "env": // env NAME…: NAME=value | NAME unset, one line each
		for _, name := range os.Args[2:] {
			if v, ok := os.LookupEnv(name); ok {
				fmt.Printf("%s=%s\n", name, v)
			} else {
				fmt.Printf("%s unset\n", name)
			}
		}
	case "pwd":
		wd, err := os.Getwd()
		check(err)
		fmt.Println(wd)
	case "id":
		fmt.Println(os.Getuid(), os.Getgid())
	case "write": // write PATH TEXT
		check(os.WriteFile(arg(2), []byte(arg(3)), 0o644))
	case "read": // read PATH
		b, err := os.ReadFile(arg(2))
		check(err)
		os.Stdout.Write(b)
	case "exists": // exists PATH: rc 0 | 1
		if _, err := os.Stat(arg(2)); err != nil {
			os.Exit(1)
		}
	case "exit": // exit N
		n, err := strconv.Atoi(arg(2))
		check(err)
		os.Exit(n)
	case "cat":
		_, err := io.Copy(os.Stdout, os.Stdin)
		check(err)
	case "bytes": // raw bytes on both streams, no trailing newline
		os.Stdout.Write([]byte("out\x00\xff\r"))
		os.Stderr.Write([]byte("err\x01\xfe"))
	case "selfkill": // selfkill SIGNUM
		n, err := strconv.Atoi(arg(2))
		check(err)
		check(syscall.Kill(os.Getpid(), syscall.Signal(n)))
		time.Sleep(10 * time.Second)
	case "trap": // trap FILE SECONDS EXIT: log each HUP/INT/TERM to FILE for SECONDS, then exit EXIT
		secs, err := strconv.Atoi(arg(3))
		check(err)
		code, err := strconv.Atoi(arg(4))
		check(err)
		ch := make(chan os.Signal, 16)
		signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		f, err := os.Create(arg(2))
		check(err)
		fmt.Fprintln(f, "ready")
		end := time.After(time.Duration(secs) * time.Second)
		for {
			select {
			case s := <-ch:
				fmt.Fprintln(f, strings.ToUpper(s.String()))
			case <-end:
				os.Exit(code)
			}
		}
	case "tick": // tick FILE READY: ignore HUP/INT/TERM, append the UnixNano time to FILE every ms
		signal.Ignore(syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
		f, err := os.OpenFile(arg(2), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		check(err)
		check(os.WriteFile(arg(3), nil, 0o644))
		for {
			fmt.Fprintln(f, time.Now().UnixNano())
			time.Sleep(time.Millisecond)
		}
	case "ignore": // ignore MARKER [READY]: ignore HUP/INT/TERM, create READY, sleep forever (MARKER names it in ps)
		signal.Ignore(syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
		if ready := arg(3); ready != "" {
			check(os.WriteFile(ready, nil, 0o644))
		}
		if os.Getenv("CTRHELPER_PIPE") == "3" { // spawn's pipe; fd 3 may be the runtime's otherwise
			_ = os.NewFile(3, "ready").Close()
		}
		for { // not `select {}`: with no other goroutine the runtime kills that as a deadlock
			time.Sleep(time.Hour)
		}
	case "spawn": // spawn MARKER EXIT: start `ignore MARKER`, exit EXIT once it ignores TERM
		code, err := strconv.Atoi(arg(3))
		check(err)
		r, w, err := os.Pipe()
		check(err)
		c := exec.Command(os.Args[0], "ignore", arg(2))
		c.ExtraFiles = []*os.File{w}
		c.Env = append(os.Environ(), "CTRHELPER_PIPE=3")
		check(c.Start())
		_ = w.Close()
		_, _ = r.Read(make([]byte, 1)) // EOF once the child closed its end, after signal.Ignore
		os.Exit(code)
	default:
		fmt.Fprintf(os.Stderr, "ctrhelper: unknown mode %q\n", os.Args[1])
		os.Exit(64)
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "ctrhelper:", err)
		os.Exit(70)
	}
}
