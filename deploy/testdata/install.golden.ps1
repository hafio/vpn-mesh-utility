# meshctl installer for Windows, served by the meshctl control plane at
# /install.ps1. From an elevated PowerShell (Run as administrator):
#
#   $env:MESHCTL_TOKEN = '<token>'; $env:MESHCTL_MESH = '<mesh>'
#   irm https://<server>/install.ps1 | iex
#
# It downloads meshctl for Windows from the same server, checks the download
# against the SHA-256 baked into this script when the server rendered it,
# installs it to Program Files\meshctl, runs `meshctl agent join` and
# installs and starts the mesh's service, meshctl-agent-<mesh> (amendment
# A51), which Windows starts at boot and restarts when it fails. The agent
# keeps its data in %ProgramData%\meshctl\<mesh>. MESHCTL_NAME, when set, is
# the node name the token must belong to. The token leaves the session's
# environment once the join has used it.
#
# It only installs (amendment A55): where the mesh is installed already it
# says what is there and stops, before it downloads anything or uses the
# token. reinstall.ps1 reinstalls, joins again or starts afresh;
# uninstall.ps1 removes it. Another mesh on the same PC is another
# MESHCTL_MESH, which runs as its own service.
& {
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
$Self = 'install.ps1'

# The part install.ps1, reinstall.ps1 and uninstall.ps1 share (amendment
# A55): the control plane writes it into each of them when it serves them,
# so every script is one file a PC pipes into iex. It runs inside the
# script's own block, so no function, variable or preference of the script
# stays in the PowerShell session after it.

$Origin = 'https://mesh.example.com'
$Version = 'v1.2.3'
$Sha256 = 'cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc'

# What a Windows install keeps where (amendment A51): one data dir and one
# service per mesh, one binary (with wintun.dll beside it) for all of them.
$Root = Join-Path $env:ProgramData 'meshctl'
$Dir = Join-Path $env:ProgramFiles 'meshctl'
$Exe = Join-Path $Dir 'meshctl.exe'

function Fail([string]$message) { throw "${Self}: $message" }

# Assert-Mesh holds MESHCTL_MESH to a mesh name: it names the service and the
# data dir.
function Assert-Mesh([string]$mesh) {
    if ($mesh -cnotmatch '^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9])?$') {
        Fail "MESHCTL_MESH must be a mesh name (a-z, 0-9 and '-', at most 32), not '$mesh'"
    }
}

# Assert-Host stops unless the script runs elevated on 64-bit Windows.
function Assert-Host {
    $principal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
    if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
        Fail 'run it from an elevated PowerShell (Run as administrator)'
    }
    if ($env:PROCESSOR_ARCHITECTURE -ne 'AMD64') {
        Fail "unsupported CPU $($env:PROCESSOR_ARCHITECTURE); meshctl ships windows/amd64"
    }
    [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
}

# Test-Joined is true when the mesh has a joined node here: its data dir
# holds an enrollment.
function Test-Joined([string]$mesh) { Test-Path -LiteralPath (Join-Path (Join-Path $Root $mesh) 'state.json') }

# Get-Meshes lists the meshes with a joined node here.
function Get-Meshes {
    Get-ChildItem -Directory -LiteralPath $Root -ErrorAction SilentlyContinue |
        Where-Object { Test-Path -LiteralPath (Join-Path $_.FullName 'state.json') } | ForEach-Object { $_.Name }
}

# Show-Mesh says what of the mesh is on this PC and returns its node's name.
function Show-Mesh([string]$mesh) {
    $state = $null
    try { $state = Get-Content -Raw -LiteralPath (Join-Path (Join-Path $Root $mesh) 'state.json') | ConvertFrom-Json } catch { }
    $node = "$($state.name)"
    $ver = "no $Exe"
    if (Test-Path -LiteralPath $Exe) {
        try { $ver = "$(& $Exe version | Select-Object -First 1)" } catch { }
    }
    $status = 'not registered'
    $svc = Get-Service -Name "meshctl-agent-$mesh" -ErrorAction SilentlyContinue
    if ($svc) { $status = "$($svc.Status)" }
    $shown = $node
    if (-not $shown) { $shown = '?' }
    Write-Host "mesh $mesh is installed here: node $shown of $($state.server), data in $(Join-Path $Root $mesh), $ver, meshctl-agent-$mesh $status"
    $node
}

# Test-Interactive is true when there is someone to ask: not a service, not
# powershell -NonInteractive.
function Test-Interactive {
    [Environment]::UserInteractive -and -not ([Environment]::GetCommandLineArgs() | Where-Object { $_ -match '^-noni' })
}

# Read-Token reads a join token at the console, not echoed.
function Read-Token([string]$mesh) {
    if (-not (Test-Interactive)) { Fail 'no join token: set $env:MESHCTL_TOKEN first' }
    $secure = Read-Host -AsSecureString -Prompt "Join token for the node (meshctl node token <name> --mesh $mesh prints one)"
    $bstr = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($secure)
    try {
        $token = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($bstr)
    } finally {
        [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr)
    }
    if (-not $token) { Fail 'no join token given; nothing was changed' }
    $token
}

# Get-Download fetches meshctl for Windows from the control plane into a
# temporary file and checks it against the SHA-256 baked into this script,
# before anything on the PC changes; it returns the file.
function Get-Download {
    $tmp = Join-Path ([IO.Path]::GetTempPath()) ('meshctl-' + [Guid]::NewGuid().ToString('N') + '.exe')
    try {
        Invoke-WebRequest -UseBasicParsing -Uri "$Origin/dl/meshctl-windows-amd64.exe" -OutFile $tmp
    } catch {
        Remove-Item -Force -ErrorAction SilentlyContinue -LiteralPath $tmp
        Fail "downloading meshctl-windows-amd64.exe from $Origin failed ($($_.Exception.Message)); nothing was changed"
    }
    if ((Get-FileHash -Algorithm SHA256 -LiteralPath $tmp).Hash.ToLowerInvariant() -ne $Sha256) {
        Remove-Item -Force -ErrorAction SilentlyContinue -LiteralPath $tmp
        Fail 'checksum mismatch for meshctl-windows-amd64.exe: the download is corrupt or was altered; nothing was changed'
    }
    $tmp
}

# Install-Binary puts the checked download $tmp in Program Files. Every
# mesh's service runs that one binary, so the ones that run stop while it is
# replaced; it returns them, for the caller to start again.
function Install-Binary([string]$tmp) {
    if ((Test-Path -LiteralPath $Exe) -and (Get-FileHash -Algorithm SHA256 -LiteralPath $Exe).Hash.ToLowerInvariant() -eq $Sha256) {
        Write-Host "meshctl $Version is in $Exe already"
        return
    }
    $running = @(Get-Service -Name 'meshctl-agent-*' -ErrorAction SilentlyContinue | Where-Object { $_.Status -eq 'Running' })
    $running | Stop-Service -Force
    try {
        New-Item -ItemType Directory -Force -Path $Dir | Out-Null
        Copy-Item -Force -LiteralPath $tmp -Destination $Exe
    } catch {
        $running | Start-Service -ErrorAction Continue
        Fail "installing $Exe failed ($($_.Exception.Message)); the services this stopped run again"
    }
    Write-Host "installed meshctl $Version to $Exe"
    $running
}

# Join-Mesh joins the mesh with the token. The join needs the mesh's data dir
# to itself, so its service stops first; a failed join starts again what was
# stopped, here and by Install-Binary ($stopped), and says so with $gone
# added.
function Join-Mesh([string]$mesh, [string]$token, [string]$name, $stopped, [string]$gone) {
    $own = @(Get-Service -Name "meshctl-agent-$mesh" -ErrorAction SilentlyContinue | Where-Object { $_.Status -eq 'Running' })
    $own | Stop-Service -Force
    $join = @('agent', 'join', '--server', $Origin, '--mesh', $mesh)
    if ($name) { $join += @('--name', $name) }
    # The token reaches meshctl in its environment, never on its command
    # line, and leaves the session's environment once the join has used it.
    $env:MESHCTL_TOKEN = $token
    try {
        & $Exe @join
        $code = $LASTEXITCODE
    } finally {
        Remove-Item -ErrorAction SilentlyContinue -Path Env:MESHCTL_TOKEN
    }
    if ($code -ne 0) {
        (@($stopped) + $own) | Start-Service -ErrorAction Continue
        Fail "meshctl agent join failed (exit $code)$gone"
    }
}

# Register-Agent creates or updates the mesh's service, which Windows starts
# at boot and restarts when it fails, also when the agent exits with an
# error, and (re)starts it.
function Register-Agent([string]$mesh) {
    $service = "meshctl-agent-$mesh"
    $command = '"' + $Exe + '" agent run --mesh ' + $mesh
    $existing = Get-CimInstance -ClassName Win32_Service -Filter "Name='$service'"
    if ($existing) {
        $r = Invoke-CimMethod -InputObject $existing -MethodName Change -Arguments @{ PathName = $command; StartMode = 'Automatic' }
        if ($r.ReturnValue -ne 0) { Fail "updating the $service service failed (code $($r.ReturnValue))" }
    } else {
        New-Service -Name $service -BinaryPathName $command -DisplayName "meshctl agent ($mesh)" `
            -Description 'Keeps this machine in its meshctl WireGuard mesh.' -StartupType Automatic | Out-Null
    }
    & sc.exe failure $service reset= 86400 actions= restart/5000/restart/30000/restart/60000 | Out-Null
    & sc.exe failureflag $service 1 | Out-Null
    Restart-Service -Name $service
    Write-Host "started the $service service"
}

# Remove-Agent takes the mesh's agent off this PC: it stops the service, has
# `meshctl agent reset` take off what the agent made, then deletes the
# service (unless $keep) and the data dir. A data dir that holds other files
# than meshctl's, or a failed reset, stops it before anything is deleted.
function Remove-Agent([string]$mesh, [bool]$keep) {
    $service = "meshctl-agent-$mesh"
    $data = Join-Path $Root $mesh
    if (Test-Path -LiteralPath $data -PathType Leaf) { Fail "$data is not a directory: nothing was removed" }
    $isDir = Test-Path -LiteralPath $data -PathType Container
    $ours = (Test-Path -LiteralPath (Join-Path $data 'state.json')) -or (Test-Path -LiteralPath (Join-Path $data 'agent.lock'))
    if ($isDir -and -not $ours -and (Get-ChildItem -Force -LiteralPath $data | Select-Object -First 1)) {
        Fail "$data holds files but no meshctl data: nothing was removed"
    }
    Get-Service -Name $service -ErrorAction SilentlyContinue | Where-Object { $_.Status -ne 'Stopped' } | Stop-Service -Force
    if (Test-Joined $mesh) {
        if (-not (Test-Path -LiteralPath $Exe)) {
            Fail "$Exe is missing, so nothing can take the agent's tunnel, routes and DNS rules off this PC: run reinstall.ps1 with `$env:MESHCTL_ACTION = 'reinstall' first; nothing was deleted"
        }
        & $Exe agent reset --mesh $mesh
        if ($LASTEXITCODE -ne 0) {
            Fail 'meshctl agent reset could not take everything off this PC (above): fix that and run this again; nothing was deleted'
        }
    }
    if (-not $keep -and (Get-Service -Name $service -ErrorAction SilentlyContinue)) {
        & sc.exe delete $service | Out-Null
        if ($LASTEXITCODE -ne 0) { Fail "deleting the $service service failed (sc.exe returned $LASTEXITCODE)" }
    }
    if ($isDir) { Remove-Item -Recurse -Force -LiteralPath $data }
    Write-Host "removed the agent of $mesh and its data ($data)"
}

