# Start local Postgres + SoftEther for nevarix-api.
# Pulls images only when you choose to run this script.
$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..\..")
docker compose up -d
docker compose ps
