# reviewer-6 — mirror survivors

verdict cell = killed-by <full-suite test> | test-added <name> | contract-open <predicate + why> | equivalent <why>. Finish = every mi row ruled; fixed set = mi1–mi70 × Y1–Y11 + Decisions. Verdict pass = evidenced kill or explicitly bounded acceptance ruling, not full product assurance. Base = e33029167b93f7c18469a165200806b9ff1c9f1b. Seed = .scratch/mut-mirror.json (70 entries); original 70 unknown rows confirmed before edits. Header originally contained the unknown-cell literal too (naive validator counted 71); header corrected, checker unchanged.

| id | site | mutator | change | ruling | verdict |
|----|------|---------|--------|--------|---------|
| mi1 | internal/mirror/mirror.go:135 | branch/case | `return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p) ⇒ _, _, _ = fmt.Errorf, ErrLocalOnly, p` | test-added TestMirrorY3NestedReplacesTracked | pass |
| mi2 | internal/mirror/mirror.go:133 | branch/case | `return nil, err ⇒ _ = err` | test-added TestMirrorY4LstatError | pass |
| mi3 | internal/mirror/mirror.go:38 | branch/if | `return Tree{}, err ⇒ } ⇒ _ = err ⇒ }` | test-added TestMirrorY1GitFailurePlainTree | pass |
| mi4 | internal/mirror/mirror.go:53 | branch/if | `return Tree{}, fmt.Errorf("git rev-parse printed %d lines, want 4", len(f)) ⇒ } ⇒ _, _ = fmt.Errorf, f ⇒ }` | test-added TestMirrorY1RevParseFraming | pass |
| mi5 | internal/mirror/mirror.go:57 | branch/if | `t.Sub = "." ⇒ } ⇒ _ = t.Sub ⇒ }` | equivalent filepath.Clean always returns nonempty; empty prefix already becomes "." (Y1) | pass |
| mi6 | internal/mirror/mirror.go:70 | branch/if | `return "", err ⇒ } ⇒ _ = err ⇒ }` | contract-open Y3: non-ENOENT .git lstat failure with split git dirs can become LocalOnly; fatal-vs-local fallback unspecified | pass |
| mi7 | internal/mirror/mirror.go:73 | branch/if | `return "it is a linked worktree", nil ⇒ } ⇒  ⇒ }` | test-added TestMirrorY3SplitGitDirs/common-outside | pass |
| mi8 | internal/mirror/mirror.go:76 | branch/if | `return "its git dir lies outside the tree", sameErr ⇒ } ⇒ _ = sameErr ⇒ }` | test-added TestMirrorY3SplitGitDirs/both-outside | pass |
| mi9 | internal/mirror/mirror.go:83 | branch/if | `return "", err ⇒ } ⇒ _ = err ⇒ }` | test-added TestMirrorY1IndexFailure | pass |
| mi10 | internal/mirror/mirror.go:96 | branch/if | `return false, err ⇒ } ⇒ _ = err ⇒ }` | contract-open Y3: git-dir Stat failure still yields LocalOnly (SameFile(nil, x)=false); hybrid fatal-vs-local fallback unspecified | pass |
| mi11 | internal/mirror/mirror.go:100 | branch/if | `return false, err ⇒ } ⇒ _ = err ⇒ }` | contract-open Y3: .git Stat failure still yields LocalOnly (SameFile(x, nil)=false); hybrid fatal-vs-local fallback unspecified | pass |
| mi12 | internal/mirror/mirror.go:113 | branch/if | `return nil, err ⇒ } ⇒ _ = err ⇒ }` | test-added TestMirrorY4ListErrors/-co | pass |
| mi13 | internal/mirror/mirror.go:117 | branch/if | `return nil, err ⇒ } ⇒ _ = err ⇒ }` | test-added TestMirrorY4ListErrors/-i | pass |
| mi14 | internal/mirror/mirror.go:123 | branch/if | `return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p) ⇒ } ⇒ _, _, _ = fmt.Errorf, ErrLocalOnly, p ⇒ }` | equivalent Y3/Y4: real-git trailing-/ entries name nested repo dirs; quiescent-tree Lstat.IsDir returns the same ErrLocalOnly | pass |
| mi15 | internal/mirror/mirror.go:173 | branch/if | `if !errors.As(err, &ge) \|\| ge.code != 1 { // 1 = unset ⇒ return "", err ⇒ } ⇒ if !errors.As(err, &ge) \|\| ge.code != 1 { ⇒ _ = err ⇒ }` | test-added TestMirrorY5ExcludesErrors/git-config | pass |
| mi16 | internal/mirror/mirror.go:246 | branch/if | `cmd.Stdin = bytes.NewReader(stdin) ⇒ } ⇒ _, _, _ = cmd.Stdin, bytes.NewReader, stdin ⇒ }` | killed-by TestY1RootAndSubdirectory | pass |
| mi17 | internal/mirror/mirror.go:250 | branch/if | `return fmt.Errorf("rsync: %w", err) ⇒ } ⇒ _, _ = fmt.Errorf, err ⇒ }` | killed-by TestY10DownloadFailureOverridesCommand/0 | pass |
| mi18 | internal/mirror/mirror.go:271 | branch/if | `return err ⇒ } ⇒ _ = err ⇒ }` | test-added TestMirrorY10FileUploadFailure | pass |
| mi19 | internal/mirror/mirror.go:305 | branch/if | `dst = filepath.Join(t.Root, filepath.Dir(rel)) + "/" ⇒ } ⇒ _, _, _, _, _ = dst, filepath.Join, t.Root, filepath.Dir, rel ⇒ }` | killed-by TestY8Get/ignored-and-directory | pass |
| mi20 | internal/mirror/mirror.go:233 | composite/field-clear | `return out, &gitError{args: args, code: ee.ExitCode(), stderr: stderr.String()} ⇒ return out, &gitError{code: ee.ExitCode(), stderr: stderr.String()}` | contract-open Y1/Y10 require tool diagnostic, not echoed git argv; stderr retained | pass |
| mi21 | internal/mirror/mirror.go:46 | conditional/negated | `if d == filepath.Dir(d) { ⇒ if d != filepath.Dir(d) {` | test-added TestMirrorY1DamagedAncestor | pass |
| mi22 | internal/mirror/mirror.go:249 | conditional/negated | `if err := cmd.Run(); err != nil { ⇒ if err := cmd.Run(); err == nil {` | killed-by TestY1RootAndSubdirectory | pass |
| mi23 | internal/mirror/mirror.go:245 | conditional/negated | `if stdin != nil { ⇒ if stdin == nil {` | killed-by TestY1RootAndSubdirectory | pass |
| mi24 | internal/mirror/mirror.go:270 | conditional/negated | `if err := rsync(ctx, in.Bytes(), files...); err != nil { ⇒ if err := rsync(ctx, in.Bytes(), files...); err == nil {` | killed-by TestY1RootAndSubdirectory | pass |
| mi25 | internal/mirror/mirror.go:304 | conditional/not | `if !strings.HasSuffix(rel, "/") { ⇒ dst = filepath.Join(t.Root, filepath.Dir(rel)) + "/" ⇒ } ⇒ if strings.HasSuffix(rel, "/") { ⇒ dst = filepath.Join(t.Root, fi` | killed-by TestY8Get/ignored-and-directory | pass |
| mi26 | internal/mirror/mirror.go:112 | expression/error-guard | `if err != nil { ⇒ return nil, err ⇒ } ⇒ if false { ⇒ return nil, err ⇒ }` | test-added TestMirrorY4ListErrors/-co | pass |
| mi27 | internal/mirror/mirror.go:116 | expression/error-guard | `if err != nil { ⇒ return nil, err ⇒ } ⇒ if false { ⇒ return nil, err ⇒ }` | test-added TestMirrorY4ListErrors/-i | pass |
| mi28 | internal/mirror/mirror.go:270 | expression/error-guard | `if err := rsync(ctx, in.Bytes(), files...); err != nil { ⇒ return err ⇒ } ⇒ if err := rsync(ctx, in.Bytes(), files...); false { ⇒ return err ⇒ }` | test-added TestMirrorY10FileUploadFailure | pass |
| mi29 | internal/mirror/mirror.go:44 | expression/errorf-wrap | `return Tree{}, fmt.Errorf("git does not recognize the repository at %s: %w", d, err) ⇒ return Tree{}, fmt.Errorf("git does not recognize the repository at %s: %` | contract-open Y1 requires failure + diagnostic, not error-chain identity; CLI only prints this error | pass |
| mi30 | internal/mirror/mirror.go:250 | expression/errorf-wrap | `return fmt.Errorf("rsync: %w", err) ⇒ return fmt.Errorf("rsync: %v", err)` | contract-open Y10 requires rsync failure + diagnostic, not error-chain identity; caller only tests nonnil | pass |
| mi31 | internal/mirror/mirror.go:135 | expression/errorf-wrap | `return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p) ⇒ return nil, fmt.Errorf("%v at %s", ErrLocalOnly, p)` | test-added TestMirrorY3NestedReplacesTracked | pass |
| mi32 | internal/mirror/mirror.go:37 | expression/logical | `if !errors.As(err, &ge) \|\| !strings.Contains(ge.stderr, "not a git repository") { ⇒ if !errors.As(err, &ge) && !strings.Contains(ge.stderr, "not a git repositor` | test-added TestMirrorY1GitFailurePlainTree | pass |
| mi33 | internal/mirror/mirror.go:75 | expression/logical | `if same, sameErr := sameFile(gitDir, dotgit); sameErr != nil \|\| !same { ⇒ if same, sameErr := sameFile(gitDir, dotgit); sameErr != nil && !same {` | test-added TestMirrorY3SplitGitDirs/both-outside | pass |
| mi34 | internal/mirror/mirror.go:143 | expression/logical | `if fi, err := os.Lstat(filepath.Join(root, p)); err == nil && fi.Mode().IsRegular() { ⇒ if fi, err := os.Lstat(filepath.Join(root, p)); err == nil \|\| fi.Mode().` | test-added TestMirrorY4IgnoredSymlink | pass |
| mi35 | internal/mirror/mirror.go:172 | expression/logical | `if !errors.As(err, &ge) \|\| ge.code != 1 { // 1 = unset ⇒ if !errors.As(err, &ge) && ge.code != 1 { // 1 = unset` | test-added TestMirrorY5ExcludesErrors/git-config | pass |
| mi36 | internal/mirror/mirror.go:37 | expression/remove | `if !errors.As(err, &ge) \|\| !strings.Contains(ge.stderr, "not a git repository") { ⇒ if !errors.As(err, &ge) \|\| false {` | test-added TestMirrorY1GitFailurePlainTree | pass |
| mi37 | internal/mirror/mirror.go:66 | expression/remove | `if errors.Is(err, fs.ErrNotExist) \|\| (err == nil && !fi.IsDir()) { ⇒ if false \|\| ⇒ (err == nil && !fi.IsDir()) {` | test-added TestMirrorY3SplitGitDirs/no-dotgit | pass |
| mi38 | internal/mirror/mirror.go:66 | expression/remove | `if errors.Is(err, fs.ErrNotExist) \|\| (err == nil && !fi.IsDir()) { ⇒ if errors.Is(err, fs.ErrNotExist) \|\| (true && ⇒ !fi.IsDir()) {` | test-added TestMirrorY3UnreadableMetadata | pass |
| mi39 | internal/mirror/mirror.go:75 | expression/remove | `if same, sameErr := sameFile(gitDir, dotgit); sameErr != nil \|\| !same { ⇒ if same, sameErr := sameFile(gitDir, dotgit); false \|\| ⇒ !same {` | equivalent sameFile error always returns same=false; !same already admits every sameErr | pass |
| mi40 | internal/mirror/mirror.go:142 | expression/remove | `if filepath.Base(p) == ".gitignore" && !strings.HasSuffix(p, "/") { ⇒ if filepath.Base(p) == ".gitignore" && true {` | equivalent trailing-slash .gitignore is a directory; following IsRegular guard still excludes it | pass |
| mi41 | internal/mirror/mirror.go:172 | expression/remove | `if !errors.As(err, &ge) \|\| ge.code != 1 { // 1 = unset ⇒ if !errors.As(err, &ge) \|\| false { // 1 = unset` | test-added TestMirrorY5ExcludesErrors/git-config | pass |
| mi42 | internal/mirror/mirror.go:126 | loop/break | `if strings.HasSuffix(p, "/") { ⇒ return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p) ⇒ } ⇒ if !inTree(root, p, realDir) { ⇒ continue // reached through a symlin` | test-added TestMirrorY4AfterSymlinkAncestor | pass |
| mi43 | internal/mirror/mirror.go:264 | loop/range_break | `in.WriteString(p) ⇒ in.WriteByte(0) ⇒ } ⇒ break ⇒ in.WriteString(p) ⇒ in.WriteByte(0) ⇒ }` | killed-by TestY1RootAndSubdirectory | pass |
| mi44 | internal/mirror/mirror.go:59 | numbers/decrementer | `t.LocalOnly, err = localOnly(ctx, t.Root, f[1], f[2]) ⇒ t.LocalOnly, err = localOnly(ctx, t.Root, f[1], f[1])` | test-added TestMirrorY3SplitGitDirs/common-outside | pass |
| mi45 | internal/mirror/mirror.go:59 | numbers/incrementer | `t.LocalOnly, err = localOnly(ctx, t.Root, f[1], f[2]) ⇒ t.LocalOnly, err = localOnly(ctx, t.Root, f[2], f[2])` | test-added TestMirrorY3SplitGitDirs/git-dir-outside | pass |
| mi46 | internal/mirror/mirror.go:266 | numbers/incrementer | `in.WriteByte(0) ⇒ in.WriteByte(1)` | killed-by TestY1RootAndSubdirectory | pass |
| mi47 | internal/mirror/mirror.go:57 | statement/remove | `t.Sub = "." ⇒ } ⇒ _ = t.Sub ⇒ }` | equivalent duplicate mi5; filepath.Clean cannot return empty | pass |
| mi48 | internal/mirror/mirror.go:208 | numbers/incrementer | `if len(e) > 0 { ⇒ if len(e) > 1 {` | test-added TestMirrorY4AfterSymlinkAncestor | pass |
| mi49 | internal/mirror/mirror.go:162 | statement/remove | `realDir[d] = ok ⇒ _, _, _ = realDir, d, ok` | equivalent memoization removed only; Y4 source-tree quiescence preserves repeated Lstat result | pass |
| mi50 | internal/mirror/mirror.go:246 | statement/remove | `cmd.Stdin = bytes.NewReader(stdin) ⇒ } ⇒ _, _, _ = cmd.Stdin, bytes.NewReader, stdin ⇒ }` | killed-by TestY1RootAndSubdirectory | pass |
| mi51 | internal/mirror/mirror.go:248 | statement/remove | `cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr ⇒ _, _, _, _ = cmd.Stdout, cmd.Stderr, os.Stderr, os.Stderr` | killed-by TestY10UploadFailure | pass |
| mi52 | internal/mirror/mirror.go:266 | statement/remove | `in.WriteString(p) ⇒ in.WriteByte(0) ⇒ } ⇒ in.WriteString(p) ⇒ _ = in.WriteByte ⇒ }` | killed-by TestY1RootAndSubdirectory | pass |
| mi53 | internal/mirror/mirror.go:265 | statement/remove | `in.WriteString(p) ⇒ in.WriteByte(0) ⇒ } ⇒ _, _ = in.WriteString, p ⇒ in.WriteByte(0) ⇒ }` | killed-by TestY1RootAndSubdirectory | pass |
| mi54 | internal/mirror/mirror.go:305 | statement/remove | `dst = filepath.Join(t.Root, filepath.Dir(rel)) + "/" ⇒ } ⇒ _, _, _, _, _ = dst, filepath.Join, t.Root, filepath.Dir, rel ⇒ }` | killed-by TestY8Get/ignored-and-directory | pass |
| mi55 | internal/mirror/mirror.go:38 | statement/return | `return Tree{}, err ⇒ } ⇒ return Tree{}, nil ⇒ }` | killed-by TestY1GitErrorIsNotPlainMode | pass |
| mi56 | internal/mirror/mirror.go:60 | statement/return | `// git also says "not a git repository" for a damaged one: plain mode's --delete mirror ⇒ // must not run over a tree that has a .git anywhere above it. ⇒  ⇒ re` | test-added TestMirrorY1IndexFailure | pass |
| mi57 | internal/mirror/mirror.go:73 | statement/return | `return "it is a linked worktree", nil ⇒ } ⇒ return "", nil ⇒ }` | test-added TestMirrorY3SplitGitDirs/common-outside | pass |
| mi58 | internal/mirror/mirror.go:76 | statement/return | `return "its git dir lies outside the tree", sameErr ⇒ } ⇒ return "", sameErr ⇒ }` | test-added TestMirrorY3SplitGitDirs/both-outside | pass |
| mi59 | internal/mirror/mirror.go:76 | statement/return | `return "its git dir lies outside the tree", sameErr ⇒ } ⇒ return "its git dir lies outside the tree", nil ⇒ }` | contract-open Y3: lookup error accompanies nonempty LocalOnly; clearing error preserves pool refusal, hybrid fatal-vs-local fallback unspecified | pass |
| mi60 | internal/mirror/mirror.go:113 | statement/return | `return nil, err ⇒ } ⇒ return nil, nil ⇒ }` | test-added TestMirrorY4ListErrors/-co | pass |
| mi61 | internal/mirror/mirror.go:117 | statement/return | `return nil, err ⇒ } ⇒ return nil, nil ⇒ }` | test-added TestMirrorY4ListErrors/-i | pass |
| mi62 | internal/mirror/mirror.go:181 | statement/return | `if !errors.As(err, &ge) \|\| ge.code != 1 { // 1 = unset ⇒ if !errors.As(err, &ge) \|\| ge.code != 1 { ⇒ data, err := os.ReadFile(path) //nolint:gosec // G304: git'` | test-added TestMirrorY5ExcludesErrors/unreadable-file | pass |
| mi63 | internal/mirror/mirror.go:173 | statement/return | `if !errors.As(err, &ge) \|\| ge.code != 1 { // 1 = unset ⇒ return "", err ⇒ } ⇒ if !errors.As(err, &ge) \|\| ge.code != 1 { ⇒ return "", nil ⇒ }` | test-added TestMirrorY5ExcludesErrors/git-config | pass |
| mi64 | internal/mirror/mirror.go:256 | statement/return | `return []string{"-a", "-e", remote.RsyncShell(), "--rsync-path=" + rsyncPath(meta)} ⇒ return nil` | killed-by TestY1RootAndSubdirectory | pass |
| mi65 | internal/mirror/mirror.go:275 | statement/return | `return rsync(ctx, nil, dotgit...) ⇒ _ = dotgit ⇒ return nil` | killed-by TestY1RootAndSubdirectory | pass |
| mi66 | internal/mirror/mirror.go:271 | statement/return | `return err ⇒ } ⇒ return nil ⇒ }` | test-added TestMirrorY10FileUploadFailure | pass |
| mi67 | internal/mirror/mirror.go:282 | statement/return | `return rsync(ctx, nil, args...) ⇒ _ = args ⇒ return nil` | killed-by TestY7PlainMirror | pass |
| mi68 | internal/mirror/mirror.go:289 | statement/return | `return rsync(ctx, nil, args...) ⇒ _ = args ⇒ return nil` | killed-by TestY6Download/command-0 | pass |
| mi69 | internal/mirror/mirror.go:296 | statement/return | `return rsync(ctx, nil, args...) ⇒ _ = args ⇒ return nil` | killed-by TestY7PlainMirror | pass |
| mi70 | internal/mirror/mirror.go:308 | statement/return | `return rsync(ctx, nil, args...) ⇒ _ = args ⇒ return nil` | killed-by TestY9NoSync | pass |

