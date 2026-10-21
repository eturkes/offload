# reviewer-7 — pool/slots/probe survivors

Check set = 78 seeded mutants × P1–P3/R1–R2/S1–S5 + Decisions; source e330291. Correctness = observable contract impact; cleanup/altitude = no production rewrite. verdict = killed-by <test> / test-added <test> / contract-open <predicate + why> / equivalent <proof>. Validator: anchored mutant-row unknown count → 0. Seed: 78 rows; original literal grep counted 79 because this preamble repeated its search token. Evidence = exact JSON diff in `.scratch/mut-{pool,slots,probe}.json`, planted individually with `.agent/plant`; all tests use private fixture dirs + fake SSH.

| id | site | mutator | change | ruling |
|----|------|---------|--------|--------|
| po1 | internal/pool/pool.go:95 | arithmetic/base | `return nil, fmt.Errorf("%s: host %d: name %q must match [A-Za-z0-9._-]+", path, i+1, h.Name) ⇒ return nil, fmt.Errorf("%s: host %d: name %q must match [A-Za-z0-` | contract-open P2: invalid name still refused + named; host ordinal wording unprescribed |
| po2 | internal/pool/pool.go:66 | branch/if | `return nil, err ⇒ } ⇒ _ = err ⇒ }` | contract-open P2/P3: read-I/O diagnostic/policy unprescribed; ordinary empty failed read still invalid; partial valid bytes + I/O error unconfirmed (see limits) |
| po3 | internal/pool/pool.go:102 | expression/comparison | `case h.Slots < 1: ⇒ case h.Slots <= 1:` | test-added TestP2OneSlotAndDotName |
| po4 | internal/pool/pool.go:80 | expression/errorf-wrap | `return nil, fmt.Errorf("%s: %w", path, err) ⇒ return nil, fmt.Errorf("%s: %v", path, err)` | equivalent CLI: decode error text unchanged; callers inspect ErrNotExist only, returned before decode |
| po5 | internal/pool/pool.go:94 | expression/remove | `case !validName.MatchString(h.Name) \|\| h.Name == "." \|\| h.Name == "..": ⇒ case !validName.MatchString(h.Name) \|\| false \|\| h.Name == "..":` | test-added TestP2OneSlotAndDotName |
| po6 | internal/pool/pool.go:94 | expression/string-literal | `case !validName.MatchString(h.Name) \|\| h.Name == "." \|\| h.Name == "..": ⇒ case !validName.MatchString(h.Name) \|\| h.Name == "" \|\| h.Name == "..":` | test-added TestP2OneSlotAndDotName |
| po7 | internal/pool/pool.go:95 | numbers/decrementer | `return nil, fmt.Errorf("%s: host %d: name %q must match [A-Za-z0-9._-]+", path, i+1, h.Name) ⇒ return nil, fmt.Errorf("%s: host %d: name %q must match [A-Za-z0-` | contract-open P2: invalid name still refused + named; host ordinal wording unprescribed |
| po8 | internal/pool/pool.go:95 | numbers/incrementer | `return nil, fmt.Errorf("%s: host %d: name %q must match [A-Za-z0-9._-]+", path, i+1, h.Name) ⇒ return nil, fmt.Errorf("%s: host %d: name %q must match [A-Za-z0-` | contract-open P2: invalid name still refused + named; host ordinal wording unprescribed |
| po9 | internal/pool/pool.go:102 | numbers/incrementer | `case h.Slots < 1: ⇒ case h.Slots < 2:` | test-added TestP2OneSlotAndDotName |
| sl1 | internal/slots/slots.go:62 | branch/if | `return Dir{}, err ⇒ } ⇒ _ = err ⇒ }` | test-added TestS1HeavyRejectsInvalidCounts |
| sl2 | internal/slots/slots.go:76 | branch/if | `return nil, false, err ⇒ } ⇒ _ = err ⇒ }` | contract-open S2: valid Total≥1; failed directory creation reaches failed slot open → I/O error; exact failing operation unprescribed |
| sl3 | internal/slots/slots.go:113 | branch/if | `continue // never created ⇒ never held ⇒ } ⇒  ⇒ }` | contract-open R2: newly-created free slot omitted as before; status creating free files unprescribed |
| sl4 | internal/slots/slots.go:131 | branch/if | `return err ⇒ } ⇒ _ = err ⇒ }` | contract-open S2/S3: failed open → nil File.WriteString ErrInvalid; still acquisition failure; error wording unprescribed |
| sl5 | internal/slots/slots.go:135 | branch/if | `err = cerr ⇒ } ⇒ _, _ = err, cerr ⇒ }` | contract-open S3: successful write followed by close error; close-error reporting/durability unprescribed |
| sl6 | internal/slots/slots.go:147 | branch/if | `return nil, false, err ⇒ } ⇒ _ = err ⇒ }` | contract-open S2: failed open → nil File.Fd → EBADF; still I/O failure rather than busy; error wording unprescribed |
| sl7 | internal/slots/slots.go:165 | branch/if | `return nil, false, err ⇒ } ⇒ _ = err ⇒ }` | contract-open S4: failed mkdir reaches failed lock-file open; still I/O failure; exact failing operation unprescribed |
| sl8 | internal/slots/slots.go:69 | composite/field-clear | `return Dir{Path: filepath.Join(RuntimeDir(getenv), "offload", "hosts", name), Total: total} ⇒ return Dir{Total: total}` | killed-by TestR2Status |
| sl9 | internal/slots/slots.go:69 | composite/field-clear | `return Dir{Path: filepath.Join(RuntimeDir(getenv), "offload", "hosts", name), Total: total} ⇒ return Dir{Path: filepath.Join(RuntimeDir(getenv), "offload", "hos` | killed-by TestR2Status |
| sl10 | internal/slots/slots.go:134 | conditional/negated | `if cerr := f.Close(); err == nil { ⇒ if cerr := f.Close(); err != nil {` | test-added TestS3OwnerWriteFailureReleasesSlot |
| sl11 | internal/slots/slots.go:164 | conditional/negated | `if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil { ⇒ if err := os.MkdirAll(filepath.Dir(path), 0o700); err == nil {` | test-added TestS4WorkdirLockCreatesParents |
| sl12 | internal/slots/slots.go:61 | expression/error-guard | `if err != nil { ⇒ return Dir{}, err ⇒ } ⇒ if false { ⇒ return Dir{}, err ⇒ }` | test-added TestS1HeavyRejectsInvalidCounts |
| sl13 | internal/slots/slots.go:146 | expression/error-guard | `if err != nil { ⇒ return nil, false, err ⇒ } ⇒ if false { ⇒ return nil, false, err ⇒ }` | contract-open S2: failed open → nil File.Fd → EBADF; still I/O failure rather than busy; error wording unprescribed |
| sl14 | internal/slots/slots.go:154 | expression/errorf-wrap | `return nil, false, fmt.Errorf("flock %s: %w", path, err) ⇒ return nil, false, fmt.Errorf("flock %s: %v", path, err)` | equivalent CLI: same flock error text; callers check nonnil only (EWOULDBLOCK handled before wrapper) |
| sl15 | internal/slots/slots.go:52 | expression/remove | `if err != nil \|\| n == 0 { ⇒ if err != nil \|\| false {` | equivalent S1: zero-only digits trim to empty → ParseInt error; accepted digits always positive |
| sl16 | internal/slots/slots.go:99 | expression/remove | `if s == nil \|\| s.file == nil { ⇒ if s == nil \|\| false {` | test-added TestS3ReleasePreservesSuccessorOwner |
| sl17 | internal/slots/slots.go:113 | loop/break | `p := filepath.Join(d.Path, "slot"+strconv.Itoa(i)) ⇒ if _, err := os.Lstat(p); errors.Is(err, os.ErrNotExist) { ⇒ continue // never created ⇒ never held ⇒ } ⇒ f` | killed-by TestR2Status |
| sl18 | internal/slots/slots.go:118 | loop/break | `p := filepath.Join(d.Path, "slot"+strconv.Itoa(i)) ⇒ if _, err := os.Lstat(p); errors.Is(err, os.ErrNotExist) { ⇒ continue // never created ⇒ never held ⇒ } ⇒ f` | test-added TestR2HoldersScanOnlyConfiguredRange |
| sl19 | internal/slots/slots.go:48 | numbers/decrementer | `return 0, errors.New("HEAVY_SLOTS must be a positive integer") ⇒ return (-1), errors.New("HEAVY_SLOTS must be a positive integer")` | equivalent CLI: error-return n discarded by Heavy; invalid HEAVY_SLOTS still refused |
| sl20 | internal/slots/slots.go:52 | numbers/decrementer | `if err != nil \|\| n == 0 { ⇒ if err != nil \|\| n == (-1) {` | equivalent S1: zero-only digits trim to empty → ParseInt error; negative digits rejected before parse |
| sl21 | internal/slots/slots.go:53 | numbers/decrementer | `return 0, errors.New("HEAVY_SLOTS must be a positive shell integer") ⇒ return (-1), errors.New("HEAVY_SLOTS must be a positive shell integer")` | equivalent CLI: error-return n discarded by Heavy; invalid HEAVY_SLOTS still refused |
| sl22 | internal/slots/slots.go:110 | numbers/decrementer | `for i := 1; i <= d.Total; i++ { ⇒ for i := 0; i <= d.Total; i++ {` | test-added TestR2HoldersScanOnlyConfiguredRange |
| sl23 | internal/slots/slots.go:164 | numbers/decrementer | `if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil { ⇒ if err := os.MkdirAll(filepath.Dir(path), 0o677); err != nil {` | test-added TestS4WorkdirLockCreatesParents |
| sl24 | internal/slots/slots.go:48 | numbers/incrementer | `return 0, errors.New("HEAVY_SLOTS must be a positive integer") ⇒ return 1, errors.New("HEAVY_SLOTS must be a positive integer")` | equivalent CLI: error-return n discarded by Heavy; invalid HEAVY_SLOTS still refused |
| sl25 | internal/slots/slots.go:75 | numbers/incrementer | `if err := os.MkdirAll(d.Path, 0o700); err != nil { ⇒ if err := os.MkdirAll(d.Path, 0o701); err != nil {` | contract-open S2: 0701 preserves owner access + flock; directory other-execute bit unprescribed |
| sl26 | internal/slots/slots.go:53 | numbers/incrementer | `return 0, errors.New("HEAVY_SLOTS must be a positive shell integer") ⇒ return 1, errors.New("HEAVY_SLOTS must be a positive shell integer")` | equivalent CLI: error-return n discarded by Heavy; invalid HEAVY_SLOTS still refused |
| sl27 | internal/slots/slots.go:164 | numbers/incrementer | `if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil { ⇒ if err := os.MkdirAll(filepath.Dir(path), 0o701); err != nil {` | contract-open S4: 0701 preserves owner access + flock; directory other-execute bit unprescribed |
| sl28 | internal/slots/slots.go:129 | numbers/incrementer | `f, err := os.OpenFile(path, os.O_WRONLY\|os.O_CREATE\|os.O_TRUNC\|syscall.O_NOFOLLOW, 0o600) //nolint:gosec // G304: runtime dir path ⇒ f, err := os.OpenFile(path,` | contract-open S3: 0601 preserves owner write/read; owner-file other-execute bit unprescribed |
| sl29 | internal/slots/slots.go:145 | numbers/incrementer | `f, err = os.OpenFile(path, os.O_WRONLY\|os.O_CREATE\|syscall.O_NOFOLLOW, 0o600) //nolint:gosec // G304: runtime dir path ⇒ f, err = os.OpenFile(path, os.O_WRONLY\|` | contract-open S2: 0601 preserves owner write/read + flock; slot-file other-execute bit unprescribed |
| sl30 | internal/slots/slots.go:89 | statement/remove | `s.Release() ⇒ return nil, false, err ⇒ } ⇒ _ = s.Release ⇒ return nil, false, err ⇒ }` | test-added TestS3OwnerWriteFailureReleasesSlot |
| sl31 | internal/slots/slots.go:104 | statement/remove | `s.file = nil ⇒ _ = s.file` | test-added TestS3ReleasePreservesSuccessorOwner |
| sl32 | internal/slots/slots.go:135 | statement/remove | `err = cerr ⇒ } ⇒ _, _ = err, cerr ⇒ }` | contract-open S3: successful write followed by close error; close-error reporting/durability unprescribed |
| sl33 | internal/slots/slots.go:35 | statement/return | `return "/tmp" ⇒ return ""` | test-added TestS1RuntimeDirFallback |
| sl34 | internal/slots/slots.go:69 | statement/return | `return Dir{Path: filepath.Join(RuntimeDir(getenv), "offload", "hosts", name), Total: total} ⇒ return Dir{}` | killed-by TestR2Status |
| sl35 | internal/slots/slots.go:62 | statement/return | `return Dir{}, err ⇒ } ⇒ return Dir{}, nil ⇒ }` | test-added TestS1HeavyRejectsInvalidCounts |
| pr1 | internal/probe/probe.go:77 | arithmetic/base | `if msg := lines[len(lines)-1]; msg != "" { ⇒ if msg := lines[len(lines) + 1]; msg != "" {` | killed-by TestR2Status |
| pr2 | internal/probe/probe.go:73 | branch/if | `return Stats{}, fmt.Errorf("no answer within %s", Timeout) ⇒ } ⇒ _ = fmt.Errorf ⇒ }` | contract-open R1: deadline still kills/reaps and reports down via ssh error; timeout-message wording unprescribed |
| pr3 | internal/probe/probe.go:76 | branch/if | `lines := strings.Split(strings.TrimSpace(stderr.String()), "\n") ⇒ if msg := lines[len(lines)-1]; msg != "" { ⇒ return Stats{}, fmt.Errorf("%s", msg) ⇒ } ⇒ retu` | killed-by TestR2Status |
| pr4 | internal/probe/probe.go:78 | branch/if | `return Stats{}, fmt.Errorf("%s", msg) ⇒ } ⇒ _, _ = fmt.Errorf, msg ⇒ }` | killed-by TestR2Status |
| pr5 | internal/probe/probe.go:72 | conditional/negated | `if ctx.Err() != nil { ⇒ if ctx.Err() == nil {` | killed-by TestR2Status |
| pr6 | internal/probe/probe.go:75 | conditional/negated | `if err != nil { ⇒ if err == nil {` | killed-by TestR2Status |
| pr7 | internal/probe/probe.go:77 | conditional/negated | `if msg := lines[len(lines)-1]; msg != "" { ⇒ if msg := lines[len(lines)-1]; msg == "" {` | killed-by TestR2Status |
| pr8 | internal/probe/probe.go:46 | expression/comparison | `if s.CPUs, err = strconv.Atoi(f[0]); err != nil \|\| s.CPUs < 1 { ⇒ if s.CPUs, err = strconv.Atoi(f[0]); err != nil \|\| s.CPUs <= 1 {` | test-added TestR1ParseBoundaries |
| pr9 | internal/probe/probe.go:49 | expression/comparison | `if s.Load, err = strconv.ParseFloat(f[1], 64); err != nil \|\| !(s.Load >= 0) \|\| math.IsInf(s.Load, 1) { ⇒ if s.Load, err = strconv.ParseFloat(f[1], 64); err != n` | test-added TestR1ParseBoundaries |
| pr10 | internal/probe/probe.go:52 | expression/comparison | `if s.MemKB, err = strconv.ParseInt(f[2], 10, 64); err != nil \|\| s.MemKB < 0 { ⇒ if s.MemKB, err = strconv.ParseInt(f[2], 10, 64); err != nil \|\| s.MemKB <= 0 {` | test-added TestR1ParseBoundaries |
| pr11 | internal/probe/probe.go:80 | expression/errorf-wrap | `return Stats{}, fmt.Errorf("ssh: %w", err) ⇒ return Stats{}, fmt.Errorf("ssh: %v", err)` | equivalent CLI: same ssh failure text; callers use nonnil/string only, never unwrap |
| pr12 | internal/probe/probe.go:46 | expression/remove | `if s.CPUs, err = strconv.Atoi(f[0]); err != nil \|\| s.CPUs < 1 { ⇒ if s.CPUs, err = strconv.Atoi(f[0]); false \|\| ⇒ s.CPUs < 1 {` | test-added TestR1MalformedCounters |
| pr13 | internal/probe/probe.go:52 | expression/remove | `if s.MemKB, err = strconv.ParseInt(f[2], 10, 64); err != nil \|\| s.MemKB < 0 { ⇒ if s.MemKB, err = strconv.ParseInt(f[2], 10, 64); false \|\| ⇒ s.MemKB < 0 {` | test-added TestR1MalformedCounters |
| pr14 | internal/probe/probe.go:55 | expression/remove | `if s.GPUs, err = strconv.Atoi(f[3]); err != nil \|\| s.GPUs < 0 { ⇒ if s.GPUs, err = strconv.Atoi(f[3]); false \|\| ⇒ s.GPUs < 0 {` | test-added TestR1MalformedCounters |
| pr15 | internal/probe/probe.go:23 | numbers/decrementer | `Timeout   = 8 * time.Second ⇒ Timeout   = 7 * time.Second` | contract-open R1: 7 s remains within ≤8 s; earliest timeout/response-acceptance window unprescribed |
| pr16 | internal/probe/probe.go:24 | numbers/decrementer | `pipeGrace = 500 * time.Millisecond ⇒ pipeGrace = 499 * time.Millisecond` | contract-open R1: pipe-grace −1 ms, execution budget +1 ms; total stays 8 s; split unprescribed |
| pr17 | internal/probe/probe.go:49 | numbers/decrementer | `if s.Load, err = strconv.ParseFloat(f[1], 64); err != nil \|\| !(s.Load >= 0) \|\| math.IsInf(s.Load, 1) { ⇒ if s.Load, err = strconv.ParseFloat(f[1], 63); err != n` | equivalent strconv.ParseFloat: every bitSize other than 32 uses float64, including 63 |
| pr18 | internal/probe/probe.go:49 | numbers/decrementer | `if s.Load, err = strconv.ParseFloat(f[1], 64); err != nil \|\| !(s.Load >= 0) \|\| math.IsInf(s.Load, 1) { ⇒ if s.Load, err = strconv.ParseFloat(f[1], 64); err != n` | equivalent R1: IsInf sign 0 adds negative infinity, already refused by !(Load≥0) |
| pr19 | internal/probe/probe.go:50 | numbers/decrementer | `return Stats{}, fmt.Errorf("probe: bad load %q", f[1]) ⇒ return Stats{}, fmt.Errorf("probe: bad load %q", f[0])` | test-added TestR1MalformedReasonNamesRejectedValue |
| pr20 | internal/probe/probe.go:52 | numbers/decrementer | `if s.MemKB, err = strconv.ParseInt(f[2], 10, 64); err != nil \|\| s.MemKB < 0 { ⇒ if s.MemKB, err = strconv.ParseInt(f[2], 10, 63); err != nil \|\| s.MemKB < 0 {` | test-added TestR1ParseBoundaries |
| pr21 | internal/probe/probe.go:52 | numbers/decrementer | `if s.MemKB, err = strconv.ParseInt(f[2], 10, 64); err != nil \|\| s.MemKB < 0 { ⇒ if s.MemKB, err = strconv.ParseInt(f[2], 10, 64); err != nil \|\| s.MemKB < (-1) {` | test-added TestR1MalformedCounters |
| pr22 | internal/probe/probe.go:53 | numbers/decrementer | `return Stats{}, fmt.Errorf("probe: bad MemAvailable %q", f[2]) ⇒ return Stats{}, fmt.Errorf("probe: bad MemAvailable %q", f[1])` | test-added TestR1MalformedReasonNamesRejectedValue |
| pr23 | internal/probe/probe.go:56 | numbers/decrementer | `return Stats{}, fmt.Errorf("probe: bad GPU count %q", f[3]) ⇒ return Stats{}, fmt.Errorf("probe: bad GPU count %q", f[2])` | test-added TestR1MalformedReasonNamesRejectedValue |
| pr24 | internal/probe/probe.go:77 | numbers/decrementer | `if msg := lines[len(lines)-1]; msg != "" { ⇒ if msg := lines[len(lines)-0]; msg != "" {` | killed-by TestR2Status |
| pr25 | internal/probe/probe.go:24 | numbers/incrementer | `pipeGrace = 500 * time.Millisecond ⇒ pipeGrace = 501 * time.Millisecond` | contract-open R1: pipe-grace +1 ms, execution budget −1 ms; total stays 8 s; split unprescribed |
| pr26 | internal/probe/probe.go:46 | numbers/incrementer | `if s.CPUs, err = strconv.Atoi(f[0]); err != nil \|\| s.CPUs < 1 { ⇒ if s.CPUs, err = strconv.Atoi(f[0]); err != nil \|\| s.CPUs < 2 {` | test-added TestR1ParseBoundaries |
| pr27 | internal/probe/probe.go:47 | numbers/incrementer | `return Stats{}, fmt.Errorf("probe: bad CPU count %q", f[0]) ⇒ return Stats{}, fmt.Errorf("probe: bad CPU count %q", f[1])` | test-added TestR1MalformedReasonNamesRejectedValue |
| pr28 | internal/probe/probe.go:52 | numbers/incrementer | `if s.MemKB, err = strconv.ParseInt(f[2], 10, 64); err != nil \|\| s.MemKB < 0 { ⇒ if s.MemKB, err = strconv.ParseInt(f[2], 10, 64); err != nil \|\| s.MemKB < 1 {` | test-added TestR1ParseBoundaries |
| pr29 | internal/probe/probe.go:49 | numbers/incrementer | `if s.Load, err = strconv.ParseFloat(f[1], 64); err != nil \|\| !(s.Load >= 0) \|\| math.IsInf(s.Load, 1) { ⇒ if s.Load, err = strconv.ParseFloat(f[1], 65); err != n` | equivalent strconv.ParseFloat: every bitSize other than 32 uses float64, including 65 |
| pr30 | internal/probe/probe.go:49 | numbers/incrementer | `if s.Load, err = strconv.ParseFloat(f[1], 64); err != nil \|\| !(s.Load >= 0) \|\| math.IsInf(s.Load, 1) { ⇒ if s.Load, err = strconv.ParseFloat(f[1], 64); err != n` | equivalent math.IsInf: positive sign 2 tests positive infinity exactly as sign 1 |
| pr31 | internal/probe/probe.go:50 | numbers/incrementer | `return Stats{}, fmt.Errorf("probe: bad load %q", f[1]) ⇒ return Stats{}, fmt.Errorf("probe: bad load %q", f[2])` | test-added TestR1MalformedReasonNamesRejectedValue |
| pr32 | internal/probe/probe.go:53 | numbers/incrementer | `return Stats{}, fmt.Errorf("probe: bad MemAvailable %q", f[2]) ⇒ return Stats{}, fmt.Errorf("probe: bad MemAvailable %q", f[3])` | test-added TestR1MalformedReasonNamesRejectedValue |
| pr33 | internal/probe/probe.go:77 | numbers/incrementer | `if msg := lines[len(lines)-1]; msg != "" { ⇒ if msg := lines[len(lines)-2]; msg != "" {` | killed-by TestR2Status |
| pr34 | internal/probe/probe.go:70 | statement/remove | `//nolint:gosec // G204: address = trusted pool file entry, after "--"; Script is a constant ⇒  ⇒ // Descendants of a killed ssh can keep its pipes open; WaitDel` | killed-by TestR2Status |

