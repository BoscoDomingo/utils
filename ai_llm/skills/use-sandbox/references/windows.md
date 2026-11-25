# Windows: Documentation Only

Windows execution is unsupported pending runner validation. Explain this path only; do not launch, share folders, enable features, or substitute host execution.

Microsoft documents `wsb` from **Windows 11 24H2**; Sandbox feature availability alone does not imply CLI support. For approved read-only discovery:

```powershell
Get-Command wsb -ErrorAction SilentlyContinue
```

If present, `wsb --help` checks local syntax. Missing tools/features block use; do not install or change virtualization/security settings. Host approvals still apply.

Defaults share network and clipboard. This illustrative configuration is **not executed or integration-tested**:

```xml
<Configuration>
  <Networking>Disable</Networking>
  <ClipboardRedirection>Disable</ClipboardRedirection>
  <AudioInput>Disable</AudioInput>
  <VideoInput>Disable</VideoInput>
  <PrinterRedirection>Disable</PrinterRedirection>
  <vGPU>Disable</vGPU>
  <ProtectedClient>Enable</ProtectedClient>
  <MappedFolders>
    <MappedFolder>
      <HostFolder>C:\SandboxInputs\Selected</HostFolder>
      <SandboxFolder>C:\Inputs</SandboxFolder>
      <ReadOnly>true</ReadOnly>
    </MappedFolder>
  </MappedFolders>
</Configuration>
```

The existing host folder must contain only approved, secret-free staged inputs. Use absolute, XML-escaped paths; never map a drive root, profile, whole repository, or credential directory. Read-only mappings disclose data. Keep work in the guest; persistent output would need separate approval.

`wsb exec` needs an ID, command, and run-as context; `ExistingLogin` needs an active guest session and does not imply non-admin execution. It returns an exit code but **does not retrieve process I/O**. Execution support needs validated isolation, session/argument handling, timeout/cancellation cleanup, and bounded result transport. Never enable network, clipboard, or broad writable mappings to work around missing I/O.

- [CLI availability and syntax](https://learn.microsoft.com/en-us/windows/security/application-security/application-isolation/windows-sandbox/windows-sandbox-cli)
- [Configuration options and defaults](https://learn.microsoft.com/en-us/windows/security/application-security/application-isolation/windows-sandbox/windows-sandbox-configure-using-wsb-file)
- [Installation prerequisites](https://learn.microsoft.com/en-us/windows/security/application-security/application-isolation/windows-sandbox/windows-sandbox-install)
