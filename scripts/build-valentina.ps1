[CmdletBinding()]
param([string]$OutputDirectory = "")
$ErrorActionPreference = "Stop"
$Repo = Split-Path $PSScriptRoot -Parent
$Upstream = Join-Path $Repo "upstream\huginn"
$Patch = Join-Path $Repo "patches\upstream-ui\001-valentina-ui-foundation.patch"
$Expected = "f01c1d84c08b1de972d8cc268da25602598adff8"
if ([string]::IsNullOrWhiteSpace($OutputDirectory)) { $OutputDirectory = Join-Path $Repo "dist" }
elseif (-not [IO.Path]::IsPathRooted($OutputDirectory)) { $OutputDirectory = Join-Path $Repo $OutputDirectory }
foreach ($tool in @("git","npm","go","robocopy")) { if (-not (Get-Command $tool -ErrorAction SilentlyContinue)) { throw "Required tool missing: $tool" } }
if (-not (Test-Path $Upstream)) { throw "Initialize upstream with scripts/bootstrap.ps1 first." }
$actual = (& git -C $Upstream rev-parse HEAD).Trim()
if ($LASTEXITCODE -ne 0 -or $actual -ne $Expected) { throw "Upstream pin mismatch. Expected $Expected, got $actual." }
if (-not (Test-Path $Patch)) { throw "Missing VALENTINA UI patch: $Patch" }
$tmp = Join-Path ([IO.Path]::GetTempPath()) ("valentina-build-" + [guid]::NewGuid().ToString("N"))
$copy = Join-Path $tmp "source"
New-Item -ItemType Directory -Path $copy -Force | Out-Null
try {
 & robocopy $Upstream $copy /E /XD .git /XF .git /R:1 /W:1 /NFL /NDL /NJH /NJS /NP | Out-Null
 if ($LASTEXITCODE -gt 7) { throw "Upstream source copy failed." }
 & git -C $copy init --quiet
 if ($LASTEXITCODE -ne 0) { throw "Could not initialize temporary source." }
 & git -C $copy apply --check $Patch
 if ($LASTEXITCODE -ne 0) { throw "VALENTINA patch does not apply to pinned source." }
 & git -C $copy apply $Patch
 if ($LASTEXITCODE -ne 0) { throw "Could not apply VALENTINA patch." }
 # Copy VALENTINA-first-party Go code into the temporary build tree.
 $ValentinaGo = Join-Path $Repo "valentina"
 $DestInternal = Join-Path $copy "internal"
 if (Test-Path $ValentinaGo) {
   & robocopy $ValentinaGo $DestInternal /E /R:1 /W:1 /NFL /NDL /NJH /NJS /NP | Out-Null
   if ($LASTEXITCODE -gt 7) { throw "VALENTINA Go source copy failed." }
 }
  # Copy VALENTINA-first-party Vue views into the temporary build tree.
  $ValentinaViews = Join-Path $Repo "valentina\web\src\views"
  $DestViews = Join-Path $copy "web\src\views"
  if (Test-Path $ValentinaViews) {
    & robocopy $ValentinaViews $DestViews /E /R:1 /W:1 /NFL /NDL /NJH /NJS /NP | Out-Null
    if ($LASTEXITCODE -gt 7) { throw "VALENTINA Vue views copy failed." }
  }
 Push-Location (Join-Path $copy "web")
 try {
  npm ci --no-audit --no-fund
  if ($LASTEXITCODE -ne 0) { throw "npm ci failed." }
  npm test
  if ($LASTEXITCODE -ne 0) { throw "Frontend tests failed." }
  npm run build
  if ($LASTEXITCODE -ne 0) { throw "Vue frontend build failed." }
 } finally { Pop-Location }
 New-Item -ItemType Directory -Path $OutputDirectory -Force | Out-Null
 $exe = Join-Path $OutputDirectory "VALENTINA.exe"
 Push-Location $copy
 try {
  & go build -tags embed_frontend -ldflags="-X main.version=valentina-dev" -o $exe .
  if ($LASTEXITCODE -ne 0) { throw "Go build failed." }
 } finally { Pop-Location }
 Write-Host "VALENTINA built locally: $exe" -ForegroundColor Green
} finally { if (Test-Path $tmp) { Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue } }
