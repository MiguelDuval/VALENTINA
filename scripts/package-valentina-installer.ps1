#requires -Version 7.0
[CmdletBinding()]
param(
    [string]$PackageDirectory = (Join-Path $PSScriptRoot '../dist/package'),
    [string]$OutputDirectory = (Join-Path $PSScriptRoot '../dist/installer'),
    [string]$NsisArchivePath,
    [string]$MakeNsisPath
)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$repo = Split-Path $PSScriptRoot -Parent
$pin = 'f01c1d84c08b1de972d8cc268da25602598adff8'
$archiveHash = 'c7d27f780ddb6cffb4730138cd1591e841f4b7edb155856901cdf5f214394fa1'
$archiveUrl = 'https://distfiles.macports.org/nsis/nsis-3.11.zip'
function Require($condition, [string]$message) { if (-not $condition) { throw $message } }
function Hash([string]$path) { (Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash.ToLowerInvariant() }
function SafePath([string]$path) {
    $full = [IO.Path]::GetFullPath($path)
    Require ($full -notmatch '[\r\n"$;]') 'Path contains unsupported NSIS characters'
    return $full
}
$package = SafePath $PackageDirectory
$output = SafePath $OutputDirectory
Require (-not (Test-Path -LiteralPath $output)) 'Output directory already exists; choose a new directory (stale outputs are not reused)'
$upstream = Join-Path $repo 'upstream/huginn'
$actual = & git -C $upstream rev-parse HEAD
Require ($LASTEXITCODE -eq 0 -and $actual -eq $pin) 'Upstream pin mismatch'
$dirty = & git -C $upstream status --porcelain
Require ($LASTEXITCODE -eq 0 -and -not $dirty) 'Upstream must be clean'
$manifest = Get-Content -LiteralPath (Join-Path $package 'VALENTINA-BUILD-MANIFEST.json') -Raw | ConvertFrom-Json
$exe = Join-Path $package 'VALENTINA.exe'
$exeHash = Hash $exe
Require ($manifest.binary.file_name -eq 'VALENTINA.exe' -and $manifest.binary.sha256 -eq $exeHash) 'EXE SHA256 mismatch'
Require ($manifest.binary.file_size -eq (Get-Item -LiteralPath $exe).Length -and $manifest.binary.file_size -gt 0) 'EXE size mismatch'
Require ($manifest.binary.architecture -eq 'x64' -and $manifest.upstream.exact_pinned_sha -eq $pin -and $manifest.upstream.verified -eq $true) 'Build platform/upstream mismatch'
Require ($manifest.source.full_commit_sha -match '^[0-9a-f]{40}$') 'Invalid source SHA'
if ($env:GITHUB_ACTIONS -eq 'true') {
    $head = & git -C $repo rev-parse HEAD
    Require ($LASTEXITCODE -eq 0 -and $manifest.source.full_commit_sha -eq $head) 'CI build must match checked-out HEAD'
}
foreach ($check in @('windows_executable_compilation','upstream_pin_verification','patch_application')) {
    Require ($manifest.validation.$check.status -eq 'passed') "Missing build validation: $check"
}
$version = [string]$manifest.product_version
Require ($version -match '^\d{1,4}\.\d{1,4}\.\d{1,4}$') 'Invalid product version'
$sums = Get-Content -LiteralPath (Join-Path $package 'SHA256SUMS.txt') -Raw
Require ($sums.Trim() -match ('(?i)^' + $exeHash + '\s+VALENTINA\.exe$')) 'EXE checksum file mismatch'
$license = Join-Path $upstream 'LICENSE'
Require ((Hash (Join-Path $package 'LICENSES/LICENSE.huginn.txt')) -eq (Hash $license)) 'Packaged upstream license must match pin verbatim'
# Verify PE signature and AMD64 machine without executing the application.
$stream = [IO.File]::OpenRead($exe)
$reader = [IO.BinaryReader]::new($stream)
try {
    Require ($reader.ReadUInt16() -eq 0x5a4d) 'Missing DOS executable header'
    $stream.Position = 0x3c; $offset = $reader.ReadInt32()
    Require ($offset -ge 64 -and $offset -lt ($stream.Length - 6)) 'Invalid PE offset'
    $stream.Position = $offset
    Require ($reader.ReadUInt32() -eq 0x4550 -and $reader.ReadUInt16() -eq 0x8664) 'Expected Windows AMD64 PE'
} finally { $reader.Dispose(); $stream.Dispose() }

$work = Join-Path $repo ('build/installer-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $work -Force | Out-Null
# Retain ignored compiler scratch for diagnostics; never publish this directory.
$archive = Join-Path $work 'nsis-3.11.zip'
if ($NsisArchivePath) { Copy-Item -LiteralPath $NsisArchivePath -Destination $archive }
else {
    # No HTTP redirects, unpinned package manager or fallback version.
    $handler = [Net.Http.HttpClientHandler]::new()
    $handler.AllowAutoRedirect = $false
    $client = [Net.Http.HttpClient]::new($handler)
    $client.Timeout = [TimeSpan]::FromSeconds(90)
    try {
        $uri = [uri]$archiveUrl
        for ($hop = 0; $hop -le 5; $hop++) {
            Require ($uri.Scheme -eq 'https') 'NSIS download requires HTTPS at every redirect'
            $response = $client.GetAsync($uri).GetAwaiter().GetResult()
            try {
                if ([int]$response.StatusCode -in @(301,302,303,307,308)) {
                    Require ($null -ne $response.Headers.Location -and $hop -lt 5) 'Invalid NSIS redirect'
                    $uri = [uri]::new($uri, $response.Headers.Location)
                    continue
                }
                $null = $response.EnsureSuccessStatusCode()
                [IO.File]::WriteAllBytes($archive, $response.Content.ReadAsByteArrayAsync().GetAwaiter().GetResult())
                break
            } finally { $response.Dispose() }
        }
    } finally { $client.Dispose(); $handler.Dispose() }
}
Require ((Hash $archive) -eq $archiveHash) 'NSIS 3.11 archive SHA256 mismatch; compiler was NOT executed'
Expand-Archive -LiteralPath $archive -DestinationPath (Join-Path $work 'tools')
$verifiedRoot = Join-Path $work 'tools/nsis-3.11'
$compiler = Join-Path $verifiedRoot 'makensis.exe'
Require (Test-Path -LiteralPath $compiler) 'Verified archive lacks makensis.exe'
if ($MakeNsisPath) {
    Require ($NsisArchivePath) 'Local makensis requires the pinned archive for independent verification'
    $localRoot = Split-Path (Resolve-Path -LiteralPath $MakeNsisPath).Path -Parent
    # Verify every distributed file (plugins/includes/stubs too), not only EXE.
    foreach ($file in Get-ChildItem -LiteralPath $verifiedRoot -Recurse -File) {
        $relative = [IO.Path]::GetRelativePath($verifiedRoot, $file.FullName)
        Require ((Hash (Join-Path $localRoot $relative)) -eq (Hash $file.FullName)) 'Local NSIS distribution differs from verified archive'
    }
    Require ((Hash $MakeNsisPath) -eq (Hash $compiler)) 'Local makensis mismatch'
    # Compile with the fresh verified extraction; extra local plugins/config are never loaded.
}
$payload = Join-Path $work 'payload'
New-Item -ItemType Directory -Path (Join-Path $payload 'LICENSES') -Force | Out-Null
foreach ($name in @('VALENTINA.exe','VALENTINA-BUILD-MANIFEST.json','SHA256SUMS.txt')) {
    Copy-Item -LiteralPath (Join-Path $package $name) -Destination (Join-Path $payload $name)
}
Copy-Item -LiteralPath $license -Destination (Join-Path $payload 'LICENSES/LICENSE.huginn.txt')
# Only tracked license/notice text from the exact clean upstream, never fonts,
# node_modules, credentials or recursively copied package contents.
$notices = [Text.StringBuilder]::new()
$null = $notices.AppendLine('Available pinned Huginn vendored dependency notices. Frontend/transitive inventory is incomplete; legal acceptance remains open.')
$paths = @(& git -C $upstream ls-files 'vendor/*')
Require ($LASTEXITCODE -eq 0) 'Could not enumerate pinned third-party notices'
foreach ($path in ($paths | Where-Object { (Split-Path $_ -Leaf) -match '^(LICENSE|NOTICE|COPYING)([.\-_].*)?$' } | Sort-Object)) {
    $null = $notices.AppendLine("`n===== $path =====")
    $null = $notices.AppendLine([IO.File]::ReadAllText((Join-Path $upstream $path)))
}
[IO.File]::WriteAllText((Join-Path $payload 'LICENSES/THIRD-PARTY-NOTICES.txt'), $notices.ToString())
@'
VALENTINA — локальная тестовая сборка Windows x64
Запустите VALENTINA из меню Пуск. Права администратора не нужны.
Перед обновлением и удалением закройте приложение.
Удаление сохраняет настройки, базу, профиль, ключи, доказательства и резервные копии.
Никакие Go/Node/Python и модели устанавливать не требуется.
Лицензия Huginn: LICENSES/LICENSE.huginn.txt (полный исходный текст).
Доступные сторонние лицензии: LICENSES/THIRD-PARTY-NOTICES.txt.
Инвентаризация лицензий frontend/transitive зависимостей пока неполная.
Это внутренняя тестовая сборка, не официальный релиз. Windows 11 acceptance отдельно.
'@ | Set-Content -LiteralPath (Join-Path $payload 'README.txt') -Encoding utf8
$candidate = Join-Path $work 'VALENTINA-Setup-x64.exe'
& $compiler /V2 /WX /NOCD /NOCONFIG "/DPAYLOAD=$payload" "/DOUTPUT_FILE=$candidate" "/DPRODUCT_VERSION=$version" (Join-Path $repo 'installer/valentina.nsi')
Require ($LASTEXITCODE -eq 0 -and (Test-Path -LiteralPath $candidate)) 'NSIS compilation failed'
$size = (Get-Item -LiteralPath $candidate).Length
Require ($size -gt 0) 'Empty setup executable'
$setupBytes = [Text.Encoding]::ASCII.GetString([IO.File]::ReadAllBytes($candidate))
Require ($setupBytes.Contains('requestedExecutionLevel level="asInvoker"')) 'Setup manifest must request current-user execution'
$setupBytes = $null
$setupHash = Hash $candidate
New-Item -ItemType Directory -Path $output | Out-Null
Copy-Item -LiteralPath $candidate -Destination (Join-Path $output 'VALENTINA-Setup-x64.exe')
"$setupHash  VALENTINA-Setup-x64.exe" | Set-Content (Join-Path $output 'SHA256SUMS.txt') -Encoding utf8
@{ schema_version=1; source_commit=$manifest.source.full_commit_sha; upstream=$pin; product_version=$version;
   nsis_version='3.11'; nsis_archive_sha256=$archiveHash; nsis_archive_url=$archiveUrl;
   binary=@{file_name='VALENTINA-Setup-x64.exe';sha256=$setupHash;file_size=$size}; payload_exe_sha256=$exeHash;
   license_inventory='partial: pinned upstream and available vendored notices; frontend/transitive audit pending';
   acceptance='not established by compilation; run test-installer.ps1 on a disposable GitHub runner'
} | ConvertTo-Json -Depth 5 | Set-Content (Join-Path $output 'VALENTINA-INSTALLER-MANIFEST.json') -Encoding utf8
Write-Output "INSTALLER_COMPILE=PASS SETUP_SHA256=$setupHash SETUP_SIZE=$size"
