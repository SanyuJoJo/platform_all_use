"""Python 插件示例，实现 contract/http_protocol.go 中定义的 HTTP 契约。"""
import json
import os
import signal
import sys

from flask import Flask, jsonify, request

app = Flask(__name__)

PLUGIN_NAME = "python-hello"
PLUGIN_VERSION = "1.0.0"
_state = {"greeting": "Hello"}


@app.get("/health")
def health():
    return jsonify(status="ok", name=PLUGIN_NAME, version=PLUGIN_VERSION)


@app.post("/initialize")
def initialize():
    body = request.get_json(silent=True) or {}
    config = body.get("config") or {}
    if "greeting" in config:
        _state["greeting"] = str(config["greeting"])
    return jsonify(error="")


@app.post("/execute")
def execute():
    body = request.get_json(silent=True) or {}
    name = body.get("input", "world")
    return jsonify(output=f"{_state['greeting']}, {name}!", error="")


@app.post("/shutdown")
def shutdown():
    return jsonify(error="")


# —— 路由转发端点 ——
# host 会把 /api/hello/* 的请求转发到此插件的对应路径。
# Flask 会按规则具体性优先匹配 /health、/initialize 等具体路由；
# 剩下的路径落入 catch-all，返回请求信息作为示例响应。
@app.route("/<path:subpath>", methods=["GET", "POST", "PUT", "DELETE", "PATCH"])
def catch_all(subpath):
    return jsonify(
        plugin=PLUGIN_NAME,
        method=request.method,
        path="/" + subpath,
        query=dict(request.args),
        body=request.get_data(as_text=True),
        greeting=_state["greeting"],
    )


def _graceful(signum, frame):
    sys.exit(0)


if __name__ == "__main__":
    port = int(os.environ.get("PLUGIN_PORT", "9100"))
    signal.signal(signal.SIGINT, _graceful)
    signal.signal(signal.SIGTERM, _graceful)
    app.run(host="127.0.0.1", port=port)
