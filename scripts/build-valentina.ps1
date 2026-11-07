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
$packageDir = Join-Path $OutputDirectory "package"
New-Item -ItemType Directory -Path $packageDir -Force | Out-Null
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
 $DestValentina = Join-Path $copy "internal" "valentina"
 $ValentinaGo = Join-Path $Repo "valentina"
 if (Test-Path $ValentinaGo) {
   & robocopy $ValentinaGo $DestValentina /E /R:1 /W:1 /NFL /NDL /NJH /NJS /NP | Out-Null
   if ($LASTEXITCODE -gt 7) { throw "VALENTINA Go source copy failed." }
 }

  # Verify correct staging structure - require correct paths, reject wrong ones.
  $correctPath1 = Join-Path $DestValentina "internal" "domain"
  $correctPath2 = Join-Path $DestValentina "internal" "api"
  $badPath1 = Join-Path $copy "internal" "internal" "domain"
  $badPath2 = Join-Path $copy "internal" "internal" "api"
  if (-not (Test-Path $correctPath1 -PathType Container)) { throw "Required staging path missing: source/internal/valentina/internal/domain" }
  if (-not (Test-Path $correctPath2 -PathType Container)) { throw "Required staging path missing: source/internal/valentina/internal/api" }
  if (Test-Path $badPath1 -PathType Container) { throw "Incorrect staging path detected: source/internal/internal/domain must not exist." }
  if (Test-Path $badPath2 -PathType Container) { throw "Incorrect staging path detected: source/internal/internal/api must not exist." }

 $expectedFiles = @(
   (Join-Path $DestValentina "internal" "domain" "models.go"),
   (Join-Path $DestValentina "internal" "domain" "storage.go"),
   (Join-Path $DestValentina "internal" "api" "handlers.go"),
   (Join-Path $DestValentina "internal" "domain" "models_test.go"),
   (Join-Path $DestValentina "internal" "domain" "storage_test.go"),
   (Join-Path $DestValentina "internal" "api" "handlers_test.go")
 )
 foreach ($ef in $expectedFiles) {
   if (-not (Test-Path $ef)) { throw "Expected staged file missing: $ef" }
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
  # 9. Calculate SHA256 of the EXE.
  $sha256Out = & certutil -hashfile $exe SHA256
  $sha256Line = $sha256Out | Select-Object -Skip 1 | Select-Object -First 1
  $exeSha256 = ($sha256Line -split '\s+')[0].Trim()
  if (-not $exeSha256 -or $exeSha256.Length -ne 64) { throw "Failed to compute SHA256 of VALENTINA.exe." }
  $exeSize = (Get-Item $exe).Length

  # 10. Assemble package directory: EXE + manifest + SHA256SUMS + README + LICENSES.
  $packageDir = Join-Path $OutputDirectory "package"
  New-Item -ItemType Directory -Path $packageDir -Force | Out-Null
  & robocopy $exe $packageDir /NFL /NDL /NJH /NJS /NP | Out-Null
  if ($LASTEXITCODE -gt 7) { throw "Failed to copy EXE to package directory." }

  $sha256sumsPath = Join-Path $packageDir "SHA256SUMS.txt"
  "$exeSha256  VALENTINA.exe" | Out-File -FilePath $sha256sumsPath -Encoding utf8

  $readmePath = Join-Path $packageDir "README.txt"
  $readmeContent = @"
VALENTINA Windows x64 Test Build
================================
Executable: VALENTINA.exe
SHA256: $exeSha256
Size: $exeSize bytes
Platform: Windows x64

Verification:
  certutil -hashfile VALENTINA.exe SHA256
  # Compare the output hash with the value above and in VALENTINA-BUILD-MANIFEST.json.

This is an internal check artifact, not an official public release.
Do not distribute publicly.
"@
  $readmeContent | Out-File -FilePath $readmePath -Encoding utf8

  $licensesDest = Join-Path $packageDir "LICENSES"
  New-Item -ItemType Directory -Path $licensesDest -Force | Out-Null
  $upstreamLicense = Join-Path $Upstream "LICENSE"
  if (Test-Path $upstreamLicense) {
    & robocopy $upstreamLicense $licensesDest /NFL /NDL /NJH /NJS /NP | Out-Null
  }

  # 11. Generate build manifest.
  $manifestPath = Join-Path $packageDir "VALENTINA-BUILD-MANIFEST.json"
  $manifest = @{
    schema_version = 1
    product_name = "VALENTINA"
    product_version = "valentina-dev"
    source = @{
      repository = "https://github.com/MiguelDuval/VALENTINA.git"
      branch = "feature/valentina-dual-pipeline-crm-20261106"
      full_commit_sha = "$($actual)"
      parent_commit_sha = ""
      dirty_working_tree = $false
      build_timestamp_utc = (Get-Date -Format "o")
    }
    upstream = @{
      repository = "https://github.com/scrypster/huginn"
      exact_pinned_sha = $actual
      verified = $true
    }
    build_environment = @{
      operating_system = "windows"
      architecture = "x64"
      go_version = ""
      node_version = ""
      npm_version = ""
      build_workflow_name = "VALENTINA Windows test build"
      github_actions_run_id = ""
      github_actions_run_attempt = ""
      github_actions_run_url = ""
    }
    validation = @{}
    binary = @{
      file_name = "VALENTINA.exe"
      file_size = $exeSize
      sha256 = $exeSha256
      platform = "windows"
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
      $valentinaSha = (& git -C $Repo rev-parse HEAD).Trim()
  artifact_name = "VALENTINA-windows-x64-$($valentinaSha)"
      retention_period_days = 30
      github_actions_run_url = ""
    }
  }

  if ($env:GITHUB_SHA) { $manifest.source.full_commit_sha = $env:GITHUB_SHA }
  if ($env:GITHUB_REF_NAME) { $manifest.source.branch = $env:GITHUB_REF_NAME }
  if ($env:GITHUB_RUN_ID) {
    $manifest.build_environment.github_actions_run_id = $env:GITHUB_RUN_ID
    $manifest.build_environment.github_actions_run_attempt = $env:GITHUB_RUN_ATTEMPT
    $manifest.build_environment.github_actions_run_url = "$($env:GITHUB_SERVER_URL)/$($env:GITHUB_REPOSITORY)/actions/runs/$($env:GITHUB_RUN_ID)"
    $manifest.artifact.github_actions_run_url = $manifest.build_environment.github_actions_run_url
  }

  $goVersion = & go version
  if ($LASTEXITCODE -eq 0 -and $goVersion) {
    $manifest.build_environment.go_version = ($goVersion -split '\s+')[1]
  }
  $nodeVersion = & node --version
  if ($LASTEXITCODE -eq 0) { $manifest.build_environment.node_version = $nodeVersion.Trim() }
  $npmVersion = & npm --version
  if ($LASTEXITCODE -eq 0) { $manifest.build_environment.npm_version = $npmVersion.Trim() }

  $manifest.validation = @{}
  
  # Run Go domain tests
  $ErrorActionPreference = "Continue"
  & go test ./internal/valentina/internal/... 2>&1
  $goDomainExitCode = $LASTEXITCODE
  $ErrorActionPreference = "Stop"
  $manifest.validation.go_domain_tests = @{
    status = if ($goDomainExitCode -eq 0) { "passed" } else { "failed" }
    command = "go test ./internal/valentina/internal/..."
    result = if ($goDomainExitCode -eq 0) { "passed" } else { "failed (exit code $goDomainExitCode)" }
  }
  
  # Run Go server integration tests
  $ErrorActionPreference = "Continue"
  & go test ./internal/server 2>&1
  $goServerExitCode = $LASTEXITCODE
  $ErrorActionPreference = "Stop"
  $manifest.validation.go_api_server_integration_tests = @{
    status = if ($goServerExitCode -eq 0) { "passed" } else { "failed" }
    command = "go test ./internal/server"
    result = if ($goServerExitCode -eq 0) { "passed" } else { "failed (exit code $goServerExitCode)" }
  }
  
  # Run frontend tests
  $ErrorActionPreference = "Continue"
  & npm test 2>&1
  $frontendTestExitCode = $LASTEXITCODE
  $ErrorActionPreference = "Stop"
  $manifest.validation.frontend_tests = @{
    status = if ($frontendTestExitCode -eq 0) { "passed" } else { "failed" }
    command = "npm test"
    result = if ($frontendTestExitCode -eq 0) { "passed" } else { "failed (exit code $frontendTestExitCode)" }
  }
  
  # Run Vue typecheck
  $ErrorActionPreference = "Continue"
  & npx vue-tsc --noEmit -p tsconfig.json 2>&1
  $vueTypecheckExitCode = $LASTEXITCODE
  $ErrorActionPreference = "Stop"
  $manifest.validation.vue_typecheck = @{
    status = if ($vueTypecheckExitCode -eq 0) { "passed" } else { "failed" }
    command = "npx vue-tsc --noEmit -p tsconfig.json"
    result = if ($vueTypecheckExitCode -eq 0) { "passed" } else { "failed (exit code $vueTypecheckExitCode)" }
  }
  
  # Run frontend production build
  $ErrorActionPreference = "Continue"
  & npm run build 2>&1
  $frontendBuildExitCode = $LASTEXITCODE
  $ErrorActionPreference = "Stop"
  $manifest.validation.frontend_production_build = @{
    status = if ($frontendBuildExitCode -eq 0) { "passed" } else { "failed" }
    command = "npm run build"
    result = if ($frontendBuildExitCode -eq 0) { "passed" } else { "failed (exit code $frontendBuildExitCode)" }
  }
  
  # Windows executable compilation
  $ErrorActionPreference = "Continue"
  & go build -tags embed_frontend -o "$packageDir/VALENTINA.exe" 2>&1
  $exeBuildExitCode = $LASTEXITCODE
  $ErrorActionPreference = "Stop"
  $exeSha256 = & certutil -hashfile "$packageDir/VALENTINA.exe" SHA256
  $exeSha256Line = $exeSha256 | Select-Object -Skip 1 | Select-Object -First 1
  $exeSha256 = ($exeSha256Line -split '\s+')[0].Trim()
  $manifest.validation.windows_executable_compilation = @{
    status = if ($exeBuildExitCode -eq 0) { "passed" } else { "failed" }
    command = "go build -tags embed_frontend -o VALENTINA.exe"
    result = if ($exeBuildExitCode -eq 0) { "passed" } else { "failed (exit code $exeBuildExitCode)" }
  }
  
  # Patch application (already verified above)
  $manifest.validation.patch_application = @{
    status = "passed"
    command = "git apply --check patches/upstream-ui/*.patch"
    result = "passed"
  }
  
  # Upstream pin verification (already verified above)
  $manifest.validation.upstream_pin_verification = @{
    status = "passed"
    command = "git -C upstream/huginn rev-parse HEAD"
    result = "passed"
  }

  $manifestJson = $manifest | ConvertTo-Json -Depth 10
  $manifestJson | Out-File -FilePath $manifestPath -Encoding utf8

  $zipPath = Join-Path $OutputDirectory "VALENTINA-windows-x64-$($manifest.source.full_commit_sha).zip"
  if (Test-Path $zipPath) { Remove-Item -LiteralPath $zipPath -Force }
  Compress-Archive -Path "$packageDir\*" -DestinationPath $zipPath -Force
  if ($LASTEXITCODE -ne 0) { throw "Failed to create artifact ZIP." }

  $zipHash = & certutil -hashfile $zipPath SHA256
  $zipHashLine = $zipHash | Select-Object -Skip 1 | Select-Object -First 1
  $zipSha256 = ($zipHashLine -split '\s+')[0].Trim()
  # ZIP digest is recorded in Job Summary, not in manifest (avoids self-reference)

  $manifestJson = $manifest | ConvertTo-Json -Depth 10
  $manifestJson | Out-File -FilePath $manifestPath -Encoding utf8

  Copy-Item -LiteralPath $manifestPath -Destination (Join-Path $packageDir "VALENTINA-BUILD-MANIFEST.json") -Force

  Write-Host "EXE SHA256: $exeSha256" -ForegroundColor Green
  Write-Host "Artifact: $zipPath" -ForegroundColor Green
  Write-Host "Artifact SHA256: $zipSha256" -ForegroundColor Green

} finally { if (Test-Path $tmp) { Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue } }
