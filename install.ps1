param(
    [string]$Version = "latest",
    [string]$InstallDir = "$env:USERPROFILE\bin"
)

$ErrorActionPreference = "Stop"
$Repo = "JeremyXIonia/wildread-cli"
$Asset = "wildread-cli-windows-amd64.zip"

if ($Version -eq "latest") {
    $Url = "https://github.com/$Repo/releases/latest/download/$Asset"
} else {
    $Url = "https://github.com/$Repo/releases/download/$Version/$Asset"
}

$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $TempDir | Out-Null
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null

try {
    $ZipPath = Join-Path $TempDir $Asset
    Write-Host "Downloading $Url"
    Invoke-WebRequest -Uri $Url -OutFile $ZipPath -UseBasicParsing

    Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force
    $ExePath = Join-Path $TempDir "wildread-cli.exe"
    Copy-Item -Path $ExePath -Destination (Join-Path $InstallDir "wildread-cli.exe") -Force

    Write-Host "Installed wildread-cli.exe to $InstallDir"
    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if (($UserPath -split ';') -contains $InstallDir) {
        Write-Host "You can now run: wildread-cli"
    } else {
        Write-Host ""
        Write-Host "To run wildread-cli from anywhere, add $InstallDir to your User PATH."
        Write-Host "For modern PowerShell, run:"
        Write-Host ""
        Write-Host "  `$installDir = '$InstallDir'"
        Write-Host '  $userPath = [Environment]::GetEnvironmentVariable(''Path'', ''User'')'
        Write-Host '  [Environment]::SetEnvironmentVariable(''Path'', "$userPath;$installDir", ''User'')'
        Write-Host ""
        Write-Host "Then reopen PowerShell."
    }
} finally {
    Remove-Item -Recurse -Force $TempDir -ErrorAction SilentlyContinue
}
