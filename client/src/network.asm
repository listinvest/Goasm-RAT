.386
.model flat, stdcall
option casemap :none

__UNICODE__     equ     1

include network.inc
include mem_alloc.inc

include /masm32/include/kernel32.inc
include /masm32/include/user32.inc
include /masm32/include/masm32rt.inc

includelib /masm32/lib/kernel32.lib
includelib /masm32/lib/user32.lib
includelib /masm32/lib/wsock32.lib


SOCKET_VERSION      equ     101h


.code
RecvData    proc    remote:SOCKET, data:ptr BYTE, data_size:DWORD
    local   @received:DWORD

    xor     ecx, ecx
    mov     @received, 0
    .while  ecx < data_size
        mov     ebx, data
        add     ebx, @received
        mov     edx, data_size
        sub     edx, @received
        invoke  recv, remote, ebx, edx, 0
        .if     eax == SOCKET_ERROR
            .break
        .endif
        mov     ecx, @received
        add     ecx, eax
        mov     @received, ecx
    .endw

    .if     eax != SOCKET_ERROR
        mov     eax, @received
    .endif
    ret
RecvData    endp


SendData    proc    remote:SOCKET, data:ptr BYTE, data_size:DWORD
    local   @sent:DWORD

    xor     ecx, ecx
    mov     @sent, 0
    .while  ecx < data_size
        mov     ebx, data
        add     ebx, @sent
        mov     edx, data_size
        sub     edx, @sent
        invoke  send, remote, ebx, edx, 0
        .if     eax == SOCKET_ERROR
            .break
        .endif
        mov     ecx, @sent
        add     ecx, eax
        mov     @sent, ecx
    .endw

    .if     eax != SOCKET_ERROR
        mov     eax, @sent
    .endif
    ret
SendData    endp


InitWinSocket       proc
    local   @wsa:WSADATA

    invoke  WSAStartup, SOCKET_VERSION, addr @wsa
    .if     eax != 0
        print   "Failed to initialize the socket library.", 0Dh, 0Ah
        mov     eax, FALSE
    .else
        mov     eax, TRUE
    .endif
    ret
InitWinSocket       endp


FreeWinSocket       proc
    invoke  WSACleanup
    ret
FreeWinSocket       endp

end