### mi16 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi16 internal/mirror/mirror.go '	cmd := remote.Command(ctx, "rsync", args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	if err := cmd.Run(); err != nil {
' '	cmd := remote.Command(ctx, "rsync", args...)
	if stdin != nil {
	_, _, _ = cmd.Stdin, bytes.NewReader, stdin
}
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	if err := cmd.Run(); err != nil {
' '^TestY1RootAndSubdirectory$'
```

`mi16: killed by --- FAIL: TestY1RootAndSubdirectory (0.15s)`

### mi17 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi17 internal/mirror/mirror.go '	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync: %w", err)
	}
	return nil
}
' '	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	if err := cmd.Run(); err != nil {
	_, _ = fmt.Errorf, err
}
	return nil
}
' '^TestY10DownloadFailureOverridesCommand/0$'
```

`mi17: killed by --- FAIL: TestY10DownloadFailureOverridesCommand (0.29s)`

### mi19 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi19 internal/mirror/mirror.go '	dst := filepath.Join(t.Root, rel) + "/"
	if !strings.HasSuffix(rel, "/") {
		dst = filepath.Join(t.Root, filepath.Dir(rel)) + "/"
	}
	args := append(base(t.Meta()), "--update", "--mkpath", address+":"+t.Workdir()+"/"+rel, dst)
	return rsync(ctx, nil, args...)
' '	dst := filepath.Join(t.Root, rel) + "/"
	if !strings.HasSuffix(rel, "/") {
	_, _, _, _, _ = dst, filepath.Join, t.Root, filepath.Dir, rel
}
	args := append(base(t.Meta()), "--update", "--mkpath", address+":"+t.Workdir()+"/"+rel, dst)
	return rsync(ctx, nil, args...)
' '^TestY8Get/ignored-and-directory$'
```

`mi19: killed by --- FAIL: TestY8GetKeepsNewerLocal (0.34s) --- FAIL: TestY8Get (0.37s)`

### mi22 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi22 internal/mirror/mirror.go '	}
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync: %w", err)
	}
' '	}
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	if err := cmd.Run(); err == nil {
		return fmt.Errorf("rsync: %w", err)
	}
' '^TestY1RootAndSubdirectory$'
```

`mi22: killed by --- FAIL: TestY1RootAndSubdirectory (0.08s)`

### mi23 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi23 internal/mirror/mirror.go 'func rsync(ctx context.Context, stdin []byte, args ...string) error {
	cmd := remote.Command(ctx, "rsync", args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
' 'func rsync(ctx context.Context, stdin []byte, args ...string) error {
	cmd := remote.Command(ctx, "rsync", args...)
	if stdin == nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
' '^TestY1RootAndSubdirectory$'
```

`mi23: killed by --- FAIL: TestY1RootAndSubdirectory (0.16s)`

### mi24 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi24 internal/mirror/mirror.go '	files := append(base(t.Meta()), "--mkpath", "--from0", "--files-from=-",
		t.Root+"/", address+":"+t.Workdir()+"/")
	if err := rsync(ctx, in.Bytes(), files...); err != nil {
		return err
	}
' '	files := append(base(t.Meta()), "--mkpath", "--from0", "--files-from=-",
		t.Root+"/", address+":"+t.Workdir()+"/")
	if err := rsync(ctx, in.Bytes(), files...); err == nil {
		return err
	}
' '^TestY1RootAndSubdirectory$'
```

`mi24: killed by --- FAIL: TestY1RootAndSubdirectory (0.09s)`

### mi25 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi25 internal/mirror/mirror.go 'func Get(ctx context.Context, t Tree, address, rel string) error {
	dst := filepath.Join(t.Root, rel) + "/"
	if !strings.HasSuffix(rel, "/") {
		dst = filepath.Join(t.Root, filepath.Dir(rel)) + "/"
	}
	args := append(base(t.Meta()), "--update", "--mkpath", address+":"+t.Workdir()+"/"+rel, dst)
	return rsync(ctx, nil, args...)
' 'func Get(ctx context.Context, t Tree, address, rel string) error {
	dst := filepath.Join(t.Root, rel) + "/"
	if strings.HasSuffix(rel, "/") {
	dst = filepath.Join(t.Root, filepath.Dir(rel)) + "/"
}
	args := append(base(t.Meta()), "--update", "--mkpath", address+":"+t.Workdir()+"/"+rel, dst)
	return rsync(ctx, nil, args...)
' '^TestY8Get/ignored-and-directory$'
```

`mi25: killed by --- FAIL: TestY8GetKeepsNewerLocal (0.30s) --- FAIL: TestY8Get (0.44s)`

### mi43 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi43 internal/mirror/mirror.go '	var in bytes.Buffer
	for _, p := range list {
		in.WriteString(p)
		in.WriteByte(0)
	}
	files := append(base(t.Meta()), "--mkpath", "--from0", "--files-from=-",
		t.Root+"/", address+":"+t.Workdir()+"/")
' '	var in bytes.Buffer
	for _, p := range list {
	break
	in.WriteString(p)
	in.WriteByte(0)
}
	files := append(base(t.Meta()), "--mkpath", "--from0", "--files-from=-",
		t.Root+"/", address+":"+t.Workdir()+"/")
' '^TestY1RootAndSubdirectory$'
```

`mi43: killed by --- FAIL: TestY1RootAndSubdirectory (0.15s)`

### mi46 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi46 internal/mirror/mirror.go '	for _, p := range list {
		in.WriteString(p)
		in.WriteByte(0)
	}
	files := append(base(t.Meta()), "--mkpath", "--from0", "--files-from=-",
' '	for _, p := range list {
		in.WriteString(p)
		in.WriteByte(1)
	}
	files := append(base(t.Meta()), "--mkpath", "--from0", "--files-from=-",
' '^TestY1RootAndSubdirectory$'
```

`mi46: killed by --- FAIL: TestY1RootAndSubdirectory (0.08s)`

### mi50 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi50 internal/mirror/mirror.go '	cmd := remote.Command(ctx, "rsync", args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	if err := cmd.Run(); err != nil {
' '	cmd := remote.Command(ctx, "rsync", args...)
	if stdin != nil {
	_, _, _ = cmd.Stdin, bytes.NewReader, stdin
}
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	if err := cmd.Run(); err != nil {
' '^TestY1RootAndSubdirectory$'
```

`mi50: killed by --- FAIL: TestY1RootAndSubdirectory (0.16s)`

### mi51 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi51 internal/mirror/mirror.go '		cmd.Stdin = bytes.NewReader(stdin)
	}
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync: %w", err)
' '		cmd.Stdin = bytes.NewReader(stdin)
	}
	_, _, _, _ = cmd.Stdout, cmd.Stderr, os.Stderr, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync: %w", err)
' '^TestY10UploadFailure$'
```

`mi51: killed by --- FAIL: TestY10UploadFailure (0.03s)`

### mi52 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi52 internal/mirror/mirror.go '	var in bytes.Buffer
	for _, p := range list {
		in.WriteString(p)
		in.WriteByte(0)
	}
	files := append(base(t.Meta()), "--mkpath", "--from0", "--files-from=-",
		t.Root+"/", address+":"+t.Workdir()+"/")
' '	var in bytes.Buffer
	for _, p := range list {
	in.WriteString(p)
	_ = in.WriteByte
}
	files := append(base(t.Meta()), "--mkpath", "--from0", "--files-from=-",
		t.Root+"/", address+":"+t.Workdir()+"/")
' '^TestY1RootAndSubdirectory$'
```

`mi52: killed by --- FAIL: TestY1RootAndSubdirectory (0.07s)`

### mi53 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi53 internal/mirror/mirror.go '	var in bytes.Buffer
	for _, p := range list {
		in.WriteString(p)
		in.WriteByte(0)
	}
	files := append(base(t.Meta()), "--mkpath", "--from0", "--files-from=-",
		t.Root+"/", address+":"+t.Workdir()+"/")
' '	var in bytes.Buffer
	for _, p := range list {
	_, _ = in.WriteString, p
	in.WriteByte(0)
}
	files := append(base(t.Meta()), "--mkpath", "--from0", "--files-from=-",
		t.Root+"/", address+":"+t.Workdir()+"/")
' '^TestY1RootAndSubdirectory$'
```

`mi53: killed by --- FAIL: TestY1RootAndSubdirectory (0.15s)`

### mi54 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi54 internal/mirror/mirror.go '	dst := filepath.Join(t.Root, rel) + "/"
	if !strings.HasSuffix(rel, "/") {
		dst = filepath.Join(t.Root, filepath.Dir(rel)) + "/"
	}
	args := append(base(t.Meta()), "--update", "--mkpath", address+":"+t.Workdir()+"/"+rel, dst)
	return rsync(ctx, nil, args...)
' '	dst := filepath.Join(t.Root, rel) + "/"
	if !strings.HasSuffix(rel, "/") {
	_, _, _, _, _ = dst, filepath.Join, t.Root, filepath.Dir, rel
}
	args := append(base(t.Meta()), "--update", "--mkpath", address+":"+t.Workdir()+"/"+rel, dst)
	return rsync(ctx, nil, args...)
' '^TestY8Get/ignored-and-directory$'
```

`mi54: killed by --- FAIL: TestY8GetKeepsNewerLocal (0.32s) --- FAIL: TestY8Get (0.40s)`

### mi55 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi55 internal/mirror/mirror.go '		var ge *gitError
		if !errors.As(err, &ge) || !strings.Contains(ge.stderr, "not a git repository") {
			return Tree{}, err
		}
		// git also says "not a git repository" for a damaged one: plain mode'"'"'s --delete mirror
		// must not run over a tree that has a .git anywhere above it.
' '		var ge *gitError
		if !errors.As(err, &ge) || !strings.Contains(ge.stderr, "not a git repository") {
	return Tree{}, nil
}
		// git also says "not a git repository" for a damaged one: plain mode'"'"'s --delete mirror
		// must not run over a tree that has a .git anywhere above it.
' '^TestY1GitErrorIsNotPlainMode$'
```

`mi55: killed by --- FAIL: TestY1GitErrorIsNotPlainMode (21.49s)`

### mi64 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi64 internal/mirror/mirror.go '
func base(meta string) []string {
	return []string{"-a", "-e", remote.RsyncShell(), "--rsync-path=" + rsyncPath(meta)}
}

' '
func base(meta string) []string {
	return nil
}

' '^TestY1RootAndSubdirectory$'
```

`mi64: killed by --- FAIL: TestY1RootAndSubdirectory (0.05s)`

### mi65 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi65 internal/mirror/mirror.go '	dotgit := append(base(t.Meta()), "--mkpath", "--delete", "--delete-excluded",
		"--filter=- /*.lock", "--filter=- /**/*.lock", t.Root+"/.git/", address+":"+t.Workdir()+"/.git/")
	return rsync(ctx, nil, dotgit...)
}

' '	dotgit := append(base(t.Meta()), "--mkpath", "--delete", "--delete-excluded",
		"--filter=- /*.lock", "--filter=- /**/*.lock", t.Root+"/.git/", address+":"+t.Workdir()+"/.git/")
	_ = dotgit
	return nil
}

