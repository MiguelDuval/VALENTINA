[CmdletBinding()]
param([string]$OutputDirectory = "")
$ErrorActionPreference = "Stop"
$Repo = Split-Path $PSScriptRoot -Parent
$Upstream = Join-Path $Repo "upstream\huginn"
$Patch1 = Join-Path $Repo "patches\upstream-ui\001-valentina-ui-foundation.patch"
$Patch2 = Join-Path $Repo "patches\upstream-ui\002-valentina-crm-integration.patch"
$Expected = "f01c1d84c08b1de972d8cc268da25602598adff8"
if ([string]::IsNullOrWhiteSpace($OutputDirectory)) { $OutputDirectory = Join-Path $Repo "dist" }
elseif (-not [IO.Path]::IsPathRooted($OutputDirectory)) { $OutputDirectory = Join-Path $Repo $OutputDirectory }
foreach ($tool in @("git","npm","go","robocopy")) { if (-not (Get-Command $tool -ErrorAction SilentlyContinue)) { throw "Required tool missing: $tool" } }
if (-not (Test-Path $Upstream)) { throw "Initialize upstream with scripts/bootstrap.ps1 first." }
$actual = (& git -C $Upstream rev-parse HEAD).Trim()
if ($LASTEXITCODE -ne 0 -or $actual -ne $Expected) { throw "Upstream pin mismatch. Expected $Expected, got $actual." }
if (-not (Test-Path $Patch1)) { throw "Missing VALENTINA UI patch: $Patch1" }
if (-not (Test-Path $Patch2)) { throw "Missing VALENTINA CRM integration patch: $Patch2" }
$tmp = Join-Path ([IO.Path]::GetTempPath()) ("valentina-build-" + [guid]::NewGuid().ToString("N"))
$copy = Join-Path $tmp "source"
New-Item -ItemType Directory -Path $copy -Force | Out-Null
try {
 # 1. Copy pinned upstream into a temporary build copy.
 & robocopy $Upstream $copy /E /XD .git /XF .git /R:1 /W:1 /NFL /NDL /NJH /NJS /NP | Out-Null
 if ($LASTEXITCODE -gt 7) { throw "Upstream source copy failed." }
 & git -C $copy init --quiet
 if ($LASTEXITCODE -ne 0) { throw "Could not initialize temporary source." }

 # 2. Apply all tracked patches with git apply --check (fail closed if context differs).
 foreach ($patch in @($Patch1, $Patch2)) {
   & git -C $copy apply --check $patch
   if ($LASTEXITCODE -ne 0) { throw "VALENTINA patch does not apply to pinned source: $patch" }
   & git -C $copy apply $patch
   if ($LASTEXITCODE -ne 0) { throw "Could not apply VALENTINA patch: $patch" }
 }

 # 3. Copy VALENTINA-first-party Go source into the correct staging path:
 #    source/internal/valentina/internal/{domain,api}
 $ValentinaGo = Join-Path $Repo "valentina"
 $DestInternal = Join-Path $copy "internal"
 if (Test-Path $ValentinaGo) {
   & robocopy $ValentinaGo $DestInternal /E /R:1 /W:1 /NFL /NDL /NJH /NJS /NP | Out-Null
   if ($LASTEXITCODE -gt 7) { throw "VALENTINA Go source copy failed." }
 }

 # 4. Copy VALENTINA-first-party Vue views and CRM tests into the staging tree.
 $ValentinaViews = Join-Path $Repo "valentina\web\src\views"
 $DestViews = Join-Path $copy "web\src\views"
 if (Test-Path $ValentinaViews) {
   & robocopy $ValentinaViews $DestViews /E /R:1 /W:1 /NFL /NDL /NJH /NJS /NP | Out-Null
   if ($LASTEXITCODE -gt 7) { throw "VALENTINA Vue views copy failed." }
 }

 # 5. Verify the upstream server really imports the CRM API package.
 $serverFiles = @("main.go", "internal\server\server.go")
 $foundImport = $false
 foreach ($sf in $serverFiles) {
   $pf = Join-Path $copy $sf
   if (Test-Path $pf) {
     $content = Get-Content -LiteralPath $pf -Raw
     if ($content -like '*internal/valentina/internal/*') {
       $foundImport = $true
     }
   }
 }
 if (-not $foundImport) { throw "CRM API package is not imported by the upstream server." }

 # 6. Targeted Go tests for first-party packages in the upstream module context.
 Push-Location $copy
 try {
   & go test ./internal/valentina/internal/...
   if ($LASTEXITCODE -ne 0) { throw "CRM Go tests failed." }
 } finally { Pop-Location }

 # 7. Frontend: typecheck, tests (with CRM views), production build.
 Push-Location (Join-Path $copy "web")
 try {
   npm ci --no-audit --no-fund
   if ($LASTEXITCODE -ne 0) { throw "npm ci failed." }
   & npx vue-tsc --noEmit -p tsconfig.json
   if ($LASTEXITCODE -ne 0) { throw "Vue typecheck failed." }
   npm test
   if ($LASTEXITCODE -ne 0) { throw "Frontend tests failed." }
   npm run build
   if ($LASTEXITCODE -ne 0) { throw "Vue frontend build failed." }
 } finally { Pop-Location }

 # 8. Compile Windows executable from the same source tree.
 New-Item -ItemType Directory -Path $OutputDirectory -Force | Out-Null
 $exe = Join-Path $OutputDirectory "VALENTINA.exe"
 Push-Location $copy
 try {
   & go build -tags embed_frontend -ldflags="-X main.version=valentina-dev" -o $exe .
   if ($LASTEXITCODE -ne 0) { throw "Go build failed." }
 } finally { Pop-Location }
 Write-Host "VALENTINA built locally: $exe" -ForegroundColor Green
} finally { if (Test-Path $tmp) { Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue } }
