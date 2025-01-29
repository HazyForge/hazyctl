$ErrorActionPreference = "Stop"

$REPO = "hazyforge/hazyctl"
$BINARY_NAME = "hazyctl.exe"
$INSTALL_DIR = "$HOME\.local\bin"

# Determine OS and architecture
$OS = (Get-CimInstance -ClassName Win32_OperatingSystem).Caption
$ARCH = (Get-CimInstance -ClassName Win32_Processor).Architecture

# Normalize architecture names
switch ($ARCH) {
    9 { $ARCH = "amd64" }
    12 { $ARCH = "arm64" }
    default { Write-Error "Unsupported architecture: $ARCH"; exit 1 }
}

# Map OS names to GitHub release names
switch -Regex ($OS) {
    "Windows" { $OS = "windows" }
    default { Write-Error "Unsupported OS: $OS"; exit 1 }
}

# Get latest release tag
$LATEST_TAG = (Invoke-RestMethod -Uri "https://api.github.com/repos/$REPO/releases/latest").tag_name
$VERSION = $LATEST_TAG.TrimStart("v")
$ARCHIVE_NAME = "hazyctl_${VERSION}_${OS}_${ARCH}"
$DOWNLOAD_URL = "https://github.com/$REPO/releases/download/$LATEST_TAG/$ARCHIVE_NAME"

$ARCHIVE_URL = "$DOWNLOAD_URL.zip"
$SHA256_URL = "$ARCHIVE_URL.sha256"

Write-Output "Downloading hazyctl $LATEST_TAG for $OS/$ARCH..."

# Download the archive
$TMP_ARCHIVE = [System.IO.Path]::GetTempFileName() + ".zip"
Invoke-WebRequest -Uri $ARCHIVE_URL -OutFile $TMP_ARCHIVE

# Verify the SHA256 checksum
Write-Output "Verifying checksum..."
$TMP_SHA256 = New-TemporaryFile
Invoke-WebRequest -Uri $SHA256_URL -OutFile $TMP_SHA256
$EXPECTED_CHECKSUM = Get-Content $TMP_SHA256 | ForEach-Object { $_.Split(" ")[0] }
$ACTUAL_CHECKSUM = Get-FileHash $TMP_ARCHIVE -Algorithm SHA256 | Select-Object -ExpandProperty Hash
if ($EXPECTED_CHECKSUM -ne $ACTUAL_CHECKSUM) {
    Write-Error "Checksum verification failed!"
    exit 1
}
Write-Output "Checksum verification passed."

# Extract the archive
$TMP_DIR = New-TemporaryFile
Remove-Item $TMP_DIR
New-Item -ItemType Directory -Path $TMP_DIR

Write-Output "Extracting archive $TMP_ARCHIVE to $TMP_DIR..."
Expand-Archive -Path $TMP_ARCHIVE -DestinationPath $TMP_DIR

# Move the binary to the install directory
Write-Output "Installing to $INSTALL_DIR (may require admin privileges)"
if (-Not (Test-Path -Path $INSTALL_DIR)) {
    New-Item -ItemType Directory -Path $INSTALL_DIR -Force
}
Move-Item -Path "$TMP_DIR\$FILENAME\$BINARY_NAME" -Destination "$INSTALL_DIR\$BINARY_NAME" -Force

# Check if .local/bin is in the PATH
$localBinPath = [System.IO.Path]::Combine($HOME, ".local", "bin")
if (-not ($env:PATH -split ";" | ForEach-Object { $_.Trim() } | Where-Object { $_ -eq $localBinPath })) {
    Write-Output "Adding $localBinPath to PATH"
    [System.Environment]::SetEnvironmentVariable("PATH", "$env:PATH;$localBinPath", [System.EnvironmentVariableTarget]::User)
} else {
    Write-Output "$localBinPath is already in PATH"
}

Write-Output "Installation complete! Verify with:"
Write-Output "hazyctl version"
