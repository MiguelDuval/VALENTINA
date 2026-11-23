[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$ExecutablePath
)
$ErrorActionPreference = 'Stop'
$exe = (Resolve-Path -LiteralPath $ExecutablePath -ErrorAction Stop).Path
$root = Join-Path ([IO.Path]::GetTempPath()) ('valentina-runtime-smoke-' + [guid]::NewGuid().ToString('N'))
$probe = [Net.Sockets.TcpListener]::new([Net.IPAddress]::Loopback, 0)
$probe.Start()
$port = ([Net.IPEndPoint]$probe.LocalEndpoint).Port
$probe.Stop()
foreach ($folder in @('.huginn', 'appdata/roaming', 'appdata/local')) {
    New-Item -ItemType Directory -Path (Join-Path $root $folder) -Force | Out-Null
}
$cfg = @{
    version = 13
    web_ui = @{ enabled = $true; port = $port; auto_open = $false; bind = '127.0.0.1' }
    scheduler_enabled = $false
    tools_enabled = $false
    git_stage_on_write = $false
    mcp_servers = @()
} | ConvertTo-Json -Depth 5
[IO.File]::WriteAllText((Join-Path $root '.huginn/config.json'), $cfg)
$env:USERPROFILE = $root
$env:HOME = $root
$env:HUGINN_HOME = $root
$env:APPDATA = Join-Path $root 'appdata/roaming'
$env:LOCALAPPDATA = Join-Path $root 'appdata/local'
$env:VALENTINA_INTAKE_HMAC_SECRET = 'SYNTHETIC-TEST-ONLY-NOT-A-PRODUCTION-SECRET-2026'
$env:VALENTINA_INTAKE_ALLOW_LEGACY_V1 = ''
$env:NVIDIA_NIM_API_KEY = ''
$base = 'http://127.0.0.1:' + $port
$intakeURL = $base + '/api/v1/valentina/intake'
$serverA = $null
$serverB = $null

