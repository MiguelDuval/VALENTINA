# Safe local checks: no installer execution, registry writes or shortcuts.
[CmdletBinding()]
param()
$ErrorActionPreference = 'Stop'
$repo = Split-Path $PSScriptRoot -Parent
function Assert($ok, $message) { if (-not $ok) { throw $message } }
$nsi = Get-Content (Join-Path $repo 'installer/valentina.nsi') -Raw
Assert ($nsi -match 'RequestExecutionLevel user') 'Installer must never elevate'
Assert ($nsi -notmatch '(?im)^\s*(RMDir\s+/r|Delete\s+/REBOOTOK|WriteReg\w+\s+HKLM|ExecWait|ExecShell)') 'Unsafe installer operation'
Assert ($nsi -notmatch '(?im)^\s*(File|Delete)\s+.*\*') 'Wildcard payload or deletion'
Assert ($nsi -match 'SetShellVarContext current') 'Current user shell required'
Assert ($nsi -match 'CreateShortcut\s+"\$SMPROGRAMS\\VALENTINA\.lnk"\s+"\$INSTDIR\\VALENTINA\.exe"\s+"serve --open-browser"') 'Start Menu shortcut must launch the local web server, not the TUI'
foreach ($file in @('package-valentina-installer.ps1', 'test-installer.ps1', 'test-installer-package.ps1', 'test-runtime-intake.ps1')) {
    $tokens = $null; $errors = $null
    $null = [Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot $file), [ref]$tokens, [ref]$errors)
    Assert ($errors.Count -eq 0) "$file PowerShell syntax"
}
$workflow = Get-Content (Join-Path $repo '.github/workflows/valentina-ui-build.yml') -Raw
$runtime = $workflow.IndexOf('./scripts/test-runtime-intake.ps1')
$package = $workflow.IndexOf('./scripts/package-valentina-installer.ps1')
$accept = $workflow.IndexOf('./scripts/test-installer.ps1')
$upload = $workflow.IndexOf('name: Upload VALENTINA Windows artifact')
Assert ($runtime -gt 0 -and $package -gt $runtime -and $accept -gt $package -and $upload -gt $accept) 'All acceptance must precede both binary uploads'
Assert ($workflow -match 'VALENTINA-installer-windows-x64-') 'Separate installer artifact required'
Assert ($workflow -notmatch 'path: build-debug.log') 'Logs must not be uploaded as artifacts'
Write-Output 'INSTALLER_CONTRACT=PASS'
