.386
.model flat, stdcall
option casemap :none

__UNICODE__     equ     1

include network.inc

include /masm32/include/windows.inc
include /masm32/include/kernel32.inc
include /masm32/include/user32.inc
include /masm32/include/masm32rt.inc

includelib /masm32/lib/kernel32.lib
includelib /masm32/lib/user32.lib


.code
start:
    print   "The client has started.", 0Dh, 0Ah
    invoke  InitWinSocket
    .if     eax == TRUE
        invoke  StartupService
        print   "The client has exited.", 0Dh, 0Ah
        invoke  FreeWinSocket
    .endif
    ret

end start