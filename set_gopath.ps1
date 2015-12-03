$goPath = "${pwd}\.gopath"
$vendorPath = "${pwd}\vendor"

# Using cmd invocation to recursively delete directories because Remove-Item -Recurse -Force 
# has a bug causing the script to fail.
Invoke-Expression "cmd /c rd /s /q $goPath"
New-Item $goPath\src\github.com\10gen -ItemType Container | Out-Null
Invoke-Expression "cmd /c mklink /J $goPath\src\github.com\10gen\sqlproxy ${pwd}" | Out-Null
$env:GOPATH = "$goPath;$vendorPath"