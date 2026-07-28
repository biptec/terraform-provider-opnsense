#!/usr/bin/env python3
"""Create a short-lived API key through QEMU Guest Agent without logging it."""

import base64
import json
import os
import re
import select
import socket
import sys
import time

QEMU_GA_SOCKET = os.environ.get("QEMU_GA_SOCKET", "/tmp/qemu-virtserialport.sock")
USERNAME = os.environ.get("OPNSENSE_API_USERNAME", "terraform")
TIMEOUT = int(os.environ.get("OPNSENSE_API_KEY_TIMEOUT", "600"))


def qga(command, timeout=10):
    with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as client:
        client.settimeout(timeout)
        client.connect(QEMU_GA_SOCKET)
        client.sendall((json.dumps(command) + "\n").encode())
        response = b""
        deadline = time.monotonic() + timeout
        while time.monotonic() < deadline:
            ready, _, _ = select.select([client], [], [], 1)
            if not ready:
                continue
            chunk = client.recv(65536)
            if not chunk:
                break
            response += chunk
            if b"\n" in response:
                break
    if not response:
        raise RuntimeError("empty QEMU Guest Agent response")
    decoded = json.loads(response.splitlines()[0])
    if "error" in decoded:
        raise RuntimeError(f"QEMU Guest Agent error: {decoded['error']}")
    return decoded.get("return", {})


def guest_exec(path, args):
    started = qga({
        "execute": "guest-exec",
        "arguments": {"path": path, "arg": args, "capture-output": True},
    })
    pid = started.get("pid")
    if pid is None:
        raise RuntimeError("guest-exec did not return a PID")
    deadline = time.monotonic() + TIMEOUT
    while time.monotonic() < deadline:
        status = qga({"execute": "guest-exec-status", "arguments": {"pid": pid}})
        if not status.get("exited"):
            time.sleep(1)
            continue
        stdout = base64.b64decode(status.get("out-data", "")).decode(errors="replace")
        stderr = base64.b64decode(status.get("err-data", "")).decode(errors="replace")
        code = status.get("exitcode", 1)
        if code != 0:
            raise RuntimeError(f"guest command failed with exit code {code}: {stderr.strip()}")
        return stdout
    raise TimeoutError("guest command did not finish before timeout")


def wait_for_bootstrap():
    deadline = time.monotonic() + TIMEOUT
    while time.monotonic() < deadline:
        try:
            result = guest_exec("/bin/sh", ["-c", "test -f /conf/bootstrap.done"])
            del result
            return
        except RuntimeError:
            time.sleep(2)
    raise TimeoutError("NoCloud bootstrap did not complete before timeout")


def main():
    if not re.fullmatch(r"[a-z_][a-z0-9_-]{0,31}", USERNAME):
        raise RuntimeError("invalid API username")
    wait_for_bootstrap()
    command = (
        "tool=$(command -v opnsense-apikey || true); "
        "test -n \"$tool\" || exit 127; "
        f'exec "$tool" -u {USERNAME!r} --json create'
    )
    output = guest_exec("/bin/sh", ["-c", command])
    try:
        credentials = json.loads(output)
        key = credentials["key"]
        secret = credentials["secret"]
    except (json.JSONDecodeError, KeyError, TypeError) as exc:
        raise RuntimeError("opnsense-apikey returned an invalid response") from exc
    if not key or not secret or "\n" in key or "\n" in secret:
        raise RuntimeError("opnsense-apikey returned invalid credentials")

    # The workflow redirects stderr directly to GITHUB_OUTPUT. Do not write
    # credentials to stdout, where they would become part of the action log.
    print(f"key={key}", file=sys.stderr)
    print(f"secret={secret}", file=sys.stderr)
    print(f"Created a temporary API key for {USERNAME}")


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        print(f"ERROR: {exc}")
        sys.exit(1)