' '^TestY1RootAndSubdirectory$'
```

`mi65: killed by --- FAIL: TestY1RootAndSubdirectory (0.12s)`

### mi67 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi67 internal/mirror/mirror.go '	args := append(base(t.Meta()), "--mkpath", "--delete", "--filter=:- .gitignore",
		t.Root+"/", address+":"+t.Workdir()+"/")
	return rsync(ctx, nil, args...)
}

' '	args := append(base(t.Meta()), "--mkpath", "--delete", "--filter=:- .gitignore",
		t.Root+"/", address+":"+t.Workdir()+"/")
	_ = args
	return nil
}

' '^TestY7PlainMirror$'
```

`mi67: killed by --- FAIL: TestY7PlainMirror (1.09s)`

### mi68 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi68 internal/mirror/mirror.go '	args := append(base(t.Meta()), "--from0", "--files-from=:"+t.Meta()+"/down", "--update",
		"--ignore-missing-args", address+":"+t.Workdir()+"/", t.Root+"/")
	return rsync(ctx, nil, args...)
}

' '	args := append(base(t.Meta()), "--from0", "--files-from=:"+t.Meta()+"/down", "--update",
		"--ignore-missing-args", address+":"+t.Workdir()+"/", t.Root+"/")
	_ = args
	return nil
}

