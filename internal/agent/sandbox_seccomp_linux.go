//go:build linux

package agent

const strictSandboxLauncher = `import ctypes
import os
import platform
import sys

libc = ctypes.CDLL(None, use_errno=True)

class SockFilter(ctypes.Structure):
    _fields_ = [("code", ctypes.c_ushort), ("jt", ctypes.c_ubyte), ("jf", ctypes.c_ubyte), ("k", ctypes.c_uint)]

class SockFprog(ctypes.Structure):
    _fields_ = [("length", ctypes.c_ushort), ("filter", ctypes.POINTER(SockFilter))]

BPF_LD = 0x00
BPF_W = 0x00
BPF_ABS = 0x20
BPF_JMP = 0x05
BPF_JEQ = 0x10
BPF_K = 0x00
BPF_RET = 0x06
SECCOMP_SET_MODE_FILTER = 1
SECCOMP_RET_KILL_PROCESS = 0x80000000
SECCOMP_RET_ERRNO = 0x00050000
SECCOMP_RET_ALLOW = 0x7fff0000
PR_SET_NO_NEW_PRIVS = 38

syscall_numbers = {
    "x86_64": {
        "arch": 0xc000003e,
        "seccomp": 317,
        "denied": [56, 101, 165, 166, 169, 175, 176, 246, 250, 272, 298, 303, 304, 308, 310, 311, 317, 321, 323, 435],
    },
    "amd64": {
        "arch": 0xc000003e,
        "seccomp": 317,
        "denied": [56, 101, 165, 166, 169, 175, 176, 246, 250, 272, 298, 303, 304, 308, 310, 311, 317, 321, 323, 435],
    },
    "aarch64": {
        "arch": 0xc00000b7,
        "seccomp": 277,
        "denied": [40, 41, 97, 104, 105, 106, 117, 142, 219, 220, 241, 264, 265, 268, 270, 271, 277, 280, 282, 435],
    },
    "arm64": {
        "arch": 0xc00000b7,
        "seccomp": 277,
        "denied": [40, 41, 97, 104, 105, 106, 117, 142, 219, 220, 241, 264, 265, 268, 270, 271, 277, 280, 282, 435],
    },
}
config = syscall_numbers.get(platform.machine())
if config is None:
    raise SystemExit("unsupported seccomp architecture")

instructions = [
    SockFilter(BPF_LD | BPF_W | BPF_ABS, 0, 0, 4),
    SockFilter(BPF_JMP | BPF_JEQ | BPF_K, 1, 0, config["arch"]),
    SockFilter(BPF_RET | BPF_K, 0, 0, SECCOMP_RET_KILL_PROCESS),
    SockFilter(BPF_LD | BPF_W | BPF_ABS, 0, 0, 0),
]
for number in config["denied"]:
    instructions.append(SockFilter(BPF_JMP | BPF_JEQ | BPF_K, 0, 1, number))
    instructions.append(SockFilter(BPF_RET | BPF_K, 0, 0, SECCOMP_RET_ERRNO | 1))
instructions.append(SockFilter(BPF_RET | BPF_K, 0, 0, SECCOMP_RET_ALLOW))

array_type = SockFilter * len(instructions)
array = array_type(*instructions)
program = SockFprog(len(instructions), ctypes.cast(array, ctypes.POINTER(SockFilter)))
if libc.prctl(PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0) != 0:
    raise OSError(ctypes.get_errno(), "PR_SET_NO_NEW_PRIVS failed")
if libc.syscall(config["seccomp"], SECCOMP_SET_MODE_FILTER, 0, ctypes.byref(program)) != 0:
    raise OSError(ctypes.get_errno(), "seccomp filter installation failed")
os.execv(sys.argv[1], sys.argv[1:])
`
