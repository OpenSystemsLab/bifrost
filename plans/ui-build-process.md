# UI Build Process for Bifrost HTTP Transport

## Problem
The `//go:embed all:ui` directive in `transports/bifrost-http/main.go` was failing with the error:
```
pattern all:ui: no matching files found
```

## Root Cause
The UI files are located in the project root's `ui/` directory, but the go:embed directive in `bifrost-http/main.go` expects them to be in a local `ui/` directory relative to the Go binary.

## Solution
The UI must be built and copied to the correct location before building the Go application:

### Steps to Build UI:
1. Navigate to the `ui/` directory
2. Install dependencies: `npm install`
3. Build the UI: `npm run build`
4. Copy built files to bifrost-http (Windows PowerShell):
   ```powershell
   if (Test-Path "../transports/bifrost-http/ui") { Remove-Item -Recurse -Force "../transports/bifrost-http/ui" }
   Copy-Item -Recurse "out" "../transports/bifrost-http/ui"
   ```

### Build Process:
1. The `npm run build` command:
   - Runs `next build` to create the static export
   - Runs `fix-paths` script to adjust relative paths
   - Attempts to run `copy-build` (Unix commands, fails on Windows)

2. Manual copy step needed on Windows:
   - Remove existing ui directory if present
   - Copy the `out/` directory to `../transports/bifrost-http/ui`

### Verification:
After copying, the go:embed directive should work correctly:
```bash
cd transports/bifrost-http
go build -o bifrost-http.exe .
```

## Notes
- The `copy-build` script in `ui/package.json` uses Unix commands (`rm`, `cp`) that don't work on Windows
- For Windows development, manual PowerShell commands are needed
- The Makefile includes automation for this process in the `dev-http` and `build` targets
- Production builds should use the Makefile or CI/CD pipeline for proper cross-platform support

## Future Improvements
- Update the `copy-build` script to be cross-platform compatible
- Consider using a build tool that works on both Windows and Linux
- Add validation to ensure UI files exist before Go build