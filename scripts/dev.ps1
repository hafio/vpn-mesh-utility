<#
scripts/dev.ps1 - build/vet/test/cov/scan/image tasks for meshctl.

Mirror of scripts/dev.sh (behaviourally identical: same task names, same
gating, same footer format). The USER runs this; CI calls task names only.
Works from any cwd. Accepts multiple tasks per invocation and stops at the
first failure.

  ./scripts/dev.ps1 build vet test
  ./scripts/dev.ps1 all          # build vet test                 (CI runs: all scan)
  ./scripts/dev.ps1 full         # all + cov image scan graphify  (pre-tag sweep)

Windows PowerShell 5.1 compatible: ASCII only, no &&, no ternary, and no
Tee-Object on native commands.
#>
#requires -Version 5.1
[CmdletBinding()]
param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Tasks)

$ErrorActionPreference = 'Continue'

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
# Captured at script scope: inside a function $MyInvocation names the function.
$ScriptName = Split-Path -Leaf $MyInvocation.MyCommand.Path
$RepoRoot  = Split-Path -Parent $ScriptDir
$LogDir    = Join-Path $ScriptDir 'logs'
$DistDir   = Join-Path $RepoRoot  'dist'
$CovDir    = Join-Path $RepoRoot  'coverage'
# The binary is always meshctl, never the directory name: the repo is called
# vpn-mesh-utility, and the release merges every leg by file name.
$BinName   = 'meshctl'
$ImageRepo = if ($env:IMAGE_REPO) { $env:IMAGE_REPO } else { 'vpn-mesh-utility' }

# Version stamp. The release workflow exports VERSION from the tag ref; locally
# git describe is used, then "dev". 2>$null, not 2>&1: on PS 5.1 piping a native
# command's stderr wraps each line as a NativeCommandError. The value reaches
# -ldflags and an image tag, so it is held to a strict charset.
$Version = $env:VERSION
if (-not $Version) {
  if (Get-Command git -ErrorAction SilentlyContinue) {
    $Version = (& git -C $RepoRoot describe --tags --always --dirty 2>$null)
    if ($LASTEXITCODE -ne 0) { $Version = $null }
  }
  if (-not $Version) { $Version = 'dev' }
}
if ($Version -notmatch '^[A-Za-z0-9][A-Za-z0-9._+-]{0,63}$') {
  Write-Host "error: VERSION '$Version' has characters outside [A-Za-z0-9._+-]; export a plain tag such as v1.2.3" -ForegroundColor Red
  exit 1
}
# Images drop the leading v (registry convention); '+' is not legal in a tag.
$ImageTag = ($Version -replace '^v', '') -replace '\+', '-'

# Local convenience cross-compile set (`dist`). CI does not use it: tag.yml
# matrixes over `build` with TARGET_OS/TARGET_ARCH from BUILD_TARGETS instead.
$DistTargets = @(
  @{ os = 'linux';   arch = 'amd64' },
  @{ os = 'linux';   arch = 'arm64' },
  @{ os = 'windows'; arch = 'amd64' }
)

# Packages whose tests own golden files (`regen` rewrites them with -update).
$RegenPkgs = @('./internal/server/planner', './internal/server/phone', './internal/server/ui', './internal/agent/firewall', './internal/agent/wg', './internal/agent/localui', './deploy', './cmd/meshctl')

# -race needs cgo + a C compiler; OFF by default here, enable with
# MESHCTL_RACE=1 (dev.sh defaults it on). `cov` never races.
$RaceFlag = @()
if ($env:MESHCTL_RACE -eq '1') { $RaceFlag = @('-race') }
if ($RaceFlag.Count -gt 0 -and -not $env:GORACE) { $env:GORACE = 'atexit_sleep_ms=0' }

# Toolchain parity: go.mod's `toolchain` pin is what local and CI must agree
# on. Apply it only when GOTOOLCHAIN is unset, so an explicit value wins.
$goMod = Join-Path $RepoRoot 'go.mod'
if (-not $env:GOTOOLCHAIN -and (Test-Path $goMod)) {
  # @() forces an array: -match on a single hit returns a bare string.
  $tcLines = @((Get-Content $goMod) -match '^toolchain\s+\S+')
  if ($tcLines.Count -gt 0) { $env:GOTOOLCHAIN = ($tcLines[0] -split '\s+')[1] }
}