## Plant evidence

Individual plants on e330291 + reviewer7 tests; GOFLAGS=-p=1; stdout below. Each reported FAIL = underlying go test rc 1; plant wrapper itself returns 0 for killed/survived. Regex = anchored test name in its verdict cell. Inputs reconstructed from original mutant JSON diff (full context, old/new lines).

- `po3: killed by --- FAIL: TestP2OneSlotAndDotName (0.00s)` (wrapper rc 0)
- `po5: killed by --- FAIL: TestP2OneSlotAndDotName (0.00s)` (wrapper rc 0)
- `po6: killed by --- FAIL: TestP2OneSlotAndDotName (0.00s)` (wrapper rc 0)
- `po9: killed by --- FAIL: TestP2OneSlotAndDotName (0.00s)` (wrapper rc 0)
- `sl1: killed by --- FAIL: TestS1HeavyRejectsInvalidCounts (0.00s)` (wrapper rc 0)
- `sl8: killed by --- FAIL: TestR2Status (0.01s)` (wrapper rc 0)
- `sl9: killed by --- FAIL: TestR2Status (0.01s)` (wrapper rc 0)
- `sl10: killed by --- FAIL: TestS3OwnerWriteFailureReleasesSlot (0.00s)` (wrapper rc 0)
- `sl11: killed by --- FAIL: TestS4WorkdirLockCreatesParents (0.00s)` (wrapper rc 0)
- `sl12: killed by --- FAIL: TestS1HeavyRejectsInvalidCounts (0.00s)` (wrapper rc 0)
- `sl16: killed by --- FAIL: TestS3ReleasePreservesSuccessorOwner (0.00s)` (wrapper rc 0)
- `sl17: killed by --- FAIL: TestR2Status (0.02s)` (wrapper rc 0)
- `sl18: killed by --- FAIL: TestR2HoldersScanOnlyConfiguredRange (0.00s)` (wrapper rc 0)
- `sl22: killed by --- FAIL: TestR2HoldersScanOnlyConfiguredRange (0.00s)` (wrapper rc 0)
- `sl23: killed by --- FAIL: TestS4WorkdirLockCreatesParents (0.00s)` (wrapper rc 0)
- `sl30: killed by --- FAIL: TestS3OwnerWriteFailureReleasesSlot (0.00s)` (wrapper rc 0)
- `sl31: killed by --- FAIL: TestS3ReleasePreservesSuccessorOwner (0.00s)` (wrapper rc 0)
- `sl33: killed by --- FAIL: TestS1RuntimeDirFallback (0.00s)` (wrapper rc 0)
- `sl34: killed by --- FAIL: TestR2Status (0.02s)` (wrapper rc 0)
- `sl35: killed by --- FAIL: TestS1HeavyRejectsInvalidCounts (0.00s)` (wrapper rc 0)
- `pr1: killed by --- FAIL: TestR2Status (0.07s)` (wrapper rc 0)
- `pr3: killed by --- FAIL: TestR2Status (0.06s)` (wrapper rc 0)
- `pr4: killed by --- FAIL: TestR2Status (0.05s)` (wrapper rc 0)
- `pr5: killed by --- FAIL: TestR2Status (0.06s)` (wrapper rc 0)
- `pr6: killed by --- FAIL: TestR2Status (0.09s)` (wrapper rc 0)
- `pr7: killed by --- FAIL: TestR2Status (0.06s)` (wrapper rc 0)
- `pr8: killed by --- FAIL: TestR1ParseBoundaries (0.00s)` (wrapper rc 0)
- `pr9: killed by --- FAIL: TestR1ParseBoundaries (0.00s)` (wrapper rc 0)
- `pr10: killed by --- FAIL: TestR1ParseBoundaries (0.00s)` (wrapper rc 0)
- `pr12: killed by --- FAIL: TestR1MalformedCounters (0.00s)` (wrapper rc 0)
- `pr13: killed by --- FAIL: TestR1MalformedCounters (0.00s)` (wrapper rc 0)
- `pr14: killed by --- FAIL: TestR1MalformedCounters (0.00s)` (wrapper rc 0)
- `pr19: killed by --- FAIL: TestR1MalformedReasonNamesRejectedValue (0.00s)` (wrapper rc 0)
- `pr20: killed by --- FAIL: TestR1ParseBoundaries (0.00s)` (wrapper rc 0)
- `pr21: killed by --- FAIL: TestR1MalformedCounters (0.00s)` (wrapper rc 0)
- `pr22: killed by --- FAIL: TestR1MalformedReasonNamesRejectedValue (0.00s)` (wrapper rc 0)
- `pr23: killed by --- FAIL: TestR1MalformedReasonNamesRejectedValue (0.00s)` (wrapper rc 0)
- `pr24: killed by --- FAIL: TestR2Status (0.06s)` (wrapper rc 0)
- `pr26: killed by --- FAIL: TestR1ParseBoundaries (0.00s)` (wrapper rc 0)
- `pr27: killed by --- FAIL: TestR1MalformedReasonNamesRejectedValue (0.00s)` (wrapper rc 0)
- `pr28: killed by --- FAIL: TestR1ParseBoundaries (0.00s)` (wrapper rc 0)
- `pr31: killed by --- FAIL: TestR1MalformedReasonNamesRejectedValue (0.00s)` (wrapper rc 0)
- `pr32: killed by --- FAIL: TestR1MalformedReasonNamesRejectedValue (0.00s)` (wrapper rc 0)
- `pr33: killed by --- FAIL: TestR2Status (0.02s)` (wrapper rc 0)
- `pr34: killed by --- FAIL: TestR2Status (0.04s)` (wrapper rc 0)
- Revised fixture (child-owned TempDir): `sl10: killed by --- FAIL: TestS3OwnerWriteFailureReleasesSlot (0.00s)` (wrapper rc 0)
- Revised fixture (child-owned TempDir): `sl30: killed by --- FAIL: TestS3OwnerWriteFailureReleasesSlot (0.00s)` (wrapper rc 0)

