<!-- This wiki-style note explains why we rely on pure Go builds for portability. -->
# Build and Release Notes

## Static binaries (no libc dependency)

The proxy is intended to run on older Debian releases without requiring newer glibc versions.  
To keep the binaries free of libc dependencies, builds must disable cgo and use the Go DNS resolver.  
Both the GitHub Actions workflow and the local build script set `CGO_ENABLED=0` and `-tags netgo`.  

## Local cross-compilation

Use the helper script to produce the same matrix as CI:

```bash
./scripts/build-release.sh <version>
```

Pass a version string (for example, `42` or `dev`) and the binaries will be placed in `dist/`.
