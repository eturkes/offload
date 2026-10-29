# Mutation survivors — IMPLEMENT close (Mutago v2.10.35, default operators, package tests only, at 69861da)

Ruling ∈ {killed-by (full-suite test name, shown with `.agent/plant`), test-added (name), contract-open (why), equivalent (why)}. Ruled by reviewer-6 (mirror) + reviewer-7 (pool/slots/probe), proofs via `.agent/plant`; remote hand-planted by MAIN.

## pool (9 escaped, MSI 0.83)

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

## slots (35 escaped, MSI 0.69)

| id | site | mutator | change | ruling |
|----|------|---------|--------|--------|
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

## mirror (70 escaped, MSI 0.65)

| id | site | mutator | change | ruling |
|----|------|---------|--------|--------|
| mi1 | internal/mirror/mirror.go:135 | branch/case | `return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p) ⇒ _, _, _ = fmt.Errorf, ErrLocalOnly, p` | test-added TestMirrorY3NestedReplacesTracked |
| mi2 | internal/mirror/mirror.go:133 | branch/case | `return nil, err ⇒ _ = err` | test-added TestMirrorY4LstatError |
| mi3 | internal/mirror/mirror.go:38 | branch/if | `return Tree{}, err ⇒ } ⇒ _ = err ⇒ }` | test-added TestMirrorY1GitFailurePlainTree |
| mi4 | internal/mirror/mirror.go:53 | branch/if | `return Tree{}, fmt.Errorf("git rev-parse printed %d lines, want 4", len(f)) ⇒ } ⇒ _, _ = fmt.Errorf, f ⇒ }` | test-added TestMirrorY1RevParseFraming |
| mi5 | internal/mirror/mirror.go:57 | branch/if | `t.Sub = "." ⇒ } ⇒ _ = t.Sub ⇒ }` | equivalent filepath.Clean always returns nonempty; empty prefix already becomes "." (Y1) |
| mi6 | internal/mirror/mirror.go:70 | branch/if | `return "", err ⇒ } ⇒ _ = err ⇒ }` | contract-open Y3: non-ENOENT .git lstat failure with split git dirs can become LocalOnly; fatal-vs-local fallback unspecified |
| mi7 | internal/mirror/mirror.go:73 | branch/if | `return "it is a linked worktree", nil ⇒ } ⇒  ⇒ }` | test-added TestMirrorY3SplitGitDirs/common-outside |
| mi8 | internal/mirror/mirror.go:76 | branch/if | `return "its git dir lies outside the tree", sameErr ⇒ } ⇒ _ = sameErr ⇒ }` | test-added TestMirrorY3SplitGitDirs/both-outside |
| mi9 | internal/mirror/mirror.go:83 | branch/if | `return "", err ⇒ } ⇒ _ = err ⇒ }` | test-added TestMirrorY1IndexFailure |
| mi10 | internal/mirror/mirror.go:96 | branch/if | `return false, err ⇒ } ⇒ _ = err ⇒ }` | contract-open Y3: git-dir Stat failure still yields LocalOnly (SameFile(nil, x)=false); hybrid fatal-vs-local fallback unspecified |
| mi11 | internal/mirror/mirror.go:100 | branch/if | `return false, err ⇒ } ⇒ _ = err ⇒ }` | contract-open Y3: .git Stat failure still yields LocalOnly (SameFile(x, nil)=false); hybrid fatal-vs-local fallback unspecified |
| mi12 | internal/mirror/mirror.go:113 | branch/if | `return nil, err ⇒ } ⇒ _ = err ⇒ }` | test-added TestMirrorY4ListErrors/-co |
| mi13 | internal/mirror/mirror.go:117 | branch/if | `return nil, err ⇒ } ⇒ _ = err ⇒ }` | test-added TestMirrorY4ListErrors/-i |
| mi14 | internal/mirror/mirror.go:123 | branch/if | `return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p) ⇒ } ⇒ _, _, _ = fmt.Errorf, ErrLocalOnly, p ⇒ }` | equivalent Y3/Y4: real-git trailing-/ entries name nested repo dirs; quiescent-tree Lstat.IsDir returns the same ErrLocalOnly |
| mi15 | internal/mirror/mirror.go:173 | branch/if | `if !errors.As(err, &ge) \|\| ge.code != 1 { // 1 = unset ⇒ return "", err ⇒ } ⇒ if !errors.As(err, &ge) \|\| ge.code != 1 { ⇒ _ = err ⇒ }` | test-added TestMirrorY5ExcludesErrors/git-config |
| mi16 | internal/mirror/mirror.go:246 | branch/if | `cmd.Stdin = bytes.NewReader(stdin) ⇒ } ⇒ _, _, _ = cmd.Stdin, bytes.NewReader, stdin ⇒ }` | killed-by TestY1RootAndSubdirectory |
| mi17 | internal/mirror/mirror.go:250 | branch/if | `return fmt.Errorf("rsync: %w", err) ⇒ } ⇒ _, _ = fmt.Errorf, err ⇒ }` | killed-by TestY10DownloadFailureOverridesCommand/0 |
| mi18 | internal/mirror/mirror.go:271 | branch/if | `return err ⇒ } ⇒ _ = err ⇒ }` | test-added TestMirrorY10FileUploadFailure |
| mi19 | internal/mirror/mirror.go:305 | branch/if | `dst = filepath.Join(t.Root, filepath.Dir(rel)) + "/" ⇒ } ⇒ _, _, _, _, _ = dst, filepath.Join, t.Root, filepath.Dir, rel ⇒ }` | killed-by TestY8Get/ignored-and-directory |
| mi20 | internal/mirror/mirror.go:233 | composite/field-clear | `return out, &gitError{args: args, code: ee.ExitCode(), stderr: stderr.String()} ⇒ return out, &gitError{code: ee.ExitCode(), stderr: stderr.String()}` | contract-open Y1/Y10 require tool diagnostic, not echoed git argv; stderr retained |
| mi21 | internal/mirror/mirror.go:46 | conditional/negated | `if d == filepath.Dir(d) { ⇒ if d != filepath.Dir(d) {` | test-added TestMirrorY1DamagedAncestor |
| mi22 | internal/mirror/mirror.go:249 | conditional/negated | `if err := cmd.Run(); err != nil { ⇒ if err := cmd.Run(); err == nil {` | killed-by TestY1RootAndSubdirectory |
| mi23 | internal/mirror/mirror.go:245 | conditional/negated | `if stdin != nil { ⇒ if stdin == nil {` | killed-by TestY1RootAndSubdirectory |
| mi24 | internal/mirror/mirror.go:270 | conditional/negated | `if err := rsync(ctx, in.Bytes(), files...); err != nil { ⇒ if err := rsync(ctx, in.Bytes(), files...); err == nil {` | killed-by TestY1RootAndSubdirectory |
| mi25 | internal/mirror/mirror.go:304 | conditional/not | `if !strings.HasSuffix(rel, "/") { ⇒ dst = filepath.Join(t.Root, filepath.Dir(rel)) + "/" ⇒ } ⇒ if strings.HasSuffix(rel, "/") { ⇒ dst = filepath.Join(t.Root, fi` | killed-by TestY8Get/ignored-and-directory |
| mi26 | internal/mirror/mirror.go:112 | expression/error-guard | `if err != nil { ⇒ return nil, err ⇒ } ⇒ if false { ⇒ return nil, err ⇒ }` | test-added TestMirrorY4ListErrors/-co |
| mi27 | internal/mirror/mirror.go:116 | expression/error-guard | `if err != nil { ⇒ return nil, err ⇒ } ⇒ if false { ⇒ return nil, err ⇒ }` | test-added TestMirrorY4ListErrors/-i |
| mi28 | internal/mirror/mirror.go:270 | expression/error-guard | `if err := rsync(ctx, in.Bytes(), files...); err != nil { ⇒ return err ⇒ } ⇒ if err := rsync(ctx, in.Bytes(), files...); false { ⇒ return err ⇒ }` | test-added TestMirrorY10FileUploadFailure |
| mi29 | internal/mirror/mirror.go:44 | expression/errorf-wrap | `return Tree{}, fmt.Errorf("git does not recognize the repository at %s: %w", d, err) ⇒ return Tree{}, fmt.Errorf("git does not recognize the repository at %s: %` | contract-open Y1 requires failure + diagnostic, not error-chain identity; CLI only prints this error |
| mi30 | internal/mirror/mirror.go:250 | expression/errorf-wrap | `return fmt.Errorf("rsync: %w", err) ⇒ return fmt.Errorf("rsync: %v", err)` | contract-open Y10 requires rsync failure + diagnostic, not error-chain identity; caller only tests nonnil |
| mi31 | internal/mirror/mirror.go:135 | expression/errorf-wrap | `return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p) ⇒ return nil, fmt.Errorf("%v at %s", ErrLocalOnly, p)` | test-added TestMirrorY3NestedReplacesTracked |
| mi32 | internal/mirror/mirror.go:37 | expression/logical | `if !errors.As(err, &ge) \|\| !strings.Contains(ge.stderr, "not a git repository") { ⇒ if !errors.As(err, &ge) && !strings.Contains(ge.stderr, "not a git repositor` | test-added TestMirrorY1GitFailurePlainTree |
| mi33 | internal/mirror/mirror.go:75 | expression/logical | `if same, sameErr := sameFile(gitDir, dotgit); sameErr != nil \|\| !same { ⇒ if same, sameErr := sameFile(gitDir, dotgit); sameErr != nil && !same {` | test-added TestMirrorY3SplitGitDirs/both-outside |
| mi34 | internal/mirror/mirror.go:143 | expression/logical | `if fi, err := os.Lstat(filepath.Join(root, p)); err == nil && fi.Mode().IsRegular() { ⇒ if fi, err := os.Lstat(filepath.Join(root, p)); err == nil \|\| fi.Mode().` | test-added TestMirrorY4IgnoredSymlink |
| mi35 | internal/mirror/mirror.go:172 | expression/logical | `if !errors.As(err, &ge) \|\| ge.code != 1 { // 1 = unset ⇒ if !errors.As(err, &ge) && ge.code != 1 { // 1 = unset` | test-added TestMirrorY5ExcludesErrors/git-config |
| mi36 | internal/mirror/mirror.go:37 | expression/remove | `if !errors.As(err, &ge) \|\| !strings.Contains(ge.stderr, "not a git repository") { ⇒ if !errors.As(err, &ge) \|\| false {` | test-added TestMirrorY1GitFailurePlainTree |
| mi37 | internal/mirror/mirror.go:66 | expression/remove | `if errors.Is(err, fs.ErrNotExist) \|\| (err == nil && !fi.IsDir()) { ⇒ if false \|\| ⇒ (err == nil && !fi.IsDir()) {` | test-added TestMirrorY3SplitGitDirs/no-dotgit |
| mi38 | internal/mirror/mirror.go:66 | expression/remove | `if errors.Is(err, fs.ErrNotExist) \|\| (err == nil && !fi.IsDir()) { ⇒ if errors.Is(err, fs.ErrNotExist) \|\| (true && ⇒ !fi.IsDir()) {` | test-added TestMirrorY3UnreadableMetadata |
| mi39 | internal/mirror/mirror.go:75 | expression/remove | `if same, sameErr := sameFile(gitDir, dotgit); sameErr != nil \|\| !same { ⇒ if same, sameErr := sameFile(gitDir, dotgit); false \|\| ⇒ !same {` | equivalent sameFile error always returns same=false; !same already admits every sameErr |
| mi40 | internal/mirror/mirror.go:142 | expression/remove | `if filepath.Base(p) == ".gitignore" && !strings.HasSuffix(p, "/") { ⇒ if filepath.Base(p) == ".gitignore" && true {` | equivalent trailing-slash .gitignore is a directory; following IsRegular guard still excludes it |
| mi41 | internal/mirror/mirror.go:172 | expression/remove | `if !errors.As(err, &ge) \|\| ge.code != 1 { // 1 = unset ⇒ if !errors.As(err, &ge) \|\| false { // 1 = unset` | test-added TestMirrorY5ExcludesErrors/git-config |
| mi42 | internal/mirror/mirror.go:126 | loop/break | `if strings.HasSuffix(p, "/") { ⇒ return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p) ⇒ } ⇒ if !inTree(root, p, realDir) { ⇒ continue // reached through a symlin` | test-added TestMirrorY4AfterSymlinkAncestor |
| mi43 | internal/mirror/mirror.go:264 | loop/range_break | `in.WriteString(p) ⇒ in.WriteByte(0) ⇒ } ⇒ break ⇒ in.WriteString(p) ⇒ in.WriteByte(0) ⇒ }` | killed-by TestY1RootAndSubdirectory |
| mi44 | internal/mirror/mirror.go:59 | numbers/decrementer | `t.LocalOnly, err = localOnly(ctx, t.Root, f[1], f[2]) ⇒ t.LocalOnly, err = localOnly(ctx, t.Root, f[1], f[1])` | test-added TestMirrorY3SplitGitDirs/common-outside |
| mi45 | internal/mirror/mirror.go:59 | numbers/incrementer | `t.LocalOnly, err = localOnly(ctx, t.Root, f[1], f[2]) ⇒ t.LocalOnly, err = localOnly(ctx, t.Root, f[2], f[2])` | test-added TestMirrorY3SplitGitDirs/git-dir-outside |
| mi46 | internal/mirror/mirror.go:266 | numbers/incrementer | `in.WriteByte(0) ⇒ in.WriteByte(1)` | killed-by TestY1RootAndSubdirectory |
| mi47 | internal/mirror/mirror.go:57 | statement/remove | `t.Sub = "." ⇒ } ⇒ _ = t.Sub ⇒ }` | equivalent duplicate mi5; filepath.Clean cannot return empty |
| mi48 | internal/mirror/mirror.go:208 | numbers/incrementer | `if len(e) > 0 { ⇒ if len(e) > 1 {` | test-added TestMirrorY4AfterSymlinkAncestor |
| mi49 | internal/mirror/mirror.go:162 | statement/remove | `realDir[d] = ok ⇒ _, _, _ = realDir, d, ok` | equivalent memoization removed only; Y4 source-tree quiescence preserves repeated Lstat result |
| mi50 | internal/mirror/mirror.go:246 | statement/remove | `cmd.Stdin = bytes.NewReader(stdin) ⇒ } ⇒ _, _, _ = cmd.Stdin, bytes.NewReader, stdin ⇒ }` | killed-by TestY1RootAndSubdirectory |
| mi51 | internal/mirror/mirror.go:248 | statement/remove | `cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr ⇒ _, _, _, _ = cmd.Stdout, cmd.Stderr, os.Stderr, os.Stderr` | killed-by TestY10UploadFailure |
| mi52 | internal/mirror/mirror.go:266 | statement/remove | `in.WriteString(p) ⇒ in.WriteByte(0) ⇒ } ⇒ in.WriteString(p) ⇒ _ = in.WriteByte ⇒ }` | killed-by TestY1RootAndSubdirectory |
| mi53 | internal/mirror/mirror.go:265 | statement/remove | `in.WriteString(p) ⇒ in.WriteByte(0) ⇒ } ⇒ _, _ = in.WriteString, p ⇒ in.WriteByte(0) ⇒ }` | killed-by TestY1RootAndSubdirectory |
| mi54 | internal/mirror/mirror.go:305 | statement/remove | `dst = filepath.Join(t.Root, filepath.Dir(rel)) + "/" ⇒ } ⇒ _, _, _, _, _ = dst, filepath.Join, t.Root, filepath.Dir, rel ⇒ }` | killed-by TestY8Get/ignored-and-directory |
| mi55 | internal/mirror/mirror.go:38 | statement/return | `return Tree{}, err ⇒ } ⇒ return Tree{}, nil ⇒ }` | killed-by TestY1GitErrorIsNotPlainMode |
| mi56 | internal/mirror/mirror.go:60 | statement/return | `// git also says "not a git repository" for a damaged one: plain mode's --delete mirror ⇒ // must not run over a tree that has a .git anywhere above it. ⇒  ⇒ re` | test-added TestMirrorY1IndexFailure |
| mi57 | internal/mirror/mirror.go:73 | statement/return | `return "it is a linked worktree", nil ⇒ } ⇒ return "", nil ⇒ }` | test-added TestMirrorY3SplitGitDirs/common-outside |
| mi58 | internal/mirror/mirror.go:76 | statement/return | `return "its git dir lies outside the tree", sameErr ⇒ } ⇒ return "", sameErr ⇒ }` | test-added TestMirrorY3SplitGitDirs/both-outside |
| mi59 | internal/mirror/mirror.go:76 | statement/return | `return "its git dir lies outside the tree", sameErr ⇒ } ⇒ return "its git dir lies outside the tree", nil ⇒ }` | contract-open Y3: lookup error accompanies nonempty LocalOnly; clearing error preserves pool refusal, hybrid fatal-vs-local fallback unspecified |
| mi60 | internal/mirror/mirror.go:113 | statement/return | `return nil, err ⇒ } ⇒ return nil, nil ⇒ }` | test-added TestMirrorY4ListErrors/-co |
| mi61 | internal/mirror/mirror.go:117 | statement/return | `return nil, err ⇒ } ⇒ return nil, nil ⇒ }` | test-added TestMirrorY4ListErrors/-i |
| mi62 | internal/mirror/mirror.go:181 | statement/return | `if !errors.As(err, &ge) \|\| ge.code != 1 { // 1 = unset ⇒ if !errors.As(err, &ge) \|\| ge.code != 1 { ⇒ data, err := os.ReadFile(path) //nolint:gosec // G304: git'` | test-added TestMirrorY5ExcludesErrors/unreadable-file |
| mi63 | internal/mirror/mirror.go:173 | statement/return | `if !errors.As(err, &ge) \|\| ge.code != 1 { // 1 = unset ⇒ return "", err ⇒ } ⇒ if !errors.As(err, &ge) \|\| ge.code != 1 { ⇒ return "", nil ⇒ }` | test-added TestMirrorY5ExcludesErrors/git-config |
| mi64 | internal/mirror/mirror.go:256 | statement/return | `return []string{"-a", "-e", remote.RsyncShell(), "--rsync-path=" + rsyncPath(meta)} ⇒ return nil` | killed-by TestY1RootAndSubdirectory |
| mi65 | internal/mirror/mirror.go:275 | statement/return | `return rsync(ctx, nil, dotgit...) ⇒ _ = dotgit ⇒ return nil` | killed-by TestY1RootAndSubdirectory |
| mi66 | internal/mirror/mirror.go:271 | statement/return | `return err ⇒ } ⇒ return nil ⇒ }` | test-added TestMirrorY10FileUploadFailure |
| mi67 | internal/mirror/mirror.go:282 | statement/return | `return rsync(ctx, nil, args...) ⇒ _ = args ⇒ return nil` | killed-by TestY7PlainMirror |
| mi68 | internal/mirror/mirror.go:289 | statement/return | `return rsync(ctx, nil, args...) ⇒ _ = args ⇒ return nil` | killed-by TestY6Download/command-0 |
| mi69 | internal/mirror/mirror.go:296 | statement/return | `return rsync(ctx, nil, args...) ⇒ _ = args ⇒ return nil` | killed-by TestY7PlainMirror |
| mi70 | internal/mirror/mirror.go:308 | statement/return | `return rsync(ctx, nil, args...) ⇒ _ = args ⇒ return nil` | killed-by TestY9NoSync |

## probe (34 escaped, MSI 0.61)

| id | site | mutator | change | ruling |
|----|------|---------|--------|--------|
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
## remote (Mutago found no source files; hand-planted with `.agent/plant`)

| id | site | mutator | change | ruling |
|----|------|---------|--------|--------|
| re1 | internal/remote/ssh.go Quote | hand | `'`-escaping removed | killed-by TestJoinRoundTripsThroughBash |
| re2 | internal/remote/ssh.go Quote | hand | safe-char shortcut taken for every non-empty word | killed-by TestJoinRoundTripsThroughBash |
| re3 | internal/remote/ssh.go Quote | hand | empty word left unquoted | killed-by TestJoinRoundTripsThroughBash, TestX2ArgvAndEnvironment |
| re4 | internal/remote/ssh.go SSHArgs | hand | `--` before the destination dropped | equivalent: P2 refuses an address starting with `-` (TestLoadRejects `option as address`), so `--` never changes how ssh parses a valid address |
