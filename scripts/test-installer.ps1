#requires -Version 7.0
# Integration only. Never invoke on an operator's workstation.
[CmdletBinding()]
param(
    [string]$InstallerDirectory = (Join-Path $PSScriptRoot '../dist/installer'),
    [switch]$DisposableRunner
)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
function Require($ok, [string]$message) { if (-not $ok) { throw "Installer acceptance: $message" } }
Require ($DisposableRunner -and $env:GITHUB_ACTIONS -eq 'true' -and $env:RUNNER_ENVIRONMENT -eq 'github-hosted' -and $env:RUNNER_OS -eq 'Windows' -and $env:RUNNER_TEMP) 'Only an explicitly acknowledged disposable GitHub-hosted Windows runner is allowed'
$setup = (Resolve-Path (Join-Path $InstallerDirectory 'VALENTINA-Setup-x64.exe')).Path
$manifest = Get-Content (Join-Path $InstallerDirectory 'VALENTINA-INSTALLER-MANIFEST.json') -Raw | ConvertFrom-Json
Require ((Get-FileHash $setup).Hash -eq $manifest.binary.sha256 -and (Get-Item $setup).Length -eq $manifest.binary.file_size) 'Setup integrity mismatch'
$regSubkey = 'Software\Microsoft\Windows\CurrentVersion\Uninstall\VALENTINA'
$reg = [Microsoft.Win32.RegistryKey]::OpenBaseKey('CurrentUser','Registry64')
$reg32 = [Microsoft.Win32.RegistryKey]::OpenBaseKey('CurrentUser','Registry32')
$shortcut = Join-Path ([Environment]::GetFolderPath('Programs')) 'VALENTINA.lnk'
Require (-not $reg.OpenSubKey($regSubkey) -and -not $reg32.OpenSubKey($regSubkey)) 'Existing installation registration; refusing to clobber it'
Require (-not (Test-Path $shortcut) -and -not (Test-Path (Join-Path $env:LOCALAPPDATA 'Programs/VALENTINA'))) 'Existing installation/shortcut; refusing to clobber it'
$root = Join-Path $env:RUNNER_TEMP ('valentina-installer-smoke-' + [guid]::NewGuid().ToString('N'))
$install = Join-Path $root 'program with spaces'
$profile = Join-Path $root 'isolated-profile'
New-Item -ItemType Directory -Path $root | Out-Null
# Synthetic pre-existing data outside the install directory must survive every phase.
$sentinels = @('.huginn/config.json','AppData/Local/VALENTINA/profile.yaml','AppData/Local/VALENTINA/records.sqlite','AppData/Local/VALENTINA/evidence/sentinel.txt','AppData/Local/VALENTINA/backups/sentinel.txt')
foreach ($relative in $sentinels) {
    $path = Join-Path $profile $relative
    New-Item -ItemType Directory -Path (Split-Path $path) -Force | Out-Null
    'SYNTHETIC-PRESERVE-ONLY' | Set-Content -LiteralPath $path
}
$baseline = @{}
$runtimeBaseline = @{}
$runtimeProfile = Join-Path $root 'valentina-runtime-smoke-preserved'
foreach ($relative in $sentinels) { $baseline[$relative] = (Get-FileHash (Join-Path $profile $relative)).Hash }
function Check-Sentinels {
    foreach ($relative in $sentinels) {
        Require ((Get-FileHash (Join-Path $profile $relative)).Hash -eq $baseline[$relative]) 'Pre-existing app data changed'
    }
    foreach ($path in $runtimeBaseline.Keys) {
        Require ((Get-FileHash -LiteralPath $path).Hash -eq $runtimeBaseline[$path]) 'Actual runtime config/database/evidence changed during setup'
    }
}
function Run-Setup([string]$binary, [string]$arguments, [int]$expected = 0) {
    $p = Start-Process -FilePath $binary -ArgumentList $arguments -PassThru -WindowStyle Hidden
    # No forced termination: a timeout fails the job; the disposable runner owns cleanup.
    Require ($p.WaitForExit(120000)) 'Installer process timed out'
    Require ($p.ExitCode -eq $expected) "Unexpected installer exit code $($p.ExitCode), expected $expected"
}
function Check-Installed {
    Require ((Get-FileHash (Join-Path $install 'VALENTINA.exe')).Hash -eq $manifest.payload_exe_sha256) 'Installed EXE differs'
    foreach ($name in @('Uninstall.exe','README.txt','SHA256SUMS.txt','VALENTINA-BUILD-MANIFEST.json','LICENSES/LICENSE.huginn.txt','LICENSES/THIRD-PARTY-NOTICES.txt')) {
        Require (Test-Path (Join-Path $install $name)) "Missing installed file $name"
    }
    Require ((Get-FileHash (Join-Path $install 'LICENSES/LICENSE.huginn.txt')).Hash -eq (Get-FileHash (Join-Path $PSScriptRoot '../upstream/huginn/LICENSE')).Hash) 'Upstream license differs'
    $key = $reg.OpenSubKey($regSubkey)
    Require ($null -ne $key) 'Missing HKCU registration'
    try {
        Require ($key.GetValue('InstallLocation') -eq $install -and $key.GetValue('DisplayVersion') -eq $manifest.product_version) 'Registry location/version'
        Require ($key.GetValue('UninstallString') -eq ('"' + $install + '\Uninstall.exe"')) 'Quoted uninstall command'
        Require ($key.GetValue('QuietUninstallString') -eq ('"' + $install + '\Uninstall.exe" /S')) 'Quiet uninstall command'
    } finally { $key.Dispose() }
    Require (Test-Path $shortcut) 'Missing Start Menu shortcut'
    $shell = New-Object -ComObject WScript.Shell
    $link = $shell.CreateShortcut($shortcut)
    Require ($link.TargetPath -eq (Join-Path $install 'VALENTINA.exe') -and $link.Arguments -eq 'serve') 'Shortcut target/arguments'
    Check-Sentinels
}
try {
    $occupied = Join-Path $root 'occupied'
    New-Item -ItemType Directory -Path $occupied | Out-Null
    'SYNTHETIC-DO-NOT-OVERWRITE' | Set-Content (Join-Path $occupied 'VALENTINA.exe')
    $occupiedHash = (Get-FileHash (Join-Path $occupied 'VALENTINA.exe')).Hash
    Run-Setup $setup "/S /D=$occupied" 3
    Require ((Get-FileHash (Join-Path $occupied 'VALENTINA.exe')).Hash -eq $occupiedHash) 'Nonempty target changed'
    $junction = Join-Path $root 'junction-target'
    $junctionData = Join-Path $root 'junction-data'
    New-Item -ItemType Directory -Path $junctionData | Out-Null
    New-Item -ItemType Junction -Path $junction -Target $junctionData | Out-Null
    Run-Setup $setup "/S /D=$junction" 3
    Require (@(Get-ChildItem -LiteralPath $junctionData -Force).Count -eq 0) 'Junction target changed'
    # Fresh install into a unique empty target, /D must be last and unquoted.
    Run-Setup $setup "/S /D=$install"
    Check-Installed
    Run-Setup $setup ('/S /D=' + (Join-Path $root 'other-location')) 3
    Check-Installed
    # Preserve unknown operator-created files even inside the program directory.
    $unknown = Join-Path $install 'operator-owned.txt'
    'SYNTHETIC-UNKNOWN-FILE' | Set-Content $unknown
    $unknownHash = (Get-FileHash $unknown).Hash
    # Locked executable must reject upgrade and uninstall without killing anything.
    $lock = [IO.File]::Open((Join-Path $install 'VALENTINA.exe'), 'Open', 'Read', 'Read')
    try {
        Run-Setup $setup "/S /D=$install" 2
        # _?= avoids NSIS background relocation so this process exit is authoritative.
        Run-Setup (Join-Path $install 'Uninstall.exe') "/S _?=$install" 2
    } finally { $lock.Dispose() }
    Check-Installed
    # Run the existing HMAC, business API, persistence and actual TCP bind checks
    # in a child PowerShell; its synthetic environment never leaks into setup.
    & pwsh -NoProfile -File (Join-Path $PSScriptRoot 'test-runtime-intake.ps1') -ExecutablePath (Join-Path $install 'VALENTINA.exe') -TestProfileRoot $runtimeProfile -RetainTestProfile
    Require ($LASTEXITCODE -eq 0) 'Installed runtime smoke failed'
    # Snapshot real synthetic config and persistent records, not only placeholders.
    foreach ($file in Get-ChildItem -LiteralPath $runtimeProfile -Recurse -File) {
        $runtimeBaseline[$file.FullName] = (Get-FileHash -LiteralPath $file.FullName).Hash
    }
    Require ($runtimeBaseline.Count -gt 2) 'Runtime did not retain test data'
    Run-Setup $setup "/S /D=$install"
    Check-Installed
    Require ((Get-FileHash $unknown).Hash -eq $unknownHash) 'Upgrade changed unknown file'
    & pwsh -NoProfile -File (Join-Path $PSScriptRoot 'test-runtime-intake.ps1') -ExecutablePath (Join-Path $install 'VALENTINA.exe')
    Require ($LASTEXITCODE -eq 0) 'Upgraded runtime smoke failed'
    # Copy only the generated uninstaller to our test directory. _?= makes it
    # synchronous while allowing deletion of the installed Uninstall.exe.
    $uninstaller = Join-Path $root 'acceptance-uninstall.exe'
    Copy-Item (Join-Path $install 'Uninstall.exe') $uninstaller
    Run-Setup $uninstaller "/S _?=$install"
    Require (-not $reg.OpenSubKey($regSubkey) -and -not (Test-Path $shortcut)) 'Registration/shortcut survived uninstall'
    $remaining = @(Get-ChildItem -LiteralPath $install -File -Recurse)
    Require ($remaining.Count -eq 1 -and $remaining[0].FullName -eq $unknown) 'Installed program files survived uninstall'
    Require ((Get-FileHash $unknown).Hash -eq $unknownHash) 'Uninstall removed unknown file'
    Check-Sentinels
    Write-Output 'INSTALLER_ACCEPTANCE=PASS (install, locked-file rejection, upgrade, installed runtime, uninstall, data preservation)'
} finally {
    $reg.Dispose(); $reg32.Dispose()
    # Deliberately no registry/data cleanup fallback. Failure is investigated on
    # the disposable runner; never turn a failing uninstaller into a passing test.
}