# Keep captured logs clean: no color / progress from tools.
$env:NO_COLOR = '1'

# The previous coverage total is the floor (S5). Read it before any task runs,
# because `cov` truncates its own log when it starts.
$PrevCov = ''
$covLog = Join-Path $LogDir 'cov.log'
if (Test-Path $covLog) {
  $hits = @(Get-Content $covLog | Where-Object { $_ -match '^total:' })
  if ($hits.Count -gt 0) { $PrevCov = ($hits[-1] -split '\s+')[-1] }
}

# --- pretty helpers -----------------------------------------------------------
function Step($m) { Write-Host "==> $m"    -ForegroundColor Cyan }
function Ok($m)   { Write-Host "[ ok ] $m" -ForegroundColor Green }
function Warn($m) { Write-Host "[warn] $m" -ForegroundColor Yellow }
function Die($m)  { Write-Host "[fail] $m" -ForegroundColor Red; exit 1 }

# Matches dev.sh's `date +%Y-%m-%dT%H:%M:%S%z`: .NET 'zzz' renders the offset as
# +08:00, so strip the colon or the two scripts' footers would not match.
function Get-Now {
  $d = Get-Date
  return $d.ToString('yyyy-MM-ddTHH:mm:ss') + $d.ToString('zzz').Replace(':', '')
}
function Get-Log($task) { Join-Path $LogDir "$task.log" }

$script:LogFile = $null
function Start-TaskLog($task) {
  if (-not (Test-Path $LogDir)) { New-Item -ItemType Directory -Path $LogDir | Out-Null }
  $script:LogFile = Get-Log $task
  # Replaced, not truncated: a log another account left behind is not ours to write.
  Remove-Item -LiteralPath $script:LogFile -Force -ErrorAction SilentlyContinue
  "=== $(Get-Now) | $task ===" | Set-Content -Path $script:LogFile -Encoding utf8
}

# Note writes a plain line to the console and the task log.
function Note($m) {
  Add-Content -Path $script:LogFile -Value $m -Encoding utf8
  Write-Host $m
}

# Skip: a task that does not run says why in its log too, so an OK footer
# never stands for a gate that did not run without a reason beside it.
function Skip($m) {
  Warn $m
  Add-Content -Path $script:LogFile -Value "[warn] $m" -Encoding utf8
}

# Contract footer, appended to the task log and echoed. Add-Content, never
# Tee-Object: Tee doubles lines and writes UTF-16 into a UTF-8 file.
function Write-Finish {
  param([string]$Task, [int]$Code, [int]$Seconds)
  $status = 'OK'
  if ($Code -ne 0) { $status = "FAILED (exit $Code)" }
  $line = '{0} | {1} | {2}s | {3}' -f (Get-Now), $Task, $Seconds, $status
  Add-Content -Path (Get-Log $Task) -Value $line -Encoding utf8
  Write-Host $line
}

# Cap runs a native command, capturing combined output once (never pipe a native
# command through Tee-Object on 5.1). Appends an ANSI-stripped copy to the log,
# echoes it, and returns the command's exit code. No param block on purpose:
# every token lands in $args verbatim instead of being parsed as a parameter.
function Cap {
  $ErrorActionPreference = 'Continue'
  $exe  = $args[0]
  $rest = @(); if ($args.Count -gt 1) { $rest = $args[1..($args.Count - 1)] }
  $out  = (& $exe @rest 2>&1 | ForEach-Object { "$_" } | Out-String -Width 4096)
  $code = $LASTEXITCODE
  $clean = $out -replace "\x1b\[[0-9;?]*[a-zA-Z]", ""
  Add-Content -Path $script:LogFile -Value $clean -Encoding utf8
  if ($clean.Trim().Length -gt 0) { Write-Host $clean.TrimEnd() }
  return $code
}

# With-Env runs a scriptblock with environment overrides, restoring them after.
function With-Env {
  param([hashtable]$Vars, [scriptblock]$Body)
  $saved = @{}
  foreach ($k in $Vars.Keys) {
    $saved[$k] = [Environment]::GetEnvironmentVariable($k)
    [Environment]::SetEnvironmentVariable($k, $Vars[$k])
  }
  try { return (& $Body) }
  finally {
    foreach ($k in $saved.Keys) { [Environment]::SetEnvironmentVariable($k, $saved[$k]) }
  }
}