function Start-SmokeServer([int]$cycle) {
    $out = Join-Path $root ('stdout-' + $cycle + '.log')
    $err = Join-Path $root ('stderr-' + $cycle + '.log')
    $p = Start-Process -FilePath $exe -ArgumentList @('-no-tray', 'serve', '--foreground') -WorkingDirectory $root -RedirectStandardOutput $out -RedirectStandardError $err -PassThru
    try {
        for ($i = 0; $i -lt 25; $i++) {
            Start-Sleep -Milliseconds 650
            $p.Refresh()
            if ($p.HasExited) { throw ('VALENTINA exited before ready, exit=' + $p.ExitCode) }
            try {
                $response = Invoke-WebRequest -Uri ($base + '/api/v1/valentina/health') -UseBasicParsing -TimeoutSec 2
                if ($response.StatusCode -eq 200) { return $p }
            } catch { }
        }
        throw 'VALENTINA failed to become ready on localhost'
    } catch {
        if (Get-Process -Id $p.Id -ErrorAction SilentlyContinue) {
            Stop-Process -Id $p.Id -ErrorAction SilentlyContinue
        }
        throw
    }
}
function Signed-Intake([string]$body, [bool]$badSignature = $false, [long]$unixSeconds = 0) {
    if ($unixSeconds -eq 0) { $unixSeconds = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds() }
    $ts = [string]$unixSeconds
    $signed = [string]::Concat('v2', [char]10, $ts, [char]10, $body)
    $mac = New-Object System.Security.Cryptography.HMACSHA256
    $mac.Key = [Text.Encoding]::UTF8.GetBytes($env:VALENTINA_INTAKE_HMAC_SECRET)
    try {
        $sig = [BitConverter]::ToString($mac.ComputeHash([Text.Encoding]::UTF8.GetBytes($signed))).Replace('-', '').ToLowerInvariant()
    } finally {
        $mac.Dispose()
    }
    if ($badSignature) { $sig = '0' * 64 }
    $headers = @{
        'X-Valentina-Signature-Version' = 'v2'
        'X-Valentina-Timestamp' = $ts
        'X-Valentina-Signature' = $sig
    }
    try {
        $r = Invoke-WebRequest -Uri $intakeURL -Method Post -Headers $headers -Body ([Text.Encoding]::UTF8.GetBytes($body)) -ContentType 'application/json' -UseBasicParsing -TimeoutSec 8
        $payload = $r.Content | ConvertFrom-Json
        return [pscustomobject]@{ Code = [int]$r.StatusCode; LeadID = [string]$payload.lead_id }
    } catch {
        if (-not $_.Exception.Response) { throw }
        return [pscustomobject]@{ Code = [int]$_.Exception.Response.StatusCode; LeadID = '' }
    }
}
function Require-Result([bool]$condition, [string]$message) {
    if (-not $condition) { throw ('VALENTINA runtime smoke failed: ' + $message) }
}
try {
    $serverA = Start-SmokeServer 1
    $rootResponse = Invoke-WebRequest -Uri ($base + '/') -UseBasicParsing -TimeoutSec 5
    Require-Result ($rootResponse.StatusCode -eq 200 -and $rootResponse.Content -match 'VALENTINA') 'branded web UI root'
    $body = '{"request_summary":"Synthetic bath repair","inbound_channel":"website","consent":true,"privacy_notice":true,"privacy_notice_version":"2026-11","idempotency_key":"synthetic-runtime-request-001","contact_email":"customer@example.test"}'
    $first = Signed-Intake $body
    $replay = Signed-Intake $body
    $conflict = Signed-Intake ($body.Replace('bath repair', 'kitchen repair'))
    $invalid = Signed-Intake $body $true
    $now = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
    $stale = Signed-Intake $body $false ($now - 601)
    $future = Signed-Intake $body $false ($now + 601)
    Require-Result ($first.Code -eq 201 -and -not [string]::IsNullOrWhiteSpace($first.LeadID)) 'initial signed intake HTTP 201'
    Require-Result ($replay.Code -eq 200 -and $replay.LeadID -eq $first.LeadID) 'idempotent duplicate HTTP 200'
    Require-Result ($conflict.Code -eq 409) 'same key different payload HTTP 409'
    Require-Result ($invalid.Code -eq 401 -and $stale.Code -eq 401 -and $future.Code -eq 401) 'signature and timestamp rejection'
    try {
        $unauthed = Invoke-WebRequest -Uri ($base + '/api/v1/valentina/leads') -UseBasicParsing -TimeoutSec 5
        $otherStatus = [int]$unauthed.StatusCode
    } catch {
        if (-not $_.Exception.Response) { throw }
        $otherStatus = [int]$_.Exception.Response.StatusCode
    }
    Require-Result ($otherStatus -eq 401) 'all other CRM APIs must retain bearer auth'
    $listen = @(Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue)
    Require-Result ($listen.Count -eq 1 -and $listen[0].LocalAddress -eq '127.0.0.1') 'loopback-only bind'
    Stop-Process -Id $serverA.Id -ErrorAction Stop
    $serverA = $null
    Start-Sleep -Milliseconds 800
    $serverB = Start-SmokeServer 2
    $afterRestart = Signed-Intake $body
    Require-Result ($afterRestart.Code -eq 200 -and $afterRestart.LeadID -eq $first.LeadID) 'idempotency persisted across restart'
    Write-Output 'VALENTINA_RUNTIME_SMOKE=PASS (UI 200, intake 201/200/409/401, stale/future 401, CRM auth 401, restart 200, loopback)'
} finally {
    foreach ($p in @($serverA, $serverB)) {
        if ($p -and (Get-Process -Id $p.Id -ErrorAction SilentlyContinue)) {
            Stop-Process -Id $p.Id -ErrorAction SilentlyContinue
        }
    }
    Start-Sleep -Milliseconds 850
    if (Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue) {
        Write-Warning 'Runtime smoke port still bound; check test child process'
    }
    # Only remove the unique disposable test profile that this script created.
    if (Test-Path -LiteralPath $root) {
        Remove-Item -LiteralPath $root -Recurse -Force -ErrorAction SilentlyContinue
    }
}
