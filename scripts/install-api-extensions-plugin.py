#!/usr/bin/env python3

"""Install the exact os-api-extensions candidate in the OPNsense CI guest."""

import base64
import json
import os
import re
import select
import socket
import sys
import time

QEMU_GA_SOCKET = os.environ.get("QEMU_GA_SOCKET", "/tmp/qemu-virtserialport.sock")
COMMIT = os.environ.get("OPNSENSE_API_EXTENSIONS_PLUGIN_COMMIT", "")
COMMIT_RE = re.compile(r"^[0-9a-f]{40}$")


def send(command, timeout=10):
    with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as client:
        client.settimeout(timeout)
        client.connect(QEMU_GA_SOCKET)
        client.sendall((json.dumps(command) + "\n").encode())
        response = b""
        while b"\n" not in response:
            ready, _, _ = select.select([client], [], [], timeout)
            if not ready:
                raise TimeoutError("QEMU guest agent response timed out")
            chunk = client.recv(65536)
            if not chunk:
                break
            response += chunk
    payload = json.loads(response.decode().strip())
    if "error" in payload:
        raise RuntimeError(payload["error"].get("desc", "guest agent error"))
    return payload.get("return", {})


def guest_exec(script):
    result = send({
        "execute": "guest-exec",
        "arguments": {
            "path": "/bin/sh",
            "arg": ["-c", script],
            "capture-output": True,
        },
    })
    pid = result["pid"]
    deadline = time.time() + 900
    while time.time() < deadline:
        status = send({"execute": "guest-exec-status", "arguments": {"pid": pid}})
        if status.get("exited"):
            stdout = base64.b64decode(status.get("out-data", "")).decode(errors="replace")
            stderr = base64.b64decode(status.get("err-data", "")).decode(errors="replace")
            if status.get("exitcode", 1) != 0:
                detail = (stderr or stdout)[-8000:]
                raise RuntimeError(f"os-api-extensions installation failed:\n{detail}")
            return stdout
        time.sleep(2)
    raise TimeoutError("os-api-extensions installation timed out")


def build_install_script(commit):
    return f"""set -eu
if [ ! -d /usr/plugins/.git ]; then
    opnsense-code plugins
fi
git -C /usr/plugins fetch --depth 1 https://github.com/biptec/opnsense-plugins.git {commit}
git -C /usr/plugins checkout --detach FETCH_HEAD
make -C /usr/plugins/sysutils/api-extensions upgrade
pkg info -e 'os-api-extensions-*'
installed_hash=$(pkg query '%At:%Av' os-api-extensions | sed -n 's/^product_hash://p')
test -n "$installed_hash"
case "{commit}" in
    "$installed_hash"*) ;;
    *) echo "installed os-api-extensions hash does not match requested commit" >&2; exit 1 ;;
esac
test -f /usr/local/opnsense/mvc/app/controllers/OPNsense/ApiExtensions/Api/WebguiController.php
"""


def main():
    if not COMMIT_RE.fullmatch(COMMIT):
        raise RuntimeError("OPNSENSE_API_EXTENSIONS_PLUGIN_COMMIT must be a full Git commit SHA")
    guest_exec(build_install_script(COMMIT))
    print(f"Installed os-api-extensions candidate {COMMIT[:12]}")


if __name__ == "__main__":
    try:
        main()
    except Exception as error:
        print(f"ERROR: {error}", file=sys.stderr)
        sys.exit(1)
