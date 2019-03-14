<#
.SYNOPSIS
    Builds an MSI for mongosqld.exe and mongodrdl.exe.
.DESCRIPTION
    .
.PARAMETER ProjectName
    The name to use when referring to the project.
.PARAMETER VersionLabel
    The label to use when referring to this version of the 
    project.
.PARAMETER WixPath
    The path to the WIX binaries.
#>
Param(
  [string]$ProjectName,
  [string]$VersionLabel,
  [string]$WixPath
)

$wixUiExt = "$WixPath\WixUIExtension.dll"
$sourceDir = pwd
$resourceDir = "$sourceDir\release\installers\msi\mongosql\"
$artifactsDir = "$sourceDir\testdata\artifacts\"
$objDir = "$artifactsDir\out\"
$binDir = "$artifactsDir\build\"

if (-not ($VersionLabel -match "(\d\.\d).*")) {
    throw "invalid version specified: $VersionLabel"
}
$version = $matches[1]

# The upgrade code needs to change everytime we
# revise the minor version (2.2 -> 2.3). That way, we
# will allow multiple minor versions to be installed 
# side-by-side.
if ([double]$version -gt 2.10) {
    throw "You must change the upgrade code for a minor revision. 
Once that is done, change the version number above to
account for the next revision that will require being
upgradeable."
}

# You can get an upgrade code from https://www.uuidgenerator.net/
$upgradeCode = "b1231b8f-c980-4237-8e16-3542bd4a9143"

# compile wxs into .wixobjs
& $WixPath\candle.exe -wx `
    -dProductId="*" `
    -dUpgradeCode="$upgradeCode" `
    -dVersion="$version" `
    -dVersionLabel="$VersionLabel" `
    -dProjectName="$ProjectName" `
    -dSourceDir="$sourceDir" `
    -dResourceDir="$resourceDir" `
    -dSslDir="$binDir" `
    -dBinaryDir="$binDir" `
    -dTargetDir="$objDir" `
    -dTargetExt=".msi" `
    -dTargetFileName="release" `
    -dOutDir="$objDir" `
    -dConfiguration="Release" `
    -arch "x64" `
    -out "$objDir" `
    -ext "$wixUiExt" `
    "$resourceDir\Product.wxs" `
    "$resourceDir\FeatureFragment.wxs" `
    "$resourceDir\BinaryFragment.wxs" `
    "$resourceDir\LicensingFragment.wxs" `
    "$resourceDir\ConfigurationFragment.wxs" `
    "$resourceDir\UIFragment.wxs"

if(-not $?) {
    exit 1
}

# link wixobjs into an msi
& $WixPath\light.exe -wx `
    -cultures:en-us `
    -out "$artifactsDir\release.msi" `
    -ext "$wixUiExt" `
    $objDir\Product.wixobj `
    $objDir\FeatureFragment.wixobj `
    $objDir\BinaryFragment.wixobj `
    $objDir\LicensingFragment.wixobj `
    $objDir\ConfigurationFragment.wixobj `
    $objDir\UIFragment.wixobj

if(-not $?) {
    exit 1
}

trap {
  write-output $_
  exit 1
}