function Get-HostOs   { return (& go env GOHOSTOS) }
function Get-HostArch { return (& go env GOHOSTARCH) }

# Docker can only build this project's Linux image when the daemon runs Linux
# containers. GitHub's windows runners are Windows-container only, so the image
# half of `image`/`scan` warn-skips there. Same rule in both scripts.
$script:DockerWhy = ''
function Test-LinuxDocker {
  if (-not (Get-Command docker -ErrorAction SilentlyContinue)) { $script:DockerWhy = 'docker not on PATH'; return $false }
  $t = ((& docker info --format '{{.OSType}}' 2>&1) | ForEach-Object { "$_" }) -join ' '
  if ("$t".Trim() -eq 'linux') { return $true }
  $script:DockerWhy = "docker info: $t"
  return $false
}

function Build-Image {
  return (Cap docker build --progress=plain --build-arg "VERSION=$Version" -t "${ImageRepo}:$ImageTag" $RepoRoot)
}

# trivy natively when it is on PATH, otherwise the container. --pull=always
# matters more than the tag: without it docker reuses whatever :latest resolved
# to months ago -- unpinned AND stale.
function Invoke-TrivyImage($Ref) {
  if (Get-Command trivy -ErrorAction SilentlyContinue) {
    return (Cap trivy image --quiet --exit-code 1 --severity HIGH,CRITICAL --ignore-unfixed $Ref)
  }
  return (Cap docker run --rm --pull=always -e NO_COLOR=1 -v /var/run/docker.sock:/var/run/docker.sock aquasec/trivy:latest image --quiet --exit-code 1 --severity HIGH,CRITICAL --ignore-unfixed $Ref)
}

# --- tasks --------------------------------------------------------------------
# Tasks return non-zero on failure; the dispatcher writes the footer and stops.
# PowerShell returns EVERY uncaptured pipeline value, so route native commands
# through Cap and pipe anything not being returned to Out-Null.

# Build-One compiles meshctl for one target into dist/. The target is in the
# file name because the release job merges every leg with merge-multiple.
function Build-One($os, $arch) {
  $out = Join-Path $DistDir "$BinName-$os-$arch"
  if ($os -eq 'windows') { $out = "$out.exe" }
  if (-not (Test-Path $DistDir)) { New-Item -ItemType Directory -Path $DistDir | Out-Null }
  Step "  $os/$arch -> $out"
  return (With-Env @{ CGO_ENABLED = '0'; GOOS = $os; GOARCH = $arch } {
    Cap go build -trimpath -ldflags "-s -w -X main.version=$Version" -o $out ./cmd/meshctl
  })
}

# CI sets TARGET_OS/TARGET_ARCH (from the BUILD_TARGETS repo variable); unset
# means build for the host, so local and CI output are the same shape.
function Task-build {
  $os = $env:TARGET_OS;     if (-not $os)   { $os = Get-HostOs }
  $arch = $env:TARGET_ARCH; if (-not $arch) { $arch = Get-HostArch }
  return (Build-One $os $arch)
}

function Task-dist {
  foreach ($t in $DistTargets) {
    $c = Build-One $t.os $t.arch
    if ($c -ne 0) { return $c }
  }
  Ok "binaries in $DistDir"
  return 0
}

# vet: go vet for the host AND the other shipped OS, staticcheck for the host
# (each CI leg covers its own OS), then gofmt. Every check runs and reports.
function Task-vet {
  $rc = 0
  $other = 'windows'
  if ((Get-HostOs) -eq 'windows') { $other = 'linux' }
  if ((Cap go vet ./...) -ne 0) { $rc = 1 }
  $c = With-Env @{ GOOS = $other } { Cap go vet ./... }
  if ($c -ne 0) { $rc = 1 }
  if ((Cap go tool staticcheck ./...) -ne 0) { $rc = 1 }
  $dirs = @('cmd', 'internal', 'deploy', 'ui', 'test') | ForEach-Object { Join-Path $RepoRoot $_ } | Where-Object { Test-Path $_ }
  $unformatted = @(& gofmt -l @dirs 2>$null)
  if ($unformatted.Count -gt 0) {
    Note 'gofmt: these files need formatting (run gofmt -w on them):'
    Note ($unformatted -join "`n")
    $rc = 1
  }
  return $rc
}

