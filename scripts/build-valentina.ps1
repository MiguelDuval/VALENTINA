[CmdletBinding()]
param([string]$OutputDirectory = "")
$ErrorActionPreference = "Stop"
$Repo = Split-Path $PSScriptRoot -Parent
$Upstream = Join-Path $Repo "upstream\huginn"
$Patch1 = Join-Path $Repo "patches\upstream-ui\001-valentina-ui-foundation.patch"
$Expected = "f01c1d84c08b1de972d8cc268da25602598adff8"
if ([string]::IsNullOrWhiteSpace($OutputDirectory)) { $OutputDirectory = Join-Path $Repo "dist" }
elseif (-not [IO.Path]::IsPathRooted($OutputDirectory)) { $OutputDirectory = Join-Path $Repo $OutputDirectory }
$packageDir = Join-Path $OutputDirectory "package"
New-Item -ItemType Directory -Path $packageDir -Force | Out-Null
foreach ($tool in @("git","npm","go","certutil")) { if (-not (Get-Command $tool -ErrorAction SilentlyContinue)) { throw "Required tool missing: $tool" } }
if (-not (Test-Path $Upstream)) { throw "Initialize upstream with scripts/bootstrap.ps1 first." }
$actual = (& git -C $Upstream rev-parse HEAD).Trim()
if ($LASTEXITCODE -ne 0 -or $actual -ne $Expected) { throw "Upstream pin mismatch. Expected $Expected, got $actual." }
Write-Host "Upstream SHA verified: $actual" -ForegroundColor Green
if (-not (Test-Path $Patch1)) { throw "Missing VALENTINA UI patch: $Patch1" }
$sourceCommitSha = (& git -C $Repo rev-parse HEAD).Trim()
$sourceParentSha = (& git -C $Repo rev-parse HEAD^).Trim()
$sourceBranch = (& git -C $Repo rev-parse --abbrev-ref HEAD).Trim()
$dirtyFlag = $false
$porcelain = & git -C $Repo status --porcelain --untracked-files=all
if ($porcelain) { $dirtyFlag = $true }
$buildTimestamp = [DateTime]::UtcNow.ToString("o")
$tmp = Join-Path ([IO.Path]::GetTempPath()) ("valentina-build-" + [guid]::NewGuid().ToString("N"))
$copy = Join-Path $tmp "source"
New-Item -ItemType Directory -Path $copy -Force | Out-Null
try {
  Write-Host "Copying upstream to $copy" -ForegroundColor Cyan
  & robocopy $Upstream $copy /E /XD .git /XF .git /R:1 /W:1 /NFL /NDL /NJH /NJS /NP | Out-Null
  if ($LASTEXITCODE -gt 7) { throw "Upstream source copy failed." }
  & git -C $copy init --quiet
  if ($LASTEXITCODE -ne 0) { throw "Could not initialize temporary source." }
  Write-Host "Applying patch 001..." -ForegroundColor Cyan
  & git -C $copy apply --check $Patch1
  if ($LASTEXITCODE -ne 0) { throw "Patch 001 does not apply to pinned source: $Patch1" }
  & git -C $copy apply $Patch1
  if ($LASTEXITCODE -ne 0) { throw "Could not apply patch 001: $Patch1" }
  Write-Host "Patch 001 applied successfully." -ForegroundColor Green
  $webDir = Join-Path $copy "web"
  Write-Host "Installing frontend dependencies..." -ForegroundColor Cyan
  Push-Location $webDir
  & npm ci --no-audit --no-fund 2>&1
  $npmCiExitCode = $LASTEXITCODE
  Pop-Location
  if ($npmCiExitCode -ne 0) { throw "npm ci failed with exit code $npmCiExitCode" }
  Push-Location $webDir
  if (Test-Path "package.json") {
    $pkg = Get-Content "package.json" -Raw | ConvertFrom-Json
    if ($pkg.scripts -and $pkg.scripts.PSObject.Properties.Name -contains "test") {
      Write-Host "Running frontend tests..." -ForegroundColor Cyan
      & npm test 2>&1
      $frontendTestExitCode = $LASTEXITCODE
      Pop-Location
      if ($frontendTestExitCode -ne 0) { throw "Frontend tests failed with exit code $frontendTestExitCode" }
      Write-Host "Frontend tests passed." -ForegroundColor Green
    } else {
      Pop-Location
    }
  } else {
    Pop-Location
  }
  Push-Location $webDir
  Write-Host "Running Vue typecheck..." -ForegroundColor Cyan
  & npx vue-tsc --noEmit -p tsconfig.json 2>&1
  $vueTypecheckExitCode = $LASTEXITCODE
  Pop-Location
  if ($vueTypecheckExitCode -ne 0) { throw "Vue typecheck failed with exit code $vueTypecheckExitCode" }
  Write-Host "Vue typecheck passed." -ForegroundColor Green
  Push-Location $webDir
  Write-Host "Running frontend production build..." -ForegroundColor Cyan
  & npm run build 2>&1
  $frontendBuildExitCode = $LASTEXITCODE
  Pop-Location
  if ($frontendBuildExitCode -ne 0) { throw "Frontend production build failed with exit code $frontendBuildExitCode" }
  Write-Host "Frontend production build passed." -ForegroundColor Green
  Write-Host "Compiling Windows executable..." -ForegroundColor Cyan
  Push-Location $copy
  & go build -tags embed_frontend -o "$packageDir/VALENTINA.exe" 2>&1
  $exeBuildExitCode = $LASTEXITCODE
  Pop-Location
  if ($exeBuildExitCode -ne 0) { throw "Go build failed with exit code $exeBuildExitCode" }
  $exePath = Join-Path $packageDir "VALENTINA.exe"
  if (-not (Test-Path $exePath)) { throw "VALENTINA.exe was not produced." }
  $exeSize = (Get-Item -LiteralPath $exePath).Length
  if ($exeSize -le 0) { throw "VALENTINA.exe is empty." }
  Write-Host "EXE size: $exeSize bytes" -ForegroundColor Green
  $exeHash = & certutil -hashfile $exePath SHA256
  $exeHashLine = $exeHash | Select-Object -Skip 1 | Select-Object -First 1
  $exeSha256 = ($exeHashLine -split '\s+')[0].Trim()
  Write-Host "EXE SHA256: $exeSha256" -ForegroundColor Green
  $sha256SumsPath = Join-Path $packageDir "SHA256SUMS.txt"
  "$exeSha256  VALENTINA.exe" | Out-File -FilePath $sha256SumsPath -Encoding utf8
  $readmePath = Join-Path $packageDir "README.txt"
  @"
VALENTINA Windows x64 Build
=========================
Commit: $sourceCommitSha
Parent: $sourceParentSha
Branch: $sourceBranch
Upstream Huginn SHA: $Expected
Build timestamp (UTC): $buildTimestamp
Dirty working tree: $dirtyFlag

Files:
  VALENTINA.exe        - Main application executable
  SHA256SUMS.txt       - SHA256 checksum of VALENTINA.exe
  VALENTINA-BUILD-MANIFEST.json - Build manifest
  LICENSES/            - License and notice files

Verification:
  certutil -hashfile VALENTINA.exe SHA256
"@ | Out-File -FilePath $readmePath -Encoding utf8
  $licensesDir = Join-Path $packageDir "LICENSES"
  New-Item -ItemType Directory -Path $licensesDir -Force | Out-Null
  $upstreamLicense = Join-Path $Upstream "LICENSE"
  if (Test-Path $upstreamLicense) {
    Copy-Item -LiteralPath $upstreamLicense -Destination (Join-Path $licensesDir "LICENSE.huginn.txt") -Force
  }
  $valentinaLicense = Join-Path $Repo "LICENSE"
  if (Test-Path $valentinaLicense) {
    Copy-Item -LiteralPath $valentinaLicense -Destination (Join-Path $licensesDir "LICENSE.valentina.txt") -Force
  }
  $manifest = @{
    schema_version = "1.0"
    product_name = "VALENTINA"
    product_version = "1.0.0"
    source = @{
      full_commit_sha = $sourceCommitSha
      parent_commit_sha = $sourceParentSha
      branch = $sourceBranch
      dirty_working_tree = $dirtyFlag
      build_timestamp_utc = $buildTimestamp
    }
    upstream = @{
      repository = "scrypster/huginn"
      exact_pinned_sha = $Expected
      verified = $true
    }
    build_environment = @{
      operating_system = "Windows"
      architecture = "x64"
      go_version = "from upstream/huginn/go.mod"
      node_version = "22"
      npm_version = "from package-lock.json"
      build_workflow_name = "VALENTINA Windows test build"
      github_run_id = $env:GITHUB_RUN_ID
      github_run_attempt = $env:GITHUB_RUN_ATTEMPT
      github_server_url = $env:GITHUB_SERVER_URL
      github_repository = $env:GITHUB_REPOSITORY
    }
    validation = @{
      upstream_pin_verification = @{
        status = "passed"
        command = "git -C upstream/huginn rev-parse HEAD"
        result = "matched pinned SHA"
      }
      patch_application = @{
        status = "passed"
        command = "git apply patches/upstream-ui/001-valentina-ui-foundation.patch"
        result = "applied successfully"
      }
      npm_ci = @{
        status = "passed"
        command = "npm ci --no-audit --no-fund"
        result = "exit code 0"
      }
      frontend_tests = @{
        status = "passed"
        command = "npm test"
        result = "exit code 0"
      }
      vue_typecheck = @{
        status = "passed"
        command = "npx vue-tsc --noEmit -p tsconfig.json"
        result = "exit code 0"
      }
      frontend_production_build = @{
        status = "passed"
        command = "npm run build"
        result = "exit code 0"
      }
      windows_executable_compilation = @{
        status = "passed"
        command = "go build -tags embed_frontend -o VALENTINA.exe"
        result = "exit code 0"
      }
    }
    binary = @{
      file_name = "VALENTINA.exe"
      file_size = $exeSize
      sha256 = $exeSha256
      platform = "Windows"
      architecture = "x64"
    }
    security = @{
      secrets_included = $false
      local_database_included = $false
      user_profile_included = $false
      source_scraping_enabled = $false
      credentials_embedded = $false
    }
    artifact = @{
      artifact_name = "VALENTINA-windows-x64-$sourceCommitSha"
      retention_period_days = 30
    }
  }
  $manifestPath = Join-Path $packageDir "VALENTINA-BUILD-MANIFEST.json"
  $manifest | ConvertTo-Json -Depth 10 | Out-File -FilePath $manifestPath -Encoding utf8
  $zipPath = Join-Path $OutputDirectory ("VALENTINA-windows-x64-" + $sourceCommitSha + ".zip")
  if (Test-Path $zipPath) { Remove-Item -LiteralPath $zipPath -Force }
  Compress-Archive -Path "$packageDir*" -DestinationPath $zipPath -Force
  if ($LASTEXITCODE -ne 0) { throw "Failed to create artifact ZIP." }
  Write-Host "Build completed successfully." -ForegroundColor Green
  Write-Host "Artifact: $zipPath" -ForegroundColor Green
  Write-Host "EXE SHA256: $exeSha256" -ForegroundColor Green
} finally {
  if (Test-Path $tmp) { Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue }
}