' '^TestY6Download/command-0$'
```

`mi68: killed by --- FAIL: TestY6Download (0.91s)`

### mi69 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi69 internal/mirror/mirror.go '	args := append(base(t.Meta()), "--update", "--filter=:- .gitignore", "--exclude=/.git/",
		address+":"+t.Workdir()+"/", t.Root+"/")
	return rsync(ctx, nil, args...)
}

' '	args := append(base(t.Meta()), "--update", "--filter=:- .gitignore", "--exclude=/.git/",
		address+":"+t.Workdir()+"/", t.Root+"/")
	_ = args
	return nil
}

' '^TestY7PlainMirror$'
```

`mi69: killed by --- FAIL: TestY7PlainMirror (1.49s)`

### mi70 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi70 internal/mirror/mirror.go '	}
	args := append(base(t.Meta()), "--update", "--mkpath", address+":"+t.Workdir()+"/"+rel, dst)
	return rsync(ctx, nil, args...)
}

' '	}
	args := append(base(t.Meta()), "--update", "--mkpath", address+":"+t.Workdir()+"/"+rel, dst)
	_ = args
	return nil
}

' '^TestY9NoSync$'
```

`mi70: killed by --- FAIL: TestY9NoSync (0.79s)`

### mi1 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi1 internal/mirror/mirror.go '			return nil, err
		case fi.IsDir():
			return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p)
		}
		list = append(list, p)
' '			return nil, err
		case fi.IsDir():
	_, _, _ = fmt.Errorf, ErrLocalOnly, p
		}
		list = append(list, p)
' '^TestMirrorY3NestedReplacesTracked$'
```

