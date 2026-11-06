[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

function Invoke-Git {
    param([Parameter(Mandatory = $true)][string[]]$GitArgs)
    & git @GitArgs
    if ($LASTEXITCODE -ne 0) {
        throw "Git command failed: git $($GitArgs -join ' ')"
    }
}

$repoRoot = (& git rev-parse --show-toplevel 2>$null)
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($repoRoot)) {
    throw 'Run this script from inside a Git working tree.'
}
$repoRoot = $repoRoot.Trim()
Set-Location -LiteralPath $repoRoot

$branch = (& git branch --show-current).Trim()
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($branch)) {
    throw 'Could not determine the current branch.'
}
if ($branch -eq 'main' -or $branch -notmatch '^(feature|fix|chore|docs|audit)/') {
    throw "Refusing to publish branch '$branch'. Use an approved task branch; direct main publishing is prohibited."
}

$remote = (& git remote get-url origin 2>$null)
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($remote)) {
    throw 'The origin remote is missing or unavailable.'
}
$remote = $remote.Trim() -replace '\.git$', ''
$remote = $remote -replace '^git@github\.com:', 'https://github.com/'
$remote = $remote -replace '^ssh://git@github\.com/', 'https://github.com/'
if ($remote.TrimEnd('/') -ine 'https://github.com/MiguelDuval/VALENTINA') {
    throw "Refusing to push to unexpected origin: $remote"
}

$upstreamPath = Join-Path $repoRoot 'upstream\huginn'
if (-not (Test-Path -LiteralPath $upstreamPath)) {
    throw 'Huginn submodule is not initialized. Run scripts/bootstrap.ps1 and verify the pin before publishing.'
}
$expectedHuginn = 'f01c1d84c08b1de972d8cc268da25602598adff8'
$actualHuginn = (& git -C $upstreamPath rev-parse HEAD 2>$null)
if ($LASTEXITCODE -ne 0 -or $actualHuginn.Trim() -ne $expectedHuginn) {
    throw "Huginn submodule pin mismatch. Expected $expectedHuginn; got $actualHuginn"
}

$status = (& git status --porcelain --untracked-files=all)
if ($LASTEXITCODE -ne 0) {
    throw 'Could not inspect the working tree.'
}
if (-not [string]::IsNullOrWhiteSpace(($status -join [Environment]::NewLine))) {
    throw "Refusing to publish a dirty working tree. Commit or resolve intended changes first: $($status -join [Environment]::NewLine)"
}

$tracking = (& git rev-parse --abbrev-ref --symbolic-full-name '@{u}' 2>$null)
$trackingExit = $LASTEXITCODE
if ($trackingExit -eq 0) {
    $tracking = $tracking.Trim()
    if ($tracking -ine "origin/$branch") {
        throw "Refusing to push: current upstream is '$tracking', expected 'origin/$branch'."
    }
    Invoke-Git -GitArgs @('push', 'origin', "HEAD:refs/heads/$branch")
}
else {
    Invoke-Git -GitArgs @('push', '--set-upstream', 'origin', 'HEAD')
}

Write-Host "Published $branch to origin. No force push, pull, rebase or merge was performed." -ForegroundColor Green
