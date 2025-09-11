import json
import os
import re
import shutil
import subprocess
import sys
from typing import List, Optional, Dict, Any

_DATA_DIR = os.path.abspath(os.getenv("GAB_DATA", os.path.join(os.getcwd(), "gab_data")))


def _gab_path() -> str:
    override = os.getenv("GAB_CLI")
    if override:
        return override
    exe = "gab.exe" if os.name == "nt" else "gab"
    path = shutil.which(exe)
    if not path:
        raise RuntimeError(f"Gab CLI not found. Ensure '{exe}' is on PATH or set GAB_CLI to its full path.")
    return path


def init(data_dir: Optional[str] = None) -> None:
    global _DATA_DIR
    if data_dir:
        _DATA_DIR = os.path.abspath(data_dir)
    cmd = [_gab_path(), "init", "--data", _DATA_DIR]
    subprocess.run(cmd, check=True)


_RECORDED_RE = re.compile(r"Recorded event\s+([0-9a-f]{64})")


def exec_cmd(args: List[str], *, agent_id: str, scope_dir: str, cwd: Optional[str] = None, env: Optional[Dict[str, str]] = None, data_dir: Optional[str] = None) -> Dict[str, Any]:
    data = os.path.abspath(data_dir or _DATA_DIR)
    cwd = cwd or os.getcwd()
    env_list = None
    if env is not None:
        env_list = [f"{k}={v}" for k, v in env.items()]

    cli = [_gab_path(), "record-cmd", "--data", data, "--agent", agent_id, "--scope", scope_dir, "--cwd", cwd, "--"] + args
    proc = subprocess.run(cli, capture_output=True, text=True, check=True, env=(os.environ if env_list is None else {**os.environ, **env}))
    out = proc.stdout or ""
    m = _RECORDED_RE.search(out)
    if not m:
        raise RuntimeError(f"Could not parse event id from output:\n{out}\n{proc.stderr}")
    event_id = m.group(1)

    show = subprocess.run([_gab_path(), "show", event_id, "--data", data], capture_output=True, text=True, check=True)
    return json.loads(show.stdout)


# Convenience alias
def run(args: List[str], **kwargs) -> Dict[str, Any]:
    return exec_cmd(args, **kwargs)