`mi1: killed by --- FAIL: TestMirrorY3NestedReplacesTracked (0.02s)`

### mi2 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi2 internal/mirror/mirror.go '			continue // deleted, not yet staged
		case err != nil:
			return nil, err
		case fi.IsDir():
			return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p)
' '			continue // deleted, not yet staged
		case err != nil:
	_ = err
		case fi.IsDir():
			return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p)
' '^TestMirrorY4LstatError$'
```

`mi2: killed by --- FAIL: TestMirrorY4LstatError (0.01s)`

### mi3 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi3 internal/mirror/mirror.go '		var ge *gitError
		if !errors.As(err, &ge) || !strings.Contains(ge.stderr, "not a git repository") {
			return Tree{}, err
		}
		// git also says "not a git repository" for a damaged one: plain mode'"'"'s --delete mirror
		// must not run over a tree that has a .git anywhere above it.
' '		var ge *gitError
		if !errors.As(err, &ge) || !strings.Contains(ge.stderr, "not a git repository") {
	_ = err
}
		// git also says "not a git repository" for a damaged one: plain mode'"'"'s --delete mirror
		// must not run over a tree that has a .git anywhere above it.
' '^TestMirrorY1GitFailurePlainTree$'
```

`mi3: killed by --- FAIL: TestMirrorY1GitFailurePlainTree (0.01s)`

