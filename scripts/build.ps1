param(
  [string]$OutputDir = "dist"
)

if (!(Test-Path -Path $OutputDir)) {
  New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

Push-Location (Join-Path $PSScriptRoot "..")
try {
  $env:CGO_ENABLED = "0"
  go build -o (Join-Path $OutputDir "lmp-server.exe") ./cmd/server
  Copy-Item -Recurse -Force (Join-Path $PSScriptRoot "..\..\frontend\public") (Join-Path $OutputDir "ui")
  $cfg = @{
    server = @{ port = 8080; lan_only = $true }
    auth = @{ enable_device_auth = $true; session_timeout = "30m" }
    files = @{ share_dirs = @() }
    rate_limit = @{ enabled = $true; requests_per_second = 10.0; burst = 20 }
  } | ConvertTo-Json -Depth 4
  $cfg | Set-Content -Path (Join-Path $OutputDir "config.json") -Encoding UTF8
} finally {
  Pop-Location
}

