#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import tempfile
import urllib.parse
import urllib.request
from pathlib import Path

CONFIG_FILE = Path(os.environ.get("CODE_OS_CONFIG", Path.home() / ".config/code-os/config.json"))
CONFIG = json.loads(CONFIG_FILE.read_text(encoding="utf-8"))
CLOUDFLARE = CONFIG["cloudflare"]
ACCOUNT_ID = CLOUDFLARE["accountId"]
ZONE_ID = CLOUDFLARE["zoneId"]
TUNNEL_ID = CLOUDFLARE["tunnelId"]
TUNNEL_TARGET = f"{TUNNEL_ID}.cfargotunnel.com"
API_ROOT = "https://api.cloudflare.com/client/v4"
TOKEN_FILE = Path(CLOUDFLARE["tokenFile"])
STATE_FILE = Path("/var/lib/code-os-port-sync/state.json")
PORTLY_BIN = CONFIG["portlyBinary"]
PORT_TEMPLATE = CONFIG["publicPortHost"]
if PORT_TEMPLATE.count("{port}") != 1:
    raise RuntimeError("Code OS publicPortHost must contain exactly one {port} placeholder")
PORT_HOST = re.compile("^" + re.escape(PORT_TEMPLATE).replace(r"\{port\}", r"(?P<port>\d+)") + "$")
BLOCKED_PORTS = {5432, 6380, 7737, 8079, 8787, 20241}


def api(path: str, method: str = "GET", body: object | None = None) -> object:
    token = TOKEN_FILE.read_text(encoding="utf-8").strip()
    data = None if body is None else json.dumps(body).encode()
    request = urllib.request.Request(
        f"{API_ROOT}{path}",
        data=data,
        method=method,
        headers={"Authorization": f"Bearer {token}", "Content-Type": "application/json"},
    )
    with urllib.request.urlopen(request, timeout=20) as response:
        payload = json.load(response)
    if not payload.get("success"):
        raise RuntimeError(json.dumps(payload.get("errors", [])))
    return payload.get("result")


def load_state() -> dict:
    if not STATE_FILE.exists():
        return {"records": {}, "disabledPorts": []}
    return json.loads(STATE_FILE.read_text(encoding="utf-8"))


def save_state(state: dict) -> None:
    STATE_FILE.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.NamedTemporaryFile("w", dir=STATE_FILE.parent, delete=False) as handle:
        json.dump(state, handle, indent=2, sort_keys=True)
        handle.write("\n")
        temporary = Path(handle.name)
    os.chmod(temporary, 0o600)
    temporary.replace(STATE_FILE)


def allowed(port: int) -> bool:
    return 1024 <= port <= 65535 and port not in BLOCKED_PORTS


def portly_ports() -> set[int]:
    payload = json.loads(subprocess.check_output([PORTLY_BIN, "status", "--json"], text=True))
    ports: set[int] = set()
    for project in payload.get("projects", []):
        for server in project.get("servers", []):
            port = server.get("port")
            if server.get("state") == "running" and server.get("healthy") is True and isinstance(port, int) and allowed(port):
                ports.add(port)
    return ports


def hostname(port: int) -> str:
    return PORT_TEMPLATE.replace("{port}", str(port))


def find_dns(name: str) -> list[dict]:
    query = urllib.parse.urlencode({"type": "CNAME", "name": name, "per_page": 100})
    result = api(f"/zones/{ZONE_ID}/dns_records?{query}")
    return result if isinstance(result, list) else []


def ensure_dns(port: int, state: dict) -> None:
    key = str(port)
    name = hostname(port)
    existing = find_dns(name)
    body = {"type": "CNAME", "name": name, "content": TUNNEL_TARGET, "ttl": 1, "proxied": True}
    if existing:
        record = existing[0]
        if record.get("content") != TUNNEL_TARGET or record.get("proxied") is not True:
            record = api(f"/zones/{ZONE_ID}/dns_records/{record['id']}", "PUT", body)
        previous = state["records"].get(key, {})
        state["records"][key] = {"id": record["id"], "owned": bool(previous.get("owned", False))}
        return
    record = api(f"/zones/{ZONE_ID}/dns_records", "POST", body)
    state["records"][key] = {"id": record["id"], "owned": True}


def remove_dns(port: int, state: dict) -> None:
    entry = state.get("records", {}).pop(str(port), None)
    if entry and entry.get("owned"):
        try:
            api(f"/zones/{ZONE_ID}/dns_records/{entry['id']}", "DELETE")
        except Exception as error:
            print(f"warning: could not delete DNS for {port}: {error}")


def update_ingress(ports: set[int]) -> None:
    result = api(f"/accounts/{ACCOUNT_ID}/cfd_tunnel/{TUNNEL_ID}/configurations")
    config = result.get("config", {}) if isinstance(result, dict) else {}
    ingress = config.get("ingress", [])
    preserved = [rule for rule in ingress if rule.get("service") != "http_status:404" and not PORT_HOST.match(str(rule.get("hostname", "")))]
    managed = [
        {
            "hostname": hostname(port),
            "service": "http://127.0.0.1:7890",
        }
        for port in sorted(ports)
    ]
    next_config = {**config, "ingress": [*preserved, *managed, {"service": "http_status:404"}]}
    if next_config != config:
        api(f"/accounts/{ACCOUNT_ID}/cfd_tunnel/{TUNNEL_ID}/configurations", "PUT", {"config": next_config})


def update_firewall(ports: set[int]) -> None:
    subprocess.run(["nft", "delete", "table", "inet", "code_os"], check=False, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    elements = ", ".join(str(port) for port in sorted(ports))
    ruleset = f"""table inet code_os {{
  set protected_ports {{ type inet_service; elements = {{ {elements} }} }}
  chain input {{
    type filter hook input priority -5; policy accept;
    iifname != \"lo\" tcp dport @protected_ports drop
  }}
}}
"""
    subprocess.run(["nft", "-f", "-"], input=ruleset, text=True, check=True)


def sync() -> dict:
    state = load_state()
    disabled = {int(port) for port in state.get("disabledPorts", [])}
    desired = portly_ports() - disabled
    current = {int(port) for port in state.get("records", {})}
    for port in sorted(desired):
        ensure_dns(port, state)
    for port in sorted(current - desired):
        remove_dns(port, state)
    update_ingress(desired)
    update_firewall(desired)
    state["activePorts"] = sorted(desired)
    save_state(state)
    return state


def set_disabled(port: int, disabled: bool) -> dict:
    if not allowed(port):
        raise SystemExit(f"port {port} is blocked")
    state = load_state()
    values = {int(value) for value in state.get("disabledPorts", [])}
    values.add(port) if disabled else values.discard(port)
    state["disabledPorts"] = sorted(values)
    save_state(state)
    return sync()


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("command", choices=["sync", "from-portly", "status", "open", "close"])
    parser.add_argument("port", nargs="?", type=int)
    args = parser.parse_args()
    if args.command == "status":
        state = load_state()
    elif args.command in {"sync", "from-portly"}:
        state = sync()
    else:
        if args.port is None:
            parser.error("open and close require a port")
        state = set_disabled(args.port, args.command == "close")
    print(json.dumps({"active": [f"https://{hostname(port)}/" for port in state.get("activePorts", [])], "disabledPorts": state.get("disabledPorts", [])}, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