### mi4 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi4 internal/mirror/mirror.go '	f := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	if len(f) != 4 {
		return Tree{}, fmt.Errorf("git rev-parse printed %d lines, want 4", len(f))
	}
	t := Tree{Root: f[0], Sub: filepath.Clean(strings.TrimSuffix(f[3], "/")), Git: true}
	if t.Sub == "" {
' '	f := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	if len(f) != 4 {
	_, _ = fmt.Errorf, f
}
	t := Tree{Root: f[0], Sub: filepath.Clean(strings.TrimSuffix(f[3], "/")), Git: true}
	if t.Sub == "" {
' '^TestMirrorY1RevParseFraming$'
```

`mi4: killed by --- FAIL: TestMirrorY1RevParseFraming (0.02s)`

### mi7 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi7 internal/mirror/mirror.go '	}
	if gitDir != commonDir {
		return "it is a linked worktree", nil
	}
	if same, sameErr := sameFile(gitDir, dotgit); sameErr != nil || !same {
		return "its git dir lies outside the tree", sameErr
' '	}
	if gitDir != commonDir {

}
	if same, sameErr := sameFile(gitDir, dotgit); sameErr != nil || !same {
		return "its git dir lies outside the tree", sameErr
' '^TestMirrorY3SplitGitDirs$/^common-outside$'
```

`mi7: killed by --- FAIL: TestMirrorY3SplitGitDirs (0.02s)`

### mi8 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi8 internal/mirror/mirror.go '	}
	if same, sameErr := sameFile(gitDir, dotgit); sameErr != nil || !same {
		return "its git dir lies outside the tree", sameErr
	}
	if _, statErr := os.Stat(filepath.Join(commonDir, "objects", "info", "alternates")); statErr == nil {
		return "it borrows objects from another repository (alternates)", nil
' '	}
	if same, sameErr := sameFile(gitDir, dotgit); sameErr != nil || !same {
	_ = sameErr
}
	if _, statErr := os.Stat(filepath.Join(commonDir, "objects", "info", "alternates")); statErr == nil {
		return "it borrows objects from another repository (alternates)", nil
' '^TestMirrorY3SplitGitDirs$/^both-outside$'
```

`mi8: killed by --- FAIL: TestMirrorY3SplitGitDirs (0.01s)`

### mi9 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi9 internal/mirror/mirror.go '	staged, err := git(ctx, root, "ls-files", "-z", "--stage")
	if err != nil {
		return "", err
	}
	for _, e := range bytes.Split(staged, []byte{0}) {
		if bytes.HasPrefix(e, []byte("160000 ")) {
' '	staged, err := git(ctx, root, "ls-files", "-z", "--stage")
	if err != nil {
	_ = err
}
	for _, e := range bytes.Split(staged, []byte{0}) {
		if bytes.HasPrefix(e, []byte("160000 ")) {
' '^TestMirrorY1IndexFailure$'
```

`mi9: killed by --- FAIL: TestMirrorY1IndexFailure (0.02s)`

### mi12 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi12 internal/mirror/mirror.go '	all, err := git(ctx, root, "ls-files", "-z", "-co", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	ignored, err := git(ctx, root, "ls-files", "-z", "-o", "-i", "--exclude-standard", "--directory")
	if err != nil {
' '	all, err := git(ctx, root, "ls-files", "-z", "-co", "--exclude-standard")
	if err != nil {
	_ = err
}
	ignored, err := git(ctx, root, "ls-files", "-z", "-o", "-i", "--exclude-standard", "--directory")
	if err != nil {
' '^TestMirrorY4ListErrors$/^-co$'
```

`mi12: killed by --- FAIL: TestMirrorY4ListErrors (0.04s)`

### mi13 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi13 internal/mirror/mirror.go '	ignored, err := git(ctx, root, "ls-files", "-z", "-o", "-i", "--exclude-standard", "--directory")
	if err != nil {
		return nil, err
	}
	var list []string
	realDir := map[string]bool{".": true}
' '	ignored, err := git(ctx, root, "ls-files", "-z", "-o", "-i", "--exclude-standard", "--directory")
	if err != nil {
	_ = err
}
	var list []string
	realDir := map[string]bool{".": true}
' '^TestMirrorY4ListErrors$/^-i$'
```

`mi13: killed by --- FAIL: TestMirrorY4ListErrors (0.04s)`

### mi15 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi15 internal/mirror/mirror.go '	if err != nil {
		var ge *gitError
		if !errors.As(err, &ge) || ge.code != 1 { // 1 = unset
			return "", err
		}
		path = defaultExcludes()
	}
' '	if err != nil {
		var ge *gitError
		if !errors.As(err, &ge) || ge.code != 1 {
	_ = err
}
		path = defaultExcludes()
	}
' '^TestMirrorY5ExcludesErrors$/^git-config$'
```

`mi15: killed by --- FAIL: TestMirrorY5ExcludesErrors (0.03s)`

### mi18 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi18 internal/mirror/mirror.go '		t.Root+"/", address+":"+t.Workdir()+"/")
	if err := rsync(ctx, in.Bytes(), files...); err != nil {
		return err
	}
	dotgit := append(base(t.Meta()), "--mkpath", "--delete", "--delete-excluded",
		"--filter=- /*.lock", "--filter=- /**/*.lock", t.Root+"/.git/", address+":"+t.Workdir()+"/.git/")
' '		t.Root+"/", address+":"+t.Workdir()+"/")
	if err := rsync(ctx, in.Bytes(), files...); err != nil {
	_ = err
}
	dotgit := append(base(t.Meta()), "--mkpath", "--delete", "--delete-excluded",
		"--filter=- /*.lock", "--filter=- /**/*.lock", t.Root+"/.git/", address+":"+t.Workdir()+"/.git/")
' '^TestMirrorY10FileUploadFailure$'
```

`mi18: killed by --- FAIL: TestMirrorY10FileUploadFailure (1.81s)`

### mi21 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi21 internal/mirror/mirror.go '				return Tree{}, fmt.Errorf("git does not recognize the repository at %s: %w", d, err)
			}
			if d == filepath.Dir(d) {
				return Tree{Root: cwd, Sub: "."}, nil
			}
' '				return Tree{}, fmt.Errorf("git does not recognize the repository at %s: %w", d, err)
			}
			if d != filepath.Dir(d) {
				return Tree{Root: cwd, Sub: "."}, nil
			}
' '^TestMirrorY1DamagedAncestor$'
```

`mi21: killed by --- FAIL: TestMirrorY1DamagedAncestor (0.03s)`

### mi26 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi26 internal/mirror/mirror.go 'func UploadList(ctx context.Context, root string) ([]string, error) {
	all, err := git(ctx, root, "ls-files", "-z", "-co", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	ignored, err := git(ctx, root, "ls-files", "-z", "-o", "-i", "--exclude-standard", "--directory")
	if err != nil {
' 'func UploadList(ctx context.Context, root string) ([]string, error) {
	all, err := git(ctx, root, "ls-files", "-z", "-co", "--exclude-standard")
	if false {
	return nil, err
}
	ignored, err := git(ctx, root, "ls-files", "-z", "-o", "-i", "--exclude-standard", "--directory")
	if err != nil {
' '^TestMirrorY4ListErrors$/^-co$'
```

`mi26: killed by --- FAIL: TestMirrorY4ListErrors (0.04s)`

### mi27 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi27 internal/mirror/mirror.go '	}
	ignored, err := git(ctx, root, "ls-files", "-z", "-o", "-i", "--exclude-standard", "--directory")
	if err != nil {
		return nil, err
	}
	var list []string
	realDir := map[string]bool{".": true}
' '	}
	ignored, err := git(ctx, root, "ls-files", "-z", "-o", "-i", "--exclude-standard", "--directory")
	if false {
	return nil, err
}
	var list []string
	realDir := map[string]bool{".": true}
' '^TestMirrorY4ListErrors$/^-i$'
```

`mi27: killed by --- FAIL: TestMirrorY4ListErrors (0.06s)`

### mi28 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi28 internal/mirror/mirror.go '	files := append(base(t.Meta()), "--mkpath", "--from0", "--files-from=-",
		t.Root+"/", address+":"+t.Workdir()+"/")
	if err := rsync(ctx, in.Bytes(), files...); err != nil {
		return err
	}
	dotgit := append(base(t.Meta()), "--mkpath", "--delete", "--delete-excluded",
		"--filter=- /*.lock", "--filter=- /**/*.lock", t.Root+"/.git/", address+":"+t.Workdir()+"/.git/")
' '	files := append(base(t.Meta()), "--mkpath", "--from0", "--files-from=-",
		t.Root+"/", address+":"+t.Workdir()+"/")
	if err := rsync(ctx, in.Bytes(), files...); false {
	return err
}
	dotgit := append(base(t.Meta()), "--mkpath", "--delete", "--delete-excluded",
		"--filter=- /*.lock", "--filter=- /**/*.lock", t.Root+"/.git/", address+":"+t.Workdir()+"/.git/")
' '^TestMirrorY10FileUploadFailure$'
```

`mi28: killed by --- FAIL: TestMirrorY10FileUploadFailure (1.61s)`

### mi31 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi31 internal/mirror/mirror.go '			return nil, err
		case fi.IsDir():
			return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p)
		}
		list = append(list, p)
' '			return nil, err
		case fi.IsDir():
			return nil, fmt.Errorf("%v at %s", ErrLocalOnly, p)
		}
		list = append(list, p)
' '^TestMirrorY3NestedReplacesTracked$'
```

`mi31: killed by --- FAIL: TestMirrorY3NestedReplacesTracked (0.05s)`

### mi32 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi32 internal/mirror/mirror.go '	if err != nil {
		var ge *gitError
		if !errors.As(err, &ge) || !strings.Contains(ge.stderr, "not a git repository") {
			return Tree{}, err
		}
' '	if err != nil {
		var ge *gitError
		if !errors.As(err, &ge) && !strings.Contains(ge.stderr, "not a git repository") {
			return Tree{}, err
		}
' '^TestMirrorY1GitFailurePlainTree$'
```

`mi32: killed by --- FAIL: TestMirrorY1GitFailurePlainTree (0.02s)`

### mi33 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi33 internal/mirror/mirror.go '		return "it is a linked worktree", nil
	}
	if same, sameErr := sameFile(gitDir, dotgit); sameErr != nil || !same {
		return "its git dir lies outside the tree", sameErr
	}
' '		return "it is a linked worktree", nil
	}
	if same, sameErr := sameFile(gitDir, dotgit); sameErr != nil && !same {
		return "its git dir lies outside the tree", sameErr
	}
' '^TestMirrorY3SplitGitDirs$/^both-outside$'
```

`mi33: killed by --- FAIL: TestMirrorY3SplitGitDirs (0.06s)`

### mi34 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi34 internal/mirror/mirror.go '		// non-ignored directory, where git still reads it.
		if filepath.Base(p) == ".gitignore" && !strings.HasSuffix(p, "/") {
			if fi, err := os.Lstat(filepath.Join(root, p)); err == nil && fi.Mode().IsRegular() {
				list = append(list, p)
			}
' '		// non-ignored directory, where git still reads it.
		if filepath.Base(p) == ".gitignore" && !strings.HasSuffix(p, "/") {
			if fi, err := os.Lstat(filepath.Join(root, p)); err == nil || fi.Mode().IsRegular() {
				list = append(list, p)
			}
' '^TestMirrorY4IgnoredSymlink$'
```

`mi34: killed by --- FAIL: TestMirrorY4IgnoredSymlink (0.03s)`

### mi35 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi35 internal/mirror/mirror.go '	if err != nil {
		var ge *gitError
		if !errors.As(err, &ge) || ge.code != 1 { // 1 = unset
			return "", err
		}
