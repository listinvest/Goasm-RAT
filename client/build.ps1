New-Item -Name 'build' -ItemType 'Directory' -Force
Set-Location -Path '.\build'

Start-Process -FilePath 'ml' -ArgumentList '/c', '/coff', '../src/*.asm' -NoNewWindow -Wait
Start-Process -FilePath 'link' -ArgumentList '/subsystem:console', '/out:client.exe', '*.obj' -NoNewWindow -Wait

Remove-Item '.\*.obj'