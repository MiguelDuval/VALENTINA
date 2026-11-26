[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$ExecutablePath,
    # Installer CI retains its synthetic runtime data to verify upgrade/uninstall
    # preservation. These options are forbidden on ordinary workstations.
    [string]$TestProfileRoot,
    [switch]$RetainTestProfile
)
$ErrorActionPreference = 'Stop'
$exe = (Resolve-Path -LiteralPath $ExecutablePath -ErrorAction Stop).Path
$root = Join-Path ([IO.Path]::GetTempPath()) ('valentina-runtime-smoke-' + [guid]::NewGuid().ToString('N'))
if ($TestProfileRoot -or $RetainTestProfile) {
    if (-not $TestProfileRoot -or $env:GITHUB_ACTIONS -ne 'true' -or $env:RUNNER_ENVIRONMENT -ne 'github-hosted' -or -not $env:RUNNER_TEMP) {
        throw 'Retained runtime profiles require a disposable GitHub-hosted runner'
    }
    $root = [IO.Path]::GetFullPath($TestProfileRoot)
    $runnerPrefix = [IO.Path]::GetFullPath($env:RUNNER_TEMP).TrimEnd('\') + '\'
    if (-not $root.StartsWith($runnerPrefix, [StringComparison]::OrdinalIgnoreCase) -or (Test-Path -LiteralPath $root)) {
        throw 'Runtime profile must be a new directory under RUNNER_TEMP'
    }
}
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
    $p = Start-Process -FilePath $exe -ArgumentList @('-no-tray', 'serve', '--foreground') -WorkingDirectory $root -RedirectStandardOutput $out -RedirectStandardError $err -PassThru -WindowStyle Hidden
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
function Invoke-ProjectAPI([string]$method, [string]$path, $body = $null, [int]$expected = 200) {
    $params = @{ Uri = ($base + '/api/v1/valentina' + $path); Method = $method; Headers = $script:crmHeaders; UseBasicParsing = $true; TimeoutSec = 8 }
    if ($null -ne $body) { $params.Body = [Text.Encoding]::UTF8.GetBytes(($body | ConvertTo-Json -Depth 8 -Compress)); $params.ContentType = 'application/json' }
    $response = Invoke-WebRequest @params
    Require-Result ([int]$response.StatusCode -eq $expected) ('project API status: ' + $method + ' ' + $path)
    return ($response.Content | ConvertFrom-Json)
}
function Require-ProjectAuth($endpoints) {
    foreach ($endpoint in $endpoints) {
        try { $null = Invoke-WebRequest -Uri ($base+'/api/v1/valentina'+$endpoint[1]) -Method $endpoint[0] -UseBasicParsing -TimeoutSec 5; $code = 200 }
        catch { if (-not $_.Exception.Response) { throw }; $code = [int]$_.Exception.Response.StatusCode }
        Require-Result ($code -eq 401) ('project route requires bearer: '+$endpoint[0]+' '+$endpoint[1]+'; HTTP '+$code)
    }
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
    # Tokens remain in process memory; never print them or pass them on a command line.
    $localToken = Invoke-RestMethod -Uri ($base + '/api/v1/token') -TimeoutSec 5
    $script:crmHeaders = @{ Authorization = ('Bearer ' + $localToken.token) }
    foreach ($status in @('triage', 'qualified')) {
        $null = Invoke-ProjectAPI 'POST' '/review-decisions' @{entity_type='inbound_lead';entity_id=$first.LeadID;new_status=$status;decision_reason='Synthetic human review'} 201
    }
    $projectBody = @{title='Synthetic connected project';source_lead_id=$first.LeadID;service_scope='Synthetic scope'}
    $project = Invoke-ProjectAPI 'POST' '/projects' $projectBody 201
    $retryProject = Invoke-ProjectAPI 'POST' '/projects' $projectBody 201
    Require-Result ($project.project_id -eq $retryProject.project_id) 'unique lead conversion'
    $projectPath = '/projects/' + $project.project_id
    $linked = Invoke-ProjectAPI 'GET' ('/leads/' + $first.LeadID + '/project')
    Require-Result ($linked.project.project_id -eq $project.project_id) 'lead project link'
    $project = Invoke-ProjectAPI 'POST' ($projectPath + '/status') @{status='active';version=$project.version}
    $estimateBody = @{base_revision=0;lines=@(@{kind='labor';name='Synthetic labor';quantity_milli=1500;unit_cents=101;vat_bps=2100},@{kind='material';name='Synthetic material';quantity_milli=1000;unit_cents=100;vat_bps=1000})}
    $preview = Invoke-ProjectAPI 'POST' ($projectPath + '/estimates/preview') $estimateBody
    Require-Result ($preview.total_cents -eq 294) 'integer estimate preview'
    $estimate = Invoke-ProjectAPI 'POST' ($projectPath + '/estimates') $estimateBody 201
    $null = Invoke-ProjectAPI 'POST' ($projectPath + '/estimates/' + $estimate.estimate_id + '/approve') @{}
    $estimateBody.base_revision = 1
    $estimateBody.lines[0].unit_cents = 200
    $null = Invoke-ProjectAPI 'POST' ($projectPath + '/estimates') $estimateBody 201
    $document = Invoke-ProjectAPI 'POST' ($projectPath + '/documents') @{type='estimate';name='Synthetic register entry';estimate_id=$estimate.estimate_id;annotation='Metadata only'} 201
    $authRoutes = @(
        @('GET','/projects'), @('POST','/projects'), @('GET',$projectPath), @('PUT',$projectPath),
        @('POST',($projectPath+'/status')), @('GET',($projectPath+'/estimates')), @('POST',($projectPath+'/estimates')),
        @('POST',($projectPath+'/estimates/preview')), @('POST',($projectPath+'/estimates/'+$estimate.estimate_id+'/approve')),
        @('GET',($projectPath+'/documents')), @('POST',($projectPath+'/documents')), @('GET',('/leads/'+$first.LeadID+'/project'))
    )
    # Split assertions across the two existing process lifecycles so each stays
    # below Huginn's ten-auth-failures/minute limit. Do not disable that limiter.
    Require-ProjectAuth @(@('GET','/backup'), @('GET','/licenses'))
    $license = Invoke-ProjectAPI 'GET' '/licenses'
    Require-Result ($license.upstream_commit -eq 'f01c1d84c08b1de972d8cc268da25602598adff8' -and $license.license.Length -gt 1000) 'embedded pinned license'
    $backupPath = Join-Path $root 'synthetic-crm-backup.sqlite'
    $backupResponse = Invoke-WebRequest -Uri ($base + '/api/v1/valentina/backup') -Headers $script:crmHeaders -OutFile $backupPath -PassThru -TimeoutSec 40
    Require-Result ((Get-FileHash $backupPath).Hash.ToLowerInvariant() -eq [string]$backupResponse.Headers['X-Backup-SHA256']) 'CRM backup bytes match SHA256'
    Require-Result ([string]$backupResponse.Headers['Cache-Control'] -eq 'no-store' -and [string]$backupResponse.Headers['X-Backup-Schema-Version'] -eq '14') 'CRM backup safety headers'
    Require-ProjectAuth $authRoutes[0..5]
    # The managed OS TCP inventory does not require CIM privileges. Fail closed
    # if the actual listening endpoint cannot be inspected or is not loopback.
    $listen = @([Net.NetworkInformation.IPGlobalProperties]::GetIPGlobalProperties().GetActiveTcpListeners() | Where-Object { $_.Port -eq $port })
    Require-Result ($listen.Count -eq 1 -and $listen[0].Address.ToString() -eq '127.0.0.1') 'loopback-only bind'
    Stop-Process -Id $serverA.Id -ErrorAction Stop
    $serverA = $null
    Start-Sleep -Milliseconds 800
    $serverB = Start-SmokeServer 2
    $afterRestart = Signed-Intake $body
    Require-Result ($afterRestart.Code -eq 200 -and $afterRestart.LeadID -eq $first.LeadID) 'idempotency persisted across restart'
    $localToken = Invoke-RestMethod -Uri ($base + '/api/v1/token') -TimeoutSec 5
    $script:crmHeaders = @{ Authorization = ('Bearer ' + $localToken.token) }
    $restored = Invoke-ProjectAPI 'GET' $projectPath
    $history = Invoke-ProjectAPI 'GET' ($projectPath + '/estimates')
    $registry = Invoke-ProjectAPI 'GET' ($projectPath + '/documents')
    Require-Result ($restored.status -eq 'active' -and $restored.source_lead_id -eq $first.LeadID) 'project persists across restart'
    Require-Result ($history.estimates.Count -eq 2 -and $history.estimates[1].total_cents -eq 294 -and $history.estimates[1].status -eq 'approved' -and $history.estimates[0].status -eq 'draft') 'immutable estimate history persists'
    Require-Result ($registry.storage -eq 'metadata_only' -and $registry.documents[0].document_id -eq $document.document_id) 'document metadata persists'
    Require-ProjectAuth $authRoutes[6..11]
    Write-Output 'VALENTINA_PROJECTS_SMOKE=PASS (human-qualified lead, unique project, cents/VAT, revisions, approval, registry, all 12 routes auth 401, restart)'
    Write-Output 'VALENTINA_RUNTIME_SMOKE=PASS (UI 200, intake 201/200/409/401, stale/future 401, CRM auth 401, restart 200, loopback)'
} finally {
    foreach ($p in @($serverA, $serverB)) {
        if ($p -and (Get-Process -Id $p.Id -ErrorAction SilentlyContinue)) {
            Stop-Process -Id $p.Id -ErrorAction SilentlyContinue
        }
    }
    Start-Sleep -Milliseconds 850
    if (@([Net.NetworkInformation.IPGlobalProperties]::GetIPGlobalProperties().GetActiveTcpListeners() | Where-Object { $_.Port -eq $port }).Count -gt 0) {
        Write-Warning 'Runtime smoke port still bound; check test child process'
    }
    # Only remove the unique disposable test profile that this script created.
    $resolvedTestRoot = [IO.Path]::GetFullPath($root)
    $tempPrefix = [IO.Path]::GetFullPath([IO.Path]::GetTempPath()).TrimEnd('\') + '\'
    if (-not $RetainTestProfile -and $resolvedTestRoot.StartsWith($tempPrefix, [StringComparison]::OrdinalIgnoreCase) -and (Split-Path $resolvedTestRoot -Leaf).StartsWith('valentina-runtime-smoke-') -and (Test-Path -LiteralPath $root)) {
        Remove-Item -LiteralPath $root -Recurse -Force -ErrorAction SilentlyContinue
    }
}