## Adjudication

78/78 rows ruled; 13 existing-suite kills + 32 added-test kills + 20 contract-open + 13 equivalent. New production defects = none demonstrated. Newly added regression tests = 10 across `internal/{pool,probe,slots}/reviewer7_test.go`. Every added test passed on e330291 and killed ≥1 seeded mutant; production diff after all plants = empty (git diff --exit-code rc 0).

## Proof notes

- po4: `internal/cli/cli.go:150–156` uses only `errors.Is(err, pool.ErrNotExist)` + nonnil; decode errors cannot carry ErrNotExist. `%w` → `%v` leaves printed text unchanged.
- sl4/sl6/sl13: Go 1.27.3 `os.File.fd` returns all-one uintptr for nil; flock(-1) returns EBADF, not contention. `File.WriteString` delegates to Write; nil File returns ErrInvalid. Error guard mutants change the reported I/O error, not admission.
- sl14/pr11: callers consume nonnil errors + text, never these wrapped identities. EWOULDBLOCK handled before the sl14 wrapper.
- sl15/sl20: digit regex excludes signs; TrimLeft removes every leading zero. Zero-only input becomes empty (ParseInt error); any successfully parsed remaining input ≥1. Overflow stays an error.
- sl19/sl21/sl24/sl26: Heavy returns immediately on HeavySlots error; numeric failure result never reaches Dir.Total.
- pr17/pr29: Go 1.27.3 `src/internal/strconv/atof.go:727–732`: `if bitSize == 32` selects atof32; every other value selects atof64. Parse rejects error without exposing strconv's error detail.
- pr18/pr30: `math.IsInf` treats every positive sign alike; sign 0 additionally tests negative infinity, already rejected by Load≥0.
- pr15/pr16/pr25: R1 caps end-to-end latency; it does not promise an acceptance window to the final instant. Deadline/grace sum stays ≤8 s.

