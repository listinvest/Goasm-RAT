New-Variable -Name 'AppName' -Value 'server' -Option Constant

Set-Location -Path '..'
New-Item -Name 'build' -ItemType 'Directory' -Force

Set-Location -Path ".\cmd\$AppName"
Start-Process -FilePath 'go' -ArgumentList 'build', "-o $AppName.exe" -NoNewWindow -Wait
Move-Item -Path ".\$AppName.exe" -Destination '..\..\build' -Force