$token = $env:MESHCTL_TOKEN
$mesh = $env:MESHCTL_MESH
$named = [bool]$mesh
if (-not $named) { $mesh = 'default' }
if (-not $token) { Fail 'no join token: set $env:MESHCTL_TOKEN first' }
Assert-Mesh $mesh
Assert-Host

if (Test-Joined $mesh) {
    Show-Mesh $mesh | Out-Null
    if (-not $named) {
        Write-Host 'To add another mesh to this PC, set $env:MESHCTL_MESH to its name: each mesh runs as its own service, meshctl-agent-<mesh>.'
    }
    Fail ("mesh $mesh is installed here already, and install.ps1 only installs; nothing was changed and the token is unused. " +
        "To reinstall meshctl, join again or start afresh: `$env:MESHCTL_MESH = '$mesh'; irm $Origin/reinstall.ps1 | iex. " +
        "To remove it: `$env:MESHCTL_MESH = '$mesh'; irm $Origin/uninstall.ps1 | iex")
}

$tmp = Get-Download
try {
    $stopped = @(Install-Binary $tmp)
} finally {
    Remove-Item -Force -ErrorAction SilentlyContinue -LiteralPath $tmp
}
Join-Mesh $mesh $token $env:MESHCTL_NAME $stopped ''
Register-Agent $mesh
$stopped | Where-Object { $_.Name -ne "meshctl-agent-$mesh" } | Start-Service
}
