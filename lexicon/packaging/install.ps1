[CmdletBinding()]
param(
    [string]$InstallDir = (Join-Path $env:LOCALAPPDATA "Programs\Lexicon"),
    [switch]$NoPath
)

$ErrorActionPreference = "Stop"

$sourceDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$sourceDir = [IO.Path]::GetFullPath($sourceDir).TrimEnd('\', '/')
$installDir = [IO.Path]::GetFullPath($InstallDir).TrimEnd('\', '/')

if (-not (Test-Path -LiteralPath (Join-Path $sourceDir "lexicon.exe") -PathType Leaf)) {
    throw "lexicon.exe was not found beside install.ps1. Run this script from an extracted Lexicon release package."
}
if (-not (Test-Path -LiteralPath (Join-Path $sourceDir "adapters") -PathType Container)) {
    throw "The adapters directory was not found beside install.ps1. The release package is incomplete."
}

$separator = [IO.Path]::DirectorySeparatorChar
if ($installDir.StartsWith($sourceDir + $separator, [StringComparison]::OrdinalIgnoreCase)) {
    throw "InstallDir cannot be inside the extracted release package."
}

if (-not [StringComparer]::OrdinalIgnoreCase.Equals($sourceDir, $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    Get-ChildItem -LiteralPath $sourceDir -Force | Copy-Item -Destination $installDir -Recurse -Force
}

if (-not $NoPath) {
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $segments = @($userPath -split ';' | Where-Object { $_ } | ForEach-Object { $_.TrimEnd('\', '/') })
    if (-not ($segments | Where-Object { [StringComparer]::OrdinalIgnoreCase.Equals($_, $installDir) })) {
        $newPath = if ([string]::IsNullOrWhiteSpace($userPath)) { $installDir } else { "$userPath;$installDir" }
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    }

    $processSegments = @($env:Path -split ';' | Where-Object { $_ } | ForEach-Object { $_.TrimEnd('\', '/') })
    if (-not ($processSegments | Where-Object { [StringComparer]::OrdinalIgnoreCase.Equals($_, $installDir) })) {
        $env:Path = "$installDir;$env:Path"
    }
}

Write-Host "Lexicon installed to $installDir"
if ($NoPath) {
    Write-Host "PATH was not changed. Run $installDir\lexicon.exe directly or add that directory to PATH."
} else {
    Write-Host "The user PATH contains the installation directory. Open a new terminal before running lexicon."
}