' '	if err != nil {
		var ge *gitError
		if !errors.As(err, &ge) && ge.code != 1 { // 1 = unset
			return "", err
		}
' '^TestMirrorY5ExcludesErrors$/^git-config$'
```

`mi35: killed by --- FAIL: TestMirrorY5ExcludesErrors (0.01s)`

### mi36 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi36 internal/mirror/mirror.go '	if err != nil {
		var ge *gitError
		if !errors.As(err, &ge) || !strings.Contains(ge.stderr, "not a git repository") {
			return Tree{}, err
		}
' '	if err != nil {
		var ge *gitError
		if !errors.As(err, &ge) || false {
			return Tree{}, err
		}
' '^TestMirrorY1GitFailurePlainTree$'
```

`mi36: killed by --- FAIL: TestMirrorY1GitFailurePlainTree (0.00s)`

### mi37 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi37 internal/mirror/mirror.go '	dotgit := filepath.Join(root, ".git")
	fi, err := os.Lstat(dotgit)
	if errors.Is(err, fs.ErrNotExist) || (err == nil && !fi.IsDir()) {
		return "its .git is not a directory (linked worktree or external git dir)", nil
	}
' '	dotgit := filepath.Join(root, ".git")
	fi, err := os.Lstat(dotgit)
	if false ||
	(err == nil && !fi.IsDir()) {
		return "its .git is not a directory (linked worktree or external git dir)", nil
	}
' '^TestMirrorY3SplitGitDirs$/^no-dotgit$'
```

`mi37: killed by --- FAIL: TestMirrorY3SplitGitDirs (0.02s)`

### mi38 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi38 internal/mirror/mirror.go '	dotgit := filepath.Join(root, ".git")
	fi, err := os.Lstat(dotgit)
	if errors.Is(err, fs.ErrNotExist) || (err == nil && !fi.IsDir()) {
		return "its .git is not a directory (linked worktree or external git dir)", nil
	}
' '	dotgit := filepath.Join(root, ".git")
	fi, err := os.Lstat(dotgit)
	if errors.Is(err, fs.ErrNotExist) || (true &&
	!fi.IsDir()) {
		return "its .git is not a directory (linked worktree or external git dir)", nil
	}
' '^TestMirrorY3UnreadableMetadata$'
```

`mi38: killed by --- FAIL: TestMirrorY3UnreadableMetadata (0.01s)`

### mi41 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi41 internal/mirror/mirror.go '	if err != nil {
		var ge *gitError
		if !errors.As(err, &ge) || ge.code != 1 { // 1 = unset
			return "", err
		}
' '	if err != nil {
		var ge *gitError
		if !errors.As(err, &ge) || false { // 1 = unset
			return "", err
		}
' '^TestMirrorY5ExcludesErrors$/^git-config$'
```

`mi41: killed by --- FAIL: TestMirrorY5ExcludesErrors (0.03s)`

### mi42 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi42 internal/mirror/mirror.go '	realDir := map[string]bool{".": true}
	for _, p := range splitZ(all) {
		if strings.HasSuffix(p, "/") {
			return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p)
		}
		if !inTree(root, p, realDir) {
			continue // reached through a symlink: not part of this tree
		}
		fi, err := os.Lstat(filepath.Join(root, p))
		switch {
		case errors.Is(err, fs.ErrNotExist):
			continue // deleted, not yet staged
		case err != nil:
			return nil, err
		case fi.IsDir():
			return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p)
		}
		list = append(list, p)
	}
	for _, p := range splitZ(ignored) {
		// Ignored directories arrive collapsed ("dir/") ⇒ a bare .gitignore entry sits in a
' '	realDir := map[string]bool{".": true}
	for _, p := range splitZ(all) {
	if strings.HasSuffix(p, "/") {
		return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p)
	}
	if !inTree(root, p, realDir) {
		break
	}
	fi, err := os.Lstat(filepath.Join(root, p))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		continue
	case err != nil:
		return nil, err
	case fi.IsDir():
		return nil, fmt.Errorf("%w at %s", ErrLocalOnly, p)
	}
	list = append(list, p)
}
	for _, p := range splitZ(ignored) {
		// Ignored directories arrive collapsed ("dir/") ⇒ a bare .gitignore entry sits in a
' '^TestMirrorY4AfterSymlinkAncestor$'
```

`mi42: killed by --- FAIL: TestMirrorY4AfterSymlinkAncestor (0.04s)`

### mi44 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi44 internal/mirror/mirror.go '		t.Sub = "."
	}
	t.LocalOnly, err = localOnly(ctx, t.Root, f[1], f[2])
	return t, err
}
' '		t.Sub = "."
	}
	t.LocalOnly, err = localOnly(ctx, t.Root, f[1], f[1])
	return t, err
}
' '^TestMirrorY3SplitGitDirs$/^common-outside$'
```

`mi44: killed by --- FAIL: TestMirrorY3SplitGitDirs (0.02s)`

### mi45 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi45 internal/mirror/mirror.go '		t.Sub = "."
	}
	t.LocalOnly, err = localOnly(ctx, t.Root, f[1], f[2])
	return t, err
}
' '		t.Sub = "."
	}
	t.LocalOnly, err = localOnly(ctx, t.Root, f[2], f[2])
	return t, err
}
' '^TestMirrorY3SplitGitDirs$/^git-dir-outside$'
```

`mi45: killed by --- FAIL: TestMirrorY3SplitGitDirs (0.01s)`