# -timeout bounds each package's test binary: a deadlocked test fails in
# minutes instead of hanging for go test's default 10.
$TestTimeout = '3m'
function Task-test { return (Cap go test @RaceFlag -count=1 "-timeout=$TestTimeout" ./...) }

function Task-cov {
  if (-not (Test-Path $CovDir)) { New-Item -ItemType Directory -Path $CovDir | Out-Null }
  $prof = Join-Path $CovDir 'coverage.out'
  $html = Join-Path $CovDir 'coverage.html'
  # -count=1 forces a real run so a cached result cannot report a stale total.
  $c = Cap go test '-covermode=atomic' "-coverprofile=$prof" -count=1 "-timeout=$TestTimeout" ./...
  if ($c -ne 0) { return $c }
  $c = Cap go tool cover "-html=$prof" '-o' $html
  if ($c -ne 0) { return $c }
  $total = (& go tool cover "-func=$prof" | Select-Object -Last 1)
  Note $total
  $nowPct = ($total -split '\s+')[-1]
  if (-not $PrevCov) {
    Note 'coverage floor: no previous total in logs/cov.log; this run is the baseline'
  } elseif ([double]($nowPct.TrimEnd('%')) -lt [double]($PrevCov.TrimEnd('%'))) {
    Warn "coverage DROPPED: $PrevCov -> $nowPct; add tests or state in the change why it moved (S5)"
    Note "coverage floor: previous $PrevCov, now $nowPct (DROPPED)"
  } else {
    Note "coverage floor: previous $PrevCov, now $nowPct"
  }
  Ok "coverage -> $html  ($nowPct)"
  return 0
}

# scan: every applicable check, FATAL on a fixable CVE. See dev.sh for the why:
# govulncheck via the go.mod tool directive in JSON mode, judged by
# internal/tools/vulnjudge, then trivy over a freshly built image.
function Task-scan {
  $raw = Join-Path $LogDir 'scan.json'
  # 5.1 decodes native output with the console codepage; read as UTF-8, then
  # put the console back as it was.
  $prevEnc = [Console]::OutputEncoding
  try {
    [Console]::OutputEncoding = New-Object System.Text.UTF8Encoding($false)
    # `-format json` always exits 0, even with findings, so non-zero here is a
    # real tool failure and stays fatal.
    $json = (& go tool govulncheck -format json ./...) -join "`n"
    $code = $LASTEXITCODE
  } finally {
    [Console]::OutputEncoding = $prevEnc
  }
  if ($code -ne 0) {
    Warn 'govulncheck failed to run; see the output above'
    return $code
  }
  # BOM-less: 5.1's utf8 writers prepend a BOM and json.Decoder rejects it.
  [System.IO.File]::WriteAllText($raw, $json, (New-Object System.Text.UTF8Encoding($false)))
  $c = Cap go run ./internal/tools/vulnjudge $raw
  if ($c -ne 0) { return $c }
  if (-not (Test-Path (Join-Path $RepoRoot 'Dockerfile'))) {
    Note 'no Dockerfile; skipping the image scan'
    return 0
  }
  if (-not (Test-LinuxDocker)) {
    Skip "no linux-container docker daemon ($script:DockerWhy); skipping the image scan"
    return 0
  }
  $c = Build-Image; if ($c -ne 0) { return $c }
  return (Invoke-TrivyImage "${ImageRepo}:$ImageTag")
}

function Task-image {
  if (-not (Test-Path (Join-Path $RepoRoot 'Dockerfile'))) { Skip 'no Dockerfile; skipping image'; return 0 }
  if (-not (Test-LinuxDocker)) { Skip "no linux-container docker daemon ($script:DockerWhy); skipping image"; return 0 }
  return (Build-Image)
}

# up/down: the local single-node dev stack (server only, bridge network).
$ComposeDev = Join-Path $RepoRoot 'deploy/compose.dev.yml'
function Task-up {
  if (-not (Test-Path $ComposeDev)) { Skip "no $ComposeDev; nothing to bring up"; return 0 }
  return (With-Env @{ IMAGE = "${ImageRepo}:$ImageTag" } { Cap docker compose -f $ComposeDev up -d })
}
function Task-down {
  if (-not (Test-Path $ComposeDev)) { Skip "no $ComposeDev; nothing to bring down"; return 0 }
  return (With-Env @{ IMAGE = "${ImageRepo}:$ImageTag" } { Cap docker compose -f $ComposeDev down })
}

