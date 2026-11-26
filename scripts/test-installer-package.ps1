#requires -Version 7.0
# Synthetic package tests. Never execute Setup or modify HKCU/Start Menu.
[CmdletBinding()]
param([string]$NsisArchivePath, [string]$ReferenceExecutable, [string]$MakeNsisPath)
$ErrorActionPreference = 'Stop'
$repo = Split-Path $PSScriptRoot -Parent
$root = Join-Path $repo ('build/installer-test-' + [guid]::NewGuid().ToString('N'))
$package = Join-Path $root 'package'
New-Item -ItemType Directory -Path (Join-Path $package 'LICENSES') -Force | Out-Null
$exe = Join-Path $package 'VALENTINA.exe'
if ($ReferenceExecutable) {
    # Previously independently tested Gate 2 build. Never trust a filename alone.
    if ((Get-FileHash $ReferenceExecutable).Hash -ne 'F25DD8C4E9C1BF666E7708FD3C8658DCBDB1AEF9B32DBC39B75C85544A75E939') { throw 'Gate 2 reference SHA mismatch' }
    Copy-Item -LiteralPath $ReferenceExecutable -Destination $exe
} else {
    $bytes = [byte[]]::new(512)
    $bytes[0]=0x4d; $bytes[1]=0x5a; $bytes[0x3c]=0x80
    $bytes[0x80]=0x50; $bytes[0x81]=0x45; $bytes[0x84]=0x64; $bytes[0x85]=0x86
    [IO.File]::WriteAllBytes($exe,$bytes)
}
$hash = (Get-FileHash $exe).Hash
$manifest = @{
    test_fixture=$true; product_version='1.0.0'; source=@{full_commit_sha=(& git -C $repo rev-parse HEAD)}
    binary=@{file_name='VALENTINA.exe';sha256=$hash;file_size=(Get-Item $exe).Length;architecture='x64'}
    upstream=@{exact_pinned_sha='f01c1d84c08b1de972d8cc268da25602598adff8';verified=$true}
    validation=@{windows_executable_compilation=@{status='passed'};upstream_pin_verification=@{status='passed'};patch_application=@{status='passed'}}
}
$manifestPath = Join-Path $package 'VALENTINA-BUILD-MANIFEST.json'
function Save-Manifest { $manifest | ConvertTo-Json -Depth 6 | Set-Content $manifestPath }
Save-Manifest
"$hash  VALENTINA.exe" | Set-Content (Join-Path $package 'SHA256SUMS.txt')
$license = Join-Path $package 'LICENSES/LICENSE.huginn.txt'
Copy-Item (Join-Path $repo 'upstream/huginn/LICENSE') $license
$badArchive = Join-Path $root 'invalid-nsis.zip'
'not a compiler' | Set-Content $badArchive
function Reject([string]$expected) {
    $out = Join-Path $root ([guid]::NewGuid().ToString('N'))
    $caught = $false
    try { & (Join-Path $PSScriptRoot 'package-valentina-installer.ps1') -PackageDirectory $package -OutputDirectory $out -NsisArchivePath $badArchive }
    catch { if ($_.Exception.Message -notlike "*$expected*") { throw }; $caught = $true }
    if (-not $caught -or (Test-Path $out)) { throw "Did not reject safely: $expected" }
    Write-Output "REJECTION=PASS ($expected)"
}
$manifest.binary.sha256 = '0' * 64; Save-Manifest
Reject 'EXE SHA256 mismatch'
$manifest.binary.sha256 = $hash; $manifest.binary.file_size++; Save-Manifest
Reject 'EXE size mismatch'
$manifest.binary.file_size--; $manifest.upstream.exact_pinned_sha = '0' * 40; Save-Manifest
Reject 'Build platform/upstream mismatch'
$manifest.upstream.exact_pinned_sha = 'f01c1d84c08b1de972d8cc268da25602598adff8'; Save-Manifest
$manifest.product_version = '1.0.0$invalid'; Save-Manifest
Reject 'Invalid product version'
$manifest.product_version = '1.0.0'; Save-Manifest
'bad checksum' | Set-Content (Join-Path $package 'SHA256SUMS.txt')
Reject 'EXE checksum file mismatch'
"$hash  VALENTINA.exe" | Set-Content (Join-Path $package 'SHA256SUMS.txt')
'truncated license' | Set-Content $license
Reject 'license must match pin verbatim'
Copy-Item (Join-Path $repo 'upstream/huginn/LICENSE') $license -Force
Reject 'NSIS 3.11 archive SHA256 mismatch'
# Guard the full integration entry point on ordinary workstations.
if ($env:GITHUB_ACTIONS -ne 'true') {
    $rejected = $false
    try { & (Join-Path $PSScriptRoot 'test-installer.ps1') }
    catch { if ($_.Exception.Message -notlike '*disposable GitHub-hosted*') { throw }; $rejected = $true }
    if (-not $rejected) { throw 'Integration workstation guard did not reject' }
    Write-Output 'WORKSTATION_GUARD=PASS'
    $rejected = $false
    try { & (Join-Path $PSScriptRoot 'test-runtime-intake.ps1') -ExecutablePath $exe -TestProfileRoot (Join-Path $root 'must-not-exist') -RetainTestProfile }
    catch { if ($_.Exception.Message -notlike '*disposable GitHub-hosted*') { throw }; $rejected = $true }
    if (-not $rejected -or (Test-Path (Join-Path $root 'must-not-exist'))) { throw 'Retained-profile workstation guard did not reject safely' }
    Write-Output 'RETAINED_PROFILE_GUARD=PASS'
}
if ($NsisArchivePath) {
    # Extra package files must never enter the explicit payload.
    'SYNTHETIC-DO-NOT-PACKAGE' | Set-Content (Join-Path $package 'private-profile.txt')
    $out = Join-Path $root 'compiled'
    & (Join-Path $PSScriptRoot 'package-valentina-installer.ps1') -PackageDirectory $package -OutputDirectory $out -NsisArchivePath $NsisArchivePath -MakeNsisPath $MakeNsisPath
    $result = Get-Content (Join-Path $out 'VALENTINA-INSTALLER-MANIFEST.json') -Raw | ConvertFrom-Json
    if ($result.payload_exe_sha256 -ne $hash -or (Get-FileHash (Join-Path $out 'VALENTINA-Setup-x64.exe')).Hash -ne $result.binary.sha256) { throw 'Compiled manifest mismatch' }
    Write-Output "LOCAL_COMPILE_FIXTURE=$out (not CI acceptance or distributable build)"
}
Write-Output 'INSTALLER_PACKAGE_TESTS=PASS'