### mi48 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi48 internal/mirror/mirror.go '	var out []string
	for _, e := range bytes.Split(b, []byte{0}) {
		if len(e) > 0 {
			out = append(out, string(e))
		}
' '	var out []string
	for _, e := range bytes.Split(b, []byte{0}) {
		if len(e) > 1 {
			out = append(out, string(e))
		}
' '^TestMirrorY4AfterSymlinkAncestor$'
```

`mi48: killed by --- FAIL: TestMirrorY4AfterSymlinkAncestor (0.02s)`

### mi56 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi56 internal/mirror/mirror.go '			return Tree{}, err
		}
		// git also says "not a git repository" for a damaged one: plain mode'"'"'s --delete mirror
		// must not run over a tree that has a .git anywhere above it.
		for d := cwd; ; d = filepath.Dir(d) {
			if _, statErr := os.Lstat(filepath.Join(d, ".git")); statErr == nil {
				return Tree{}, fmt.Errorf("git does not recognize the repository at %s: %w", d, err)
			}
			if d == filepath.Dir(d) {
				return Tree{Root: cwd, Sub: "."}, nil
			}
		}
	}
	f := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	if len(f) != 4 {
		return Tree{}, fmt.Errorf("git rev-parse printed %d lines, want 4", len(f))
	}
	t := Tree{Root: f[0], Sub: filepath.Clean(strings.TrimSuffix(f[3], "/")), Git: true}
	if t.Sub == "" {
		t.Sub = "."
	}
	t.LocalOnly, err = localOnly(ctx, t.Root, f[1], f[2])
	return t, err
}

' '			return Tree{}, err
		}

		for d := cwd; ; d = filepath.Dir(d) {
			if _, statErr := os.Lstat(filepath.Join(d, ".git")); statErr == nil {
				return Tree{}, fmt.Errorf("git does not recognize the repository at %s: %w", d, err)
			}
			if d == filepath.Dir(d) {
				return Tree{Root: cwd, Sub: "."}, nil
			}
		}
	}
	f := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	if len(f) != 4 {
		return Tree{}, fmt.Errorf("git rev-parse printed %d lines, want 4", len(f))
	}
	t := Tree{Root: f[0], Sub: filepath.Clean(strings.TrimSuffix(f[3], "/")), Git: true}
	if t.Sub == "" {
		t.Sub = "."
	}
	t.LocalOnly, err = localOnly(ctx, t.Root, f[1], f[2])
	return t, nil
}

' '^TestMirrorY1IndexFailure$'
```

`mi56: killed by --- FAIL: TestMirrorY1IndexFailure (0.01s)`

### mi57 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi57 internal/mirror/mirror.go '	}
	if gitDir != commonDir {
		return "it is a linked worktree", nil
	}
	if same, sameErr := sameFile(gitDir, dotgit); sameErr != nil || !same {
		return "its git dir lies outside the tree", sameErr
' '	}
	if gitDir != commonDir {
	return "", nil
}
	if same, sameErr := sameFile(gitDir, dotgit); sameErr != nil || !same {
		return "its git dir lies outside the tree", sameErr
' '^TestMirrorY3SplitGitDirs$/^common-outside$'
```

`mi57: killed by --- FAIL: TestMirrorY3SplitGitDirs (0.01s)`

### mi58 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi58 internal/mirror/mirror.go '	}
	if same, sameErr := sameFile(gitDir, dotgit); sameErr != nil || !same {
		return "its git dir lies outside the tree", sameErr
	}
	if _, statErr := os.Stat(filepath.Join(commonDir, "objects", "info", "alternates")); statErr == nil {
		return "it borrows objects from another repository (alternates)", nil
' '	}
	if same, sameErr := sameFile(gitDir, dotgit); sameErr != nil || !same {
	return "", sameErr
}
	if _, statErr := os.Stat(filepath.Join(commonDir, "objects", "info", "alternates")); statErr == nil {
		return "it borrows objects from another repository (alternates)", nil
' '^TestMirrorY3SplitGitDirs$/^both-outside$'
```

`mi58: killed by --- FAIL: TestMirrorY3SplitGitDirs (0.02s)`

### mi60 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi60 internal/mirror/mirror.go '	all, err := git(ctx, root, "ls-files", "-z", "-co", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	ignored, err := git(ctx, root, "ls-files", "-z", "-o", "-i", "--exclude-standard", "--directory")
	if err != nil {
' '	all, err := git(ctx, root, "ls-files", "-z", "-co", "--exclude-standard")
	if err != nil {
	return nil, nil
}
	ignored, err := git(ctx, root, "ls-files", "-z", "-o", "-i", "--exclude-standard", "--directory")
	if err != nil {
' '^TestMirrorY4ListErrors$/^-co$'
```

`mi60: killed by --- FAIL: TestMirrorY4ListErrors (0.01s)`

### mi61 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi61 internal/mirror/mirror.go '	ignored, err := git(ctx, root, "ls-files", "-z", "-o", "-i", "--exclude-standard", "--directory")
	if err != nil {
		return nil, err
	}
	var list []string
	realDir := map[string]bool{".": true}
' '	ignored, err := git(ctx, root, "ls-files", "-z", "-o", "-i", "--exclude-standard", "--directory")
	if err != nil {
	return nil, nil
}
	var list []string
	realDir := map[string]bool{".": true}
' '^TestMirrorY4ListErrors$/^-i$'
```

`mi61: killed by --- FAIL: TestMirrorY4ListErrors (0.01s)`

### mi62 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi62 internal/mirror/mirror.go '	if err != nil {
		var ge *gitError
		if !errors.As(err, &ge) || ge.code != 1 { // 1 = unset
			return "", err
		}
		path = defaultExcludes()
	}
	data, err := os.ReadFile(path) //nolint:gosec // G304: git'"'"'s own excludes path
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	return string(data), err
}

' '	if err != nil {
		var ge *gitError
		if !errors.As(err, &ge) || ge.code != 1 {
			return "", err
		}
		path = defaultExcludes()
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	return string(data), nil
}

' '^TestMirrorY5ExcludesErrors$/^unreadable-file$'
```

`mi62: killed by --- FAIL: TestMirrorY5ExcludesErrors (0.01s)`

### mi63 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi63 internal/mirror/mirror.go '	if err != nil {
		var ge *gitError
		if !errors.As(err, &ge) || ge.code != 1 { // 1 = unset
			return "", err
		}
		path = defaultExcludes()
	}
' '	if err != nil {
		var ge *gitError
		if !errors.As(err, &ge) || ge.code != 1 {
	return "", nil
}
		path = defaultExcludes()
	}
' '^TestMirrorY5ExcludesErrors$/^git-config$'
```

`mi63: killed by --- FAIL: TestMirrorY5ExcludesErrors (0.01s)`

### mi66 — plant proof

```sh
/home/eturkes/Projects/offload/.agent/plant /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-6 mi66 internal/mirror/mirror.go '		t.Root+"/", address+":"+t.Workdir()+"/")
	if err := rsync(ctx, in.Bytes(), files...); err != nil {
		return err
	}
	dotgit := append(base(t.Meta()), "--mkpath", "--delete", "--delete-excluded",
		"--filter=- /*.lock", "--filter=- /**/*.lock", t.Root+"/.git/", address+":"+t.Workdir()+"/.git/")
' '		t.Root+"/", address+":"+t.Workdir()+"/")
	if err := rsync(ctx, in.Bytes(), files...); err != nil {
	return nil
}
	dotgit := append(base(t.Meta()), "--mkpath", "--delete", "--delete-excluded",
		"--filter=- /*.lock", "--filter=- /**/*.lock", t.Root+"/.git/", address+":"+t.Workdir()+"/.git/")
' '^TestMirrorY10FileUploadFailure$'
```

`mi66: killed by --- FAIL: TestMirrorY10FileUploadFailure (0.52s)`

## Acceptance bounds

- 70/70 adjudicated: 21 existing-test kills; 36 added-test kills; 6 equivalent; 7 contract-open. Each behavioral plant below ran through `.agent/plant`, then production bytes were compared to the pre-plant source.
- mi6/mi10/mi11/mi59: no ruling that metadata lookup errors are harmless. Y3 explicitly specifies unsupported layouts; it does not separately specify the hybrid response to failed filesystem metadata lookup. Each mutant still refuses pool placement in the error branch at issue; hybrid error-vs-local fallback remains open.
- mi20/mi29/mi30: CLI diagnostics preserve tool text; subcommand echo and Go error-chain identity are not contract predicates.
- mi14: source-tree quiescence + real git output are required premises; synthetic trailing-slash non-directories or concurrent removal are not an equivalence claim.
- LSP unavailable for .go in this session; Go compilation/tests + pinned golangci-lint replace editor diagnostics.

## Verification

- Seed SHA256 (`.scratch/mut-mirror.json`) = `3abd73b41c9393ac8262d3e8c259eaa185cc73107e5bb5b322f0cb67db15f176`.
- Baseline selected existing e2e predicates: `go test -count=1 -run '^Test(Y1GitErrorIsNotPlainMode|Y1RootAndSubdirectory|Y3UnsupportedGitLayouts|Y4GitUploadDifferential|Y6Download|Y7PlainMirror|Y8Get|Y9NoSync|Y10UploadFailure|Y10DownloadFailureOverridesCommand|ReviewY8GetExistingDirectory)$' ./...` → rc0, e2e 5.999s.
- Added-test baseline: `go test -count=1 -run '^TestMirror' ./internal/mirror ./test/e2e` → rc0 (mirror 0.287s; e2e 1.721s). All 13 added tests have a red plant recorded above.
- Pinned lint initial attempt → rc1: `redefines-builtin-id: redefinition of the built-in function real` in both added files. Renamed bindings `realGit`/`realRsync`; final rerun passed below.
- `git diff --exit-code -- internal/mirror/mirror.go` → rc0 after all 57 killed plants; source unchanged.
- Structural validator: ids exactly mi1–mi70, all pass; 57/57 killed/test-added rows have explicit plant command + FAIL output. Brief's grep-count validator → 0, rc1 (zero matches), positive control before edits = 70 target rows.
- Final pinned `golangci-lint fmt --diff` + `golangci-lint run ./internal/mirror/... ./test/e2e/...` → rc0, 0 issues. `offload run --local -- go test -p 1 -race -shuffle=on -count=1 ./internal/mirror ./test/e2e` → rc0 (mirror 1.252s; e2e 49.866s); initially waited for both local slots, then acquired slot1.
- Full project gate not run in this review; scope = mirror mutants + added mirror/e2e tests. Live pool hosts not used.
- Test commit = `7f7adc5cb5a1e91f42085f4166771ea08eb41dd8` on `wt/reviewer-6`; two test files only, 241 added lines. Commit body preserves all 36 added-test plant commands + FAIL output. Worktree clean; production untouched.

reviewer-6-DONE-1
