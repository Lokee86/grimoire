# Release packaging

`tools/package_release.py` creates a clean distribution directory with the Lexicon executable and all adapter runtime files:

```bash
python tools/package_release.py --output release
release/lexicon init --repo /path/to/repository
```

The distribution contains `adapters/go/lexicon-go`, `adapters/gdscript/lexicon-gdscript`, and `adapters/rust/lexicon-rust` binaries; the compiled TypeScript `dist/cli.js` with `package.json`, `package-lock.json`, and production `node_modules`; and only Python/Ruby adapter source files. Tests, fixtures, caches, and build trees are not copied. `lexicon init --adapters release/adapters` may be used when the executable is not next to the distribution directory.

Build requirements are Go, Cargo/Rust, Node.js, and npm. Runtime requirements for a packaged distribution are the operating-system libraries required by the binaries, Node.js for the TypeScript adapter, Python for the Python adapter, and Ruby for the Ruby adapter. Go, Cargo, npm, and TypeScript source compilation are not required at runtime. The runner prefers packaged binaries and the compiled TypeScript entrypoint, while source-development execution remains available when those packaged paths are absent.
