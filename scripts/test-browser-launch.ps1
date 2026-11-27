#requires -Version 7.0
# Actual installed shortcut + default-browser dispatch. Disposable hosted CI only.
[CmdletBinding()]
param([Parameter(Mandatory=$true)][string]$ShortcutPath)
$ErrorActionPreference = 'Stop'
if ($env:GITHUB_ACTIONS -ne 'true' -or $env:RUNNER_ENVIRONMENT -ne 'github-hosted' -or $env:RUNNER_OS -ne 'Windows') { throw 'Browser launch acceptance requires disposable hosted Windows' }
function Require($ok,$message) { if (-not $ok) { throw "Browser launch acceptance: $message" } }
$link = (New-Object -ComObject WScript.Shell).CreateShortcut((Resolve-Path $ShortcutPath).Path)
Require ($link.Arguments -eq 'serve --open-browser') 'Unexpected installed shortcut'
$exe = $link.TargetPath
$edge = Join-Path ${env:ProgramFiles(x86)} 'Microsoft/Edge/Application/msedge.exe'
Require (Test-Path $edge) 'Installed Edge required'
function Free-Port {
    $listener = [Net.Sockets.TcpListener]::new([Net.IPAddress]::Loopback,0)
    $listener.Start(); $value=$listener.LocalEndpoint.Port; $listener.Stop(); return $value
}
$port=Free-Port; $debugPort=Free-Port
$root=Join-Path $env:RUNNER_TEMP ('valentina-launch-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path "$root/.huginn","$root/local","$root/roaming" -Force | Out-Null
$config=@{version=13;web_ui=@{enabled=$true;port=$port;auto_open=$false;bind='127.0.0.1'};scheduler_enabled=$false;tools_enabled=$false;mcp_servers=@()} | ConvertTo-Json -Depth 4
[IO.File]::WriteAllText("$root/.huginn/config.json",$config)
# Only this child script's environment changes; never use the runner's app profile.
$env:USERPROFILE=$root; $env:HUGINN_HOME=$root
$env:APPDATA="$root/roaming"; $env:LOCALAPPDATA="$root/local"
$env:NVIDIA_NIM_API_KEY=''; $env:VALENTINA_INTAKE_HMAC_SECRET=''
$choice=Get-ItemProperty 'HKCU:/Software/Microsoft/Windows/Shell/Associations/UrlAssociations/http/UserChoice' -ErrorAction SilentlyContinue
$progId=if ($choice -and $choice.ProgId) { [string]$choice.ProgId } else { 'http' }
Require ($progId -match '^[A-Za-z0-9._-]+$') 'Unexpected HTTP association'
$commandKey="HKCU:/Software/Classes/$progId/shell/open/command"
$existing=Test-Path $commandKey
$previous=if ($existing) { (Get-Item $commandKey).GetValue('') } else { $null }
$daemonId=$null
try {
    New-Item -Path $commandKey -Force | Out-Null
    # Exercise the real default-browser OS dispatch with a headless Edge profile.
    Set-Item -LiteralPath $commandKey -Value ('"'+$edge+'" --headless --no-first-run --no-default-browser-check --remote-debugging-port='+$debugPort+' --user-data-dir="'+$root+'/browser" "%1"')
    $base="http://127.0.0.1:$port"
    for ($cycle=1;$cycle -le 2;$cycle++) {
        Start-Process -FilePath $ShortcutPath -WindowStyle Hidden
        $observed=$false
        for ($attempt=0;$attempt -lt 60;$attempt++) {
            Start-Sleep -Milliseconds 500
            try {
                $pages=@(Invoke-RestMethod "http://127.0.0.1:$debugPort/json/list" -TimeoutSec 1 | Where-Object { $_.type -eq 'page' -and $_.url.StartsWith($base) })
                if ($pages.Count -ge $cycle) { $observed=$true; break }
            } catch { }
        }
        Require $observed 'Shortcut did not open a ready local browser tab'
        $currentId=[int](Get-Content "$root/.huginn/serve.pid" -Raw).Trim()
        if ($cycle -eq 1) { $daemonId=$currentId } else { Require ($currentId -eq $daemonId) 'Second launch replaced the server' }
        $process=Get-Process -Id $daemonId
        Require ($process.Path -eq $exe) 'Unexpected server process'
        $listeners=@([Net.NetworkInformation.IPGlobalProperties]::GetIPGlobalProperties().GetActiveTcpListeners() | Where-Object Port -eq $port)
        Require ($listeners.Count -eq 1 -and $listeners[0].Address.ToString() -eq '127.0.0.1') 'Non-loopback or duplicate listener'
        $token=(Invoke-RestMethod "$base/api/v1/token").token
        $headers=@{Authorization="Bearer $token"}
        if ($cycle -eq 1) {
            $body=@{request_summary='Synthetic shortcut persistence';inbound_channel='manual';privacy_notice=$true} | ConvertTo-Json
            $lead=Invoke-RestMethod "$base/api/v1/valentina/leads" -Method Post -Headers $headers -ContentType 'application/json' -Body $body
        } else {
            $retained=Invoke-RestMethod "$base/api/v1/valentina/leads/$($lead.lead_id)" -Headers $headers
            Require ($null -ne $retained) 'Synthetic CRM record lost'
        }
    }
    Write-Output 'BROWSER_LAUNCH_ACCEPTANCE=PASS (installed shortcut twice, real Edge tabs, same daemon, loopback, retained CRM)'
} finally {
    if ($existing) { Set-Item -LiteralPath $commandKey -Value $previous }
    else { Remove-Item -LiteralPath $commandKey -ErrorAction SilentlyContinue }
    # Stop only this isolated test profile's processes, never an arbitrary server.
    if (-not $daemonId -and (Test-Path "$root/.huginn/serve.pid")) { $daemonId=[int](Get-Content "$root/.huginn/serve.pid" -Raw).Trim() }
    if ($daemonId) {
        $owned=Get-Process -Id $daemonId -ErrorAction SilentlyContinue
        if ($owned -and $owned.Path -eq $exe) { Stop-Process -Id $daemonId }
    }
    if (Test-Path "$root/.huginn/tray.pid") {
        $trayId=[int](Get-Content "$root/.huginn/tray.pid" -Raw).Trim()
        $tray=Get-Process -Id $trayId -ErrorAction SilentlyContinue
        if ($tray -and $tray.Path -eq $exe) { Stop-Process -Id $trayId }
    }
    foreach ($process in Get-CimInstance Win32_Process | Where-Object { $_.CommandLine -and $_.CommandLine.Contains($root) -and $_.Name -eq 'msedge.exe' }) { Stop-Process -Id $process.ProcessId -ErrorAction SilentlyContinue }
}