## Checks + limits

- Baseline focused new tests: rc 0, all 10 passed.
- 45 individual original mutants: killed; no SURVIVED or compile-only failures. sl10/sl30 killed again after test-fixture repair.
- Fast gate = `offload run --local -- env GOFLAGS=-p=1 GOCACHE=<WT>/.tools/go-cache GOLANGCI_LINT_CACHE=<WT>/.tools/lint-cache .agent/gate fast ./internal/pool ./internal/probe ./internal/slots ./test/e2e`.
- First fast gate: fmt/test/spec-lint passed; lint failed `internal/slots/slots.go:164:23: G703: Path traversal via taint analysis (gosec)` because the added helper passed its fixture path through env. Test-only repair: child creates TempDir; env selects helper mode only.
- Repaired fast gate: rc 0; fmt 0.1 s, lint 0.5 s (0 issues), test 0.5 s, spec-lint 0.0 s. 190 tests; pool/probe/e2e cached from preceding successful 57.250 s test run; slots reran (94 ms).
- Full gate + default-operator mutation sweep not rerun; this work adjudicates exactly the supplied 78 package-scope survivors. No real hosts or real heavy paths touched except offload's own --local acquisition. LSP call on probe.go: no Go server available.
- po2 partial-read caveat: os.ReadFile can theoretically return valid-prefix bytes + nonnil I/O error. No controllable read-error filesystem was exercised; whole-file validation under that fault remains unconfirmed. Ordinary open/read failures return empty data → invalid config. Ruling is contract-open for read-I/O policy, not universal equivalence.
- Other static contract-open/equivalent entries are source/contract adjudications, not claimed mutant kills. Mode-bit entries concern owner access/locking contract only; broader permission policy lies outside this brief.
- Seed validator positive control: 78 mutant rows plus one matching preamble gave original count 79; preamble corrected. Final validator: 0, rc 1 (grep's no-match status). No instruction/spec edits.

Commit = `499014b2252b48c8fc308d5eee71762de134e130` on `wt/reviewer-7`; tracked working tree clean. Staged patch check rc 0; all source files byte-identical to e330291 (git diff --exit-code rc 0). reviewer-7-DONE-1
