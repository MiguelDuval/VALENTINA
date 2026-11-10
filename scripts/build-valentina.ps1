[CmdletBinding()]
param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("Foundation","Full")]
    [string]$Profile,
    [string]$OutputDirectory = ""
)
$ErrorActionPreference = "Stop"

trap {
    Write-Host "UNHANDLED ERROR: $_" -ForegroundColor Red
    if ($env:GITHUB_STEP_SUMMARY) {
        Add-Content -Path $env:GITHUB_STEP_SUMMARY -Value ""
        Add-Content -Path $env:GITHUB_STEP_SUMMARY -Value "### Build Error"
        Add-Content -Path $env:GITHUB_STEP_SUMMARY -Value ""
        Add-Content -Path $env:GITHUB_STEP_SUMMARY -Value "Error: $_"
        Add-Content -Path $env:GITHUB_STEP_SUMMARY -Value ""
        Add-Content -Path $env:GITHUB_STEP_SUMMARY -Value "Profile: $Profile"
        Add-Content -Path $env:GITHUB_STEP_SUMMARY -Value "Commit: $sourceCommitSha"
    }
    exit 1
}

$validationResults = @{}
function Throw-BuildError([string]$message) {
    Write-Host "BUILD ERROR: $message" -ForegroundColor Red
    Throw $message
}
$Repo = Split-Path $PSScriptRoot -Parent
$Upstream = Join-Path $Repo "upstream\huginn"
$Patch1 = Join-Path $Repo "patches\upstream-ui\001-valentina-ui-foundation.patch"
$Patch2 = Join-Path $Repo "patches\upstream-ui\002-valentina-crm-integration.patch"
$ExpectedUpstreamSha = "f01c1d84c08b1de972d8cc268da25602598adff8"
if ([string]::IsNullOrWhiteSpace($OutputDirectory)) { $OutputDirectory = Join-Path $Repo "dist" }
elseif (-not [IO.Path]::IsPathRooted($OutputDirectory)) { $OutputDirectory = Join-Path $Repo $OutputDirectory }
if (Test-Path $OutputDirectory) { Remove-Item -LiteralPath $OutputDirectory -Recurse -Force }
$packageDir = Join-Path $OutputDirectory "package"
New-Item -ItemType Directory -Path $packageDir -Force | Out-Null
foreach ($tool in @("git","npm","go","certutil")) { if (-not (Get-Command $tool -ErrorAction SilentlyContinue)) { Throw-BuildError "Build step failed" } }
if (-not (Test-Path $Upstream)) { Throw-BuildError "Build step failed" }
$actualUpstream = (& git -C $Upstream rev-parse HEAD).Trim()
if ($LASTEXITCODE -ne 0 -or $actualUpstream -ne $ExpectedUpstreamSha) { Throw-BuildError "Build step failed" }
Write-Host "Upstream SHA verified: $actualUpstream" -ForegroundColor Green
$sourceCommitSha = (& git -C $Repo rev-parse HEAD).Trim()
$sourceParentSha = (& git -C $Repo rev-parse HEAD^).Trim()
$sourceBranch = $env:GITHUB_REF_NAME
if ([string]::IsNullOrWhiteSpace($sourceBranch)) { $sourceBranch = (& git -C $Repo rev-parse --abbrev-ref HEAD).Trim() }
$dirtyFlag = $false
$porcelain = & git -C $Repo status --porcelain --untracked-files=all
if ($porcelain) { $dirtyFlag = $true }
$buildTimestamp = [DateTime]::UtcNow.ToString("o")
$goVersion = (& go version).Trim()
$nodeVersion = (& node --version).Trim()
$npmVersion = (& npm --version).Trim()
$patches = @($Patch1)
if ($Profile -eq "Full") {
    if (-not (Test-Path $Patch2)) { Throw-BuildError "Build step failed" }
    $patches = @($Patch1, $Patch2)
}
$tmp = Join-Path ([IO.Path]::GetTempPath()) ("valentina-build-" + [guid]::NewGuid().ToString("N"))
$copy = Join-Path $tmp "source"
New-Item -ItemType Directory -Path $copy -Force | Out-Null
try {
    Write-Host "DEBUG: Step 1 - Copying upstream" -ForegroundColor Cyan; Write-Host "Copying upstream to $copy" -ForegroundColor Cyan
    & robocopy $Upstream $copy /E /XD .git /XF .git /R:1 /W:1 /NFL /NDL /NJH /NJS /NP
    $rc = $LASTEXITCODE
    if ($rc -gt 7) { Throw-BuildError "Build step failed with exit code $rc" }
    & git -C $copy init --quiet
    if ($LASTEXITCODE -ne 0) { Throw-BuildError "Build step failed" }
    foreach ($patch in $patches) {
        Write-Host "DEBUG: Step 3 - Applying patch: $patch" -ForegroundColor Cyan; Write-Host "Applying patch: $patch" -ForegroundColor Cyan
        & git -C $copy apply --check $patch
        if ($LASTEXITCODE -ne 0) { Throw-BuildError "Build step failed" }
        & git -C $copy apply $patch
        if ($LASTEXITCODE -ne 0) { Throw-BuildError "Build step failed" }
    }
    $validationResults.patch_application = @{ status = "passed"; command = "git apply --check <patches>"; result = "applied successfully" }
    $validationResults.upstream_pin_verification = @{ status = "passed"; command = "git -C upstream/huginn rev-parse HEAD"; result = "matched pinned SHA" }
    if ($Profile -eq "Full") {
        Write-Host "DEBUG: Step 5 - Copying CRM sources" -ForegroundColor Cyan; Write-Host "Copying CRM Go source and Vue views for Full profile" -ForegroundColor Cyan
        $destValentina = Join-Path $copy "internal\valentina"
        New-Item -ItemType Directory -Path $destValentina -Force | Out-Null
        $destDomain = Join-Path $destValentina "domain"
        New-Item -ItemType Directory -Path $destDomain -Force | Out-Null
        & robocopy (Join-Path $Repo "valentina\domain") $destDomain /E /R:1 /W:1 /NFL /NDL /NJH /NJS /NP
        $rc = $LASTEXITCODE
    if ($rc -gt 7) { Throw-BuildError "Build step failed with exit code $rc" }
        $destApi = Join-Path $destValentina "api"
        New-Item -ItemType Directory -Path $destApi -Force | Out-Null
        & robocopy (Join-Path $Repo "valentina\api") $destApi /E /R:1 /W:1 /NFL /NDL /NJH /NJS /NP
        $rc = $LASTEXITCODE
    if ($rc -gt 7) { Throw-BuildError "Build step failed with exit code $rc" }
        $destViews = Join-Path $copy "web\src\views"
        New-Item -ItemType Directory -Path $destViews -Force | Out-Null
        & robocopy (Join-Path $Repo "valentina\web\src\views") $destViews /E /R:1 /W:1 /NFL /NDL /NJH /NJS /NP
        $rc = $LASTEXITCODE
    if ($rc -gt 7) { Throw-BuildError "Build step failed with exit code $rc" }
        Write-Host "CRM sources copied successfully." -ForegroundColor Green
        $destShims = Join-Path $copy "web\src\shims-vue.d.ts"
        $sourceShims = Join-Path $Repo "valentina\web\src\shims-vue.d.ts"
        New-Item -ItemType Directory -Path (Split-Path $destShims) -Force | Out-Null
        if (-not (Test-Path -LiteralPath $sourceShims -PathType Leaf)) {
            Throw-BuildError "CRM Vue shim file is missing: $sourceShims"
        }
        Copy-Item -LiteralPath $sourceShims -Destination $destShims -Force
        if (-not (Test-Path -LiteralPath $destShims -PathType Leaf)) {
            Throw-BuildError "CRM Vue shim copy failed: $destShims"
        }
    }
    $webDir = Join-Path $copy "web"
    Write-Host "DEBUG: Step 6 - npm ci" -ForegroundColor Cyan; Write-Host "Installing frontend dependencies..." -ForegroundColor Cyan
    Push-Location $webDir
    & npm ci --no-audit --no-fund 2>&1
    $npmCiExitCode = $LASTEXITCODE
    Pop-Location
    if ($npmCiExitCode -ne 0) { Throw-BuildError "Build step failed" }
    $validationResults.npm_ci = @{ status = "passed"; command = "npm ci --no-audit --no-fund"; result = "exit code 0" }
    Push-Location $webDir
    if (Test-Path "package.json") {
        $pkg = Get-Content "package.json" -Raw | ConvertFrom-Json
        if ($pkg.scripts -and $pkg.scripts.PSObject.Properties.Name -contains "test") {
            Write-Host "DEBUG: Step 7 - npm test" -ForegroundColor Cyan; Write-Host "Running frontend tests..." -ForegroundColor Cyan
            & npm test 2>&1
            $frontendTestExitCode = $LASTEXITCODE
            Pop-Location
            if ($frontendTestExitCode -ne 0) { Throw-BuildError "Build step failed" }
            Write-Host "Frontend tests passed." -ForegroundColor Green
            $validationResults.frontend_tests = @{ status = "passed"; command = "npm test"; result = "exit code 0" }
        } else {
            Pop-Location
        }
    } else {
        Pop-Location
    }
    Push-Location $webDir
    Write-Host "DEBUG: Step 8 - vue-tsc" -ForegroundColor Cyan; Write-Host "Running Vue typecheck..." -ForegroundColor Cyan
    & npx vue-tsc -b 2>&1
    $vueTypecheckExitCode = $LASTEXITCODE
    Pop-Location
    if ($vueTypecheckExitCode -ne 0) { Throw-BuildError "Build step failed" }
    Write-Host "Vue typecheck passed." -ForegroundColor Green
    $validationResults.vue_typecheck = @{ status = "passed"; command = "vue-tsc -b"; result = "exit code 0" }
    Push-Location $webDir
    Write-Host "DEBUG: Step 9 - npm run build" -ForegroundColor Cyan; Write-Host "Running frontend production build..." -ForegroundColor Cyan
    & npm run build 2>&1
    $frontendBuildExitCode = $LASTEXITCODE
    Pop-Location
    if ($frontendBuildExitCode -ne 0) { Throw-BuildError "Build step failed" }
    Write-Host "Frontend production build passed." -ForegroundColor Green
    $validationResults.frontend_production_build = @{ status = "passed"; command = "npm run build"; result = "exit code 0" }
    Write-Host "DEBUG: Step 4 - Patches applied" -ForegroundColor Green; Write-Host "All patches applied successfully." -ForegroundColor Green
    Push-Location $copy
    if ($Profile -eq "Full") {
        Write-Host "DEBUG: Step 10 - Go domain tests" -ForegroundColor Cyan; Write-Host "Running Go domain tests..." -ForegroundColor Cyan
        & go test ./internal/valentina/... 2>&1
        $goDomainTestExitCode = $LASTEXITCODE
        if ($goDomainTestExitCode -ne 0) { Throw-BuildError "Build step failed" }
        Write-Host "Go domain tests passed." -ForegroundColor Green
        $validationResults.go_domain_tests = @{ status = "passed"; command = "go test ./internal/valentina/..."; result = "exit code 0" }
        Write-Host "Running Go server integration tests..." -ForegroundColor Cyan
        & go test ./internal/server 2>&1
        $goServerTestExitCode = $LASTEXITCODE
        if ($goServerTestExitCode -ne 0) { Throw-BuildError "Build step failed" }
        Write-Host "Go server integration tests passed." -ForegroundColor Green
        $validationResults.server_integration_tests = @{ status = "passed"; command = "go test ./internal/server"; result = "exit code 0" }
    } else {
        Write-Host "Skipping Go tests (Foundation profile)" -ForegroundColor Yellow
    }
    Write-Host "DEBUG: Step 11 - Go build" -ForegroundColor Cyan; Write-Host "Compiling Windows executable..." -ForegroundColor Cyan
    & go build -tags embed_frontend -o "$packageDir/VALENTINA.exe" 2>&1
    $exeBuildExitCode = $LASTEXITCODE
    Pop-Location
    if ($exeBuildExitCode -ne 0) { Throw-BuildError "Build step failed" }
    $validationResults.windows_executable_compilation = @{ status = "passed"; command = "go build -tags embed_frontend -o VALENTINA.exe"; result = "exit code 0" }
    $exePath = Join-Path $packageDir "VALENTINA.exe"
    if (-not (Test-Path $exePath)) { Throw-BuildError "Build step failed" }
    $exeSize = (Get-Item -LiteralPath $exePath).Length
    if ($exeSize -le 0) { Throw-BuildError "Build step failed" }
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
Profile: $Profile
Upstream Huginn SHA: $ExpectedUpstreamSha
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
    $manifest = @{
        schema_version = 1
        product_name = "VALENTINA"
        product_version = "1.0.0"
        source = @{
            repository = "https://github.com/MiguelDuval/VALENTINA.git"
            branch = $sourceBranch
            full_commit_sha = $sourceCommitSha
            parent_commit_sha = $sourceParentSha
            dirty_working_tree = $dirtyFlag
            build_timestamp_utc = $buildTimestamp
        }
        upstream = @{
            repository = "https://github.com/scrypster/huginn"
            exact_pinned_sha = $ExpectedUpstreamSha
            verified = $true
        }
        build_environment = @{
            operating_system = "Windows"
            architecture = "x64"
            go_version = $goVersion
            node_version = $nodeVersion
            npm_version = $npmVersion
            build_workflow_name = "VALENTINA Windows test build"
            github_actions_run_id = $env:GITHUB_RUN_ID
            github_actions_run_attempt = $env:GITHUB_RUN_ATTEMPT
            github_actions_run_url = "${env:GITHUB_SERVER_URL}/${env:GITHUB_REPOSITORY}/actions/runs/${env:GITHUB_RUN_ID}"
            github_repository = $env:GITHUB_REPOSITORY
        }
        validation = $validationResults
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
            github_actions_run_url = "${env:GITHUB_SERVER_URL}/${env:GITHUB_REPOSITORY}/actions/runs/${env:GITHUB_RUN_ID}"
        }
    }

    $manifest.validation.upstream_pin_verification = @{
        status = "passed"
        command = "git -C upstream/huginn rev-parse HEAD"
        result = "matched pinned SHA"
    }
    $manifest.validation.patch_application = @{
        status = "passed"
        command = "git apply --check <patches>"
        result = "applied successfully"
    }
    $manifestPath = Join-Path $packageDir "VALENTINA-BUILD-MANIFEST.json"
    $manifest | ConvertTo-Json -Depth 10 | Out-File -FilePath $manifestPath -Encoding utf8
    $zipPath = Join-Path $OutputDirectory ("VALENTINA-windows-x64-" + $sourceCommitSha + ".zip")
    if (Test-Path $zipPath) { Remove-Item -LiteralPath $zipPath -Force }
    Compress-Archive -Path "$packageDir*" -DestinationPath $zipPath -Force
    if ($LASTEXITCODE -ne 0) { Throw-BuildError "Build step failed" }
    Write-Host "Build completed successfully." -ForegroundColor Green
    Write-Host "Profile: $Profile" -ForegroundColor Green
    Write-Host "Artifact: $zipPath" -ForegroundColor Green
    Write-Host "EXE SHA256: $exeSha256" -ForegroundColor Green
} finally {
    if (Test-Path $tmp) { Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue }
}