# regen rewrites the committed goldens from the code that generates them.
# Deliberately NOT in all/full: a gate that rewrote what it compares against
# could never fail. Review the diff before committing it.
function Task-regen {
  $rc = 0
  foreach ($pkg in $RegenPkgs) {
    $code = Cap go test $pkg -count=1 -update
    if ($code -ne 0 -and $rc -eq 0) { $rc = $code }
  }
  return $rc
}

# e2e: network-namespace tests need Linux and root; dev.sh runs them.
function Task-e2e {
  Skip 'e2e runs on Linux only; skipping'
  return 0
}

# Local only: the graph is a developer artifact, not a CI output.
function Task-graphify {
  if ($env:CI) { Skip 'graphify is local-only; skipping in CI'; return 0 }
  if (-not (Get-Command graphify -ErrorAction SilentlyContinue)) {
    Skip 'graphify not on PATH; skipping'; return 0
  }
  return (Cap graphify update .)
}

# --- dispatch -----------------------------------------------------------------
$All  = @('build', 'vet', 'test')
$Full = @('build', 'vet', 'test', 'cov', 'image', 'scan', 'graphify')

function Show-Usage {
  $raceDesc = ($RaceFlag -join ' ')
  $targetsDesc = (($DistTargets | ForEach-Object { "$($_.os)/$($_.arch)" }) -join ' ')
  Write-Host @"
$ScriptName - build/test/scan tooling for meshctl

Usage: $ScriptName <task> [task...]

Tasks:
  build    compile -> dist\$BinName-<os>-<arch>[.exe]; TARGET_OS/TARGET_ARCH pick
           the target, unset means host; stamps ``meshctl version`` from VERSION,
           else git describe, else "dev"
  vet      go vet (host + cross OS), staticcheck (host), gofmt -l
  test     go test $raceDesc -count=1 ./...
  cov      coverage profile -> coverage/coverage.html + printed total, compared
           with the previous total in logs/cov.log (the floor)
  scan     govulncheck (fatal on a fixable called vulnerability; no released
           fix warns) + trivy over a freshly built image
  image    docker build -> ${ImageRepo}:<version>
  up/down  local dev stack (deploy/compose.dev.yml)
  dist     cross-compile $targetsDesc
  regen    rewrite goldens in $($RegenPkgs -join ' ') (not in all/full)
  e2e      root-only netns tests (Linux only; skipped here)
  graphify refresh graphify-out/ (local only; skipped when CI is set)
  all      $($All -join ' ')   (what CI runs, as: all scan)
  full     $($Full -join ' ')   (pre-tag sweep)

Env: VERSION overrides the stamp; MESHCTL_RACE=1 enables -race on ``test``;
     GORACE defaults to atexit_sleep_ms=0 on race runs; GOTOOLCHAIN defaults to
     go.mod's ``toolchain`` pin; IMAGE_REPO overrides the local image name.
Logs: $LogDir\<task>.log (each run closes with a timestamped footer)
"@
}

Set-Location $RepoRoot
if (-not $Tasks -or $Tasks.Count -eq 0) { Show-Usage; exit 0 }
if ($Tasks[0] -in @('-h', '--help', 'help')) { Show-Usage; exit 0 }

$queue = @()
foreach ($t in $Tasks) {
  switch ($t) {
    'all'   { $queue += $All }
    'full'  { $queue += $Full }
    default { $queue += $t }
  }
}

foreach ($task in $queue) {
  if (-not (Get-Command "Task-$task" -ErrorAction SilentlyContinue)) {
    Die "unknown task: $task (try: $ScriptName help)"
  }
  Step $task
  Start-TaskLog $task
  $sw = [Diagnostics.Stopwatch]::StartNew()
  $code = 0
  try { $code = & "Task-$task" } catch { Write-Host $_; $code = 1 }
  # A stray pipeline value inside a task would make this an array; take the
  # command's own code, which is always the last value returned.
  if ($code -is [array]) { $code = $code[-1] }
  if ($null -eq $code) { $code = 0 }
  $sw.Stop()
  Write-Finish -Task $task -Code ([int]$code) -Seconds ([int]$sw.Elapsed.TotalSeconds)
  if ([int]$code -ne 0) { Warn "$task failed; stopping"; exit 1 }
  Ok $task
}
Ok "done: $($queue -join ' ')"
