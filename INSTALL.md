# Install

## macOS

If macOS blocks the unsigned alpha binary after download, remove the quarantine flag from the extracted `gither` binary:

```bash
xattr -d com.apple.quarantine gither
chmod +x gither
```

If you extracted the whole archive into a folder and want to clear the folder recursively instead:

```bash
xattr -dr com.apple.quarantine .
```

## Linux

Mark the binary executable if needed:

```bash
chmod +x gither
```

## Windows

Extract the archive and run `gither.exe` from PowerShell or Command Prompt.
