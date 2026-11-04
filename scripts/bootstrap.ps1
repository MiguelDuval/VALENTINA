$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$expectedCommit = "f01c1d84c08b1de972d8cc268da25602598adff8"
$submodulePath = Join-Path $repoRoot "upstream/huginn"

Push-Location $repoRoot
try {
    Write-Host "Synchronizing and initializing VALENTINA upstream submodules..."
    & git submodule sync --recursive
    if ($LASTEXITCODE -ne 0) { throw "git submodule sync failed." }

    & git submodule update --init --recursive
    if ($LASTEXITCODE -ne 0) { throw "git submodule update failed." }

    if (-not (Test-Path (Join-Path $submodulePath ".git"))) {
        throw "Huginn submodule was not initialized at upstream/huginn."
    }

    $actualCommit = (& git -C $submodulePath rev-parse HEAD).Trim()
    if ($LASTEXITCODE -ne 0) { throw "Could not read the Huginn submodule HEAD." }

    if ($actualCommit -ne $expectedCommit) {
        throw "Unexpected Huginn revision. Expected $expectedCommit, got $actualCommit."
    }

    Write-Host "Huginn source is ready at upstream/huginn."
    Write-Host "Pinned revision: $actualCommit"
}
finally {
    Pop-Location
}
