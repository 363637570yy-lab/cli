package management

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// dynamicProxyPageHTML 是动态代理管理页的完整 HTML 内容（内嵌在二进制中）。
const dynamicProxyPageHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>动态代理管理 · CLIProxyAPI</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
<style>
  :root {
    --bg: #0d0f14;
    --surface: #161a24;
    --surface2: #1e2433;
    --border: #2a3149;
    --accent: #6366f1;
    --accent-hover: #818cf8;
    --accent-glow: rgba(99,102,241,0.18);
    --success: #22c55e;
    --error: #ef4444;
    --warn: #f59e0b;
    --text: #e2e8f0;
    --text-muted: #64748b;
    --text-dim: #94a3b8;
    --radius: 12px;
    --radius-sm: 8px;
    --transition: 0.2s ease;
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: 'Inter', sans-serif;
    background: var(--bg);
    color: var(--text);
    min-height: 100vh;
    padding: 0;
  }

  /* ── header ──────────────────────────────────── */
  .header {
    background: linear-gradient(135deg, #1a1f33 0%, #0d1117 100%);
    border-bottom: 1px solid var(--border);
    padding: 20px 40px;
    display: flex;
    align-items: center;
    gap: 16px;
    position: sticky; top: 0; z-index: 100;
    backdrop-filter: blur(12px);
  }
  .header-icon {
    width: 40px; height: 40px;
    background: var(--accent-glow);
    border: 1px solid var(--accent);
    border-radius: 10px;
    display: flex; align-items: center; justify-content: center;
    font-size: 20px;
  }
  .header-title { font-size: 18px; font-weight: 600; color: var(--text); }
  .header-sub { font-size: 12px; color: var(--text-muted); margin-top: 2px; }
  .header-badge {
    margin-left: auto;
    background: var(--accent-glow);
    border: 1px solid var(--accent);
    color: var(--accent-hover);
    font-size: 11px; font-weight: 600;
    padding: 4px 10px; border-radius: 20px;
  }

  /* ── layout ──────────────────────────────────── */
  .container { max-width: 860px; margin: 0 auto; padding: 40px 24px; }

  /* ── auth gate ───────────────────────────────── */
  #auth-gate {
    display: flex; justify-content: center; align-items: center;
    min-height: calc(100vh - 88px);
  }
  .auth-card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 40px;
    width: 100%; max-width: 420px;
    text-align: center;
  }
  .auth-card h2 { font-size: 20px; margin-bottom: 8px; }
  .auth-card p { color: var(--text-muted); font-size: 14px; margin-bottom: 24px; }

  /* ── card ────────────────────────────────────── */
  .card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 28px;
    margin-bottom: 20px;
    transition: border-color var(--transition);
  }
  .card:hover { border-color: #3a4160; }
  .card-header {
    display: flex; align-items: center; gap: 10px;
    margin-bottom: 22px;
  }
  .card-icon {
    width: 36px; height: 36px;
    background: var(--accent-glow);
    border-radius: var(--radius-sm);
    display: flex; align-items: center; justify-content: center;
    font-size: 16px;
  }
  .card-title { font-size: 15px; font-weight: 600; }
  .card-desc { font-size: 12px; color: var(--text-muted); margin-top: 2px; }

  /* ── toggle ──────────────────────────────────── */
  .toggle-row {
    display: flex; align-items: center; justify-content: space-between;
    padding: 14px 18px;
    background: var(--surface2);
    border-radius: var(--radius-sm);
    margin-bottom: 20px;
  }
  .toggle-label { font-weight: 500; font-size: 14px; }
  .toggle-hint { font-size: 12px; color: var(--text-muted); margin-top: 3px; }
  .switch { position: relative; width: 48px; height: 26px; flex-shrink: 0; }
  .switch input { opacity: 0; width: 0; height: 0; }
  .slider {
    position: absolute; inset: 0;
    background: #2a3149;
    border-radius: 26px;
    cursor: pointer;
    transition: background var(--transition);
  }
  .slider::before {
    content: '';
    position: absolute;
    width: 20px; height: 20px;
    left: 3px; top: 3px;
    background: white;
    border-radius: 50%;
    transition: transform var(--transition);
  }
  .switch input:checked + .slider { background: var(--accent); }
  .switch input:checked + .slider::before { transform: translateX(22px); }

  /* ── form fields ─────────────────────────────── */
  .field { margin-bottom: 18px; }
  .field label {
    display: block;
    font-size: 13px; font-weight: 500;
    color: var(--text-dim);
    margin-bottom: 8px;
  }
  .field label span {
    font-size: 11px; font-weight: 400;
    color: var(--text-muted);
    margin-left: 6px;
  }
  .input, .select {
    width: 100%;
    background: var(--surface2);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    color: var(--text);
    font-family: inherit;
    font-size: 14px;
    padding: 10px 14px;
    outline: none;
    transition: border-color var(--transition), box-shadow var(--transition);
  }
  .input:focus, .select:focus {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--accent-glow);
  }
  .input.mono { font-family: 'JetBrains Mono', 'Fira Code', monospace; font-size: 13px; }
  .input-row { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }

  /* ── hint ────────────────────────────────────── */
  .hint {
    font-size: 12px; color: var(--text-muted);
    margin-top: 6px;
    line-height: 1.5;
  }

  /* ── status badge ─────────────────────────────── */
  .status-indicator {
    display: flex; align-items: center; gap: 10px;
    padding: 12px 16px;
    border-radius: var(--radius-sm);
    margin-bottom: 20px;
    font-size: 13px;
    background: var(--surface2);
    border: 1px solid var(--border);
  }
  .status-dot {
    width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0;
    animation: pulse 2s infinite;
  }
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.4; }
  }
  .status-dot.green { background: var(--success); box-shadow: 0 0 6px var(--success); }
  .status-dot.red   { background: var(--error);   box-shadow: 0 0 6px var(--error); }
  .status-dot.gray  { background: var(--text-muted); animation: none; }

  /* ── buttons ─────────────────────────────────── */
  .btn-row { display: flex; gap: 10px; flex-wrap: wrap; margin-top: 24px; }
  .btn {
    display: inline-flex; align-items: center; gap: 8px;
    padding: 10px 20px;
    border-radius: var(--radius-sm);
    font-family: inherit; font-size: 14px; font-weight: 500;
    cursor: pointer; border: none;
    transition: all var(--transition);
    text-decoration: none;
  }
  .btn-primary {
    background: var(--accent);
    color: white;
  }
  .btn-primary:hover { background: var(--accent-hover); transform: translateY(-1px); box-shadow: 0 4px 15px var(--accent-glow); }
  .btn-secondary {
    background: var(--surface2);
    color: var(--text-dim);
    border: 1px solid var(--border);
  }
  .btn-secondary:hover { border-color: var(--accent); color: var(--accent-hover); }
  .btn-danger {
    background: rgba(239,68,68,0.1);
    color: var(--error);
    border: 1px solid rgba(239,68,68,0.3);
  }
  .btn-danger:hover { background: rgba(239,68,68,0.2); }
  .btn:disabled { opacity: 0.5; cursor: not-allowed; transform: none !important; }

  /* ── toast ───────────────────────────────────── */
  #toast {
    position: fixed; bottom: 30px; right: 30px;
    padding: 14px 20px;
    border-radius: var(--radius-sm);
    font-size: 14px; font-weight: 500;
    display: none;
    animation: slideIn 0.3s ease;
    z-index: 9999;
    max-width: 400px;
    box-shadow: 0 8px 30px rgba(0,0,0,0.4);
  }
  @keyframes slideIn {
    from { transform: translateY(20px); opacity: 0; }
    to   { transform: translateY(0);    opacity: 1; }
  }
  .toast-success { background: rgba(34,197,94,0.15); border: 1px solid var(--success); color: var(--success); }
  .toast-error   { background: rgba(239,68,68,0.15);  border: 1px solid var(--error);   color: var(--error);  }
  .toast-info    { background: rgba(99,102,241,0.15); border: 1px solid var(--accent);  color: var(--accent-hover); }

  /* ── test result ─────────────────────────────── */
  .test-result {
    margin-top: 16px;
    padding: 14px 16px;
    border-radius: var(--radius-sm);
    font-size: 13px;
    line-height: 1.6;
    display: none;
    word-break: break-all;
  }
  .test-result.success { background: rgba(34,197,94,0.08); border: 1px solid rgba(34,197,94,0.3); }
  .test-result.error   { background: rgba(239,68,68,0.08);  border: 1px solid rgba(239,68,68,0.3); }
  .test-result .proxy-url {
    font-family: 'JetBrains Mono', monospace;
    font-size: 12px;
    background: rgba(0,0,0,0.3);
    padding: 6px 10px;
    border-radius: 6px;
    margin-top: 8px;
    display: block;
  }

  /* ── divider ─────────────────────────────────── */
  .divider { height: 1px; background: var(--border); margin: 20px 0; }

  /* ── spinner ─────────────────────────────────── */
  .spinner {
    display: inline-block;
    width: 14px; height: 14px;
    border: 2px solid rgba(255,255,255,0.3);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }

  /* ── priority legend ─────────────────────────── */
  .priority-list { display: flex; flex-direction: column; gap: 8px; }
  .priority-item {
    display: flex; align-items: center; gap: 10px;
    font-size: 13px; color: var(--text-dim);
  }
  .priority-badge {
    width: 24px; height: 24px; border-radius: 50%;
    display: flex; align-items: center; justify-content: center;
    font-size: 11px; font-weight: 700; flex-shrink: 0;
  }
  .p0 { background: rgba(99,102,241,0.2); color: var(--accent-hover); border: 1px solid var(--accent); }
  .p1 { background: rgba(34,197,94,0.1); color: var(--success); border: 1px solid rgba(34,197,94,0.4); }
  .p2 { background: rgba(100,116,139,0.2); color: var(--text-muted); border: 1px solid var(--border); }

  /* ── responsive ──────────────────────────────── */
  @media (max-width: 600px) {
    .header { padding: 16px 20px; }
    .container { padding: 24px 16px; }
    .input-row { grid-template-columns: 1fr; }
    .btn-row { flex-direction: column; }
  }
</style>
</head>
<body>

<div class="header">
  <div class="header-icon">🔀</div>
  <div>
    <div class="header-title">动态代理管理</div>
    <div class="header-sub">CLIProxyAPI · Dynamic Proxy Configuration</div>
  </div>
  <div class="header-badge" id="status-badge">未连接</div>
</div>

<!-- Auth Gate -->
<div id="auth-gate">
  <div class="auth-card">
    <div style="font-size:40px;margin-bottom:16px">🔑</div>
    <h2>管理员认证</h2>
    <p>请输入管理密钥以访问动态代理配置面板</p>
    <div class="field">
      <input class="input" type="password" id="mgmt-key" placeholder="Management Key" autocomplete="off">
    </div>
    <div class="btn-row" style="justify-content:center">
      <button class="btn btn-primary" onclick="doAuth()">🔓 验证并进入</button>
    </div>
    <div id="auth-error" style="color:var(--error);font-size:13px;margin-top:14px;display:none"></div>
  </div>
</div>

<!-- Main Page (hidden until auth) -->
<div id="main-page" style="display:none">
<div class="container">

  <!-- Status -->
  <div class="card">
    <div class="card-header">
      <div class="card-icon">📡</div>
      <div>
        <div class="card-title">当前状态</div>
        <div class="card-desc">动态代理运行状态及优先级说明</div>
      </div>
    </div>

    <div id="status-row" class="status-indicator">
      <span class="status-dot gray"></span>
      <span id="status-text">加载中...</span>
    </div>

    <div class="priority-list">
      <div class="priority-item">
        <span class="priority-badge p0">0</span>
        <span><strong style="color:var(--text)">动态代理 API</strong>（本配置，最高优先级）</span>
      </div>
      <div class="priority-item">
        <span class="priority-badge p1">1</span>
        <span>凭证级 proxy-url（auth.ProxyURL）</span>
      </div>
      <div class="priority-item">
        <span class="priority-badge p2">2</span>
        <span>全局静态 proxy-url（config.yaml）</span>
      </div>
    </div>
  </div>

  <!-- Config Form -->
  <div class="card">
    <div class="card-header">
      <div class="card-icon">⚙️</div>
      <div>
        <div class="card-title">代理配置</div>
        <div class="card-desc">设置动态代理 API 及轮换策略</div>
      </div>
    </div>

    <!-- Enable Toggle -->
    <div class="toggle-row">
      <div>
        <div class="toggle-label">启用动态代理</div>
        <div class="toggle-hint">开启后，所有凭证的出站请求将通过动态代理 IP 转发</div>
      </div>
      <label class="switch">
        <input type="checkbox" id="enable">
        <span class="slider"></span>
      </label>
    </div>

    <!-- API URL -->
    <div class="field">
      <label>代理 API 地址 <span>返回代理 URL 的 HTTP 接口</span></label>
      <input class="input mono" type="text" id="api-url"
        placeholder="https://api.example.com/get-proxy?time=10">
      <div class="hint">
        响应可以是纯文本代理 URL（如 <code style="font-family:monospace;color:var(--accent-hover)">http://user:pass@1.2.3.4:8080</code>）
        或 JSON 对象。<br>
        如果 URL 中包含 <code style="font-family:monospace;color:var(--accent-hover)">time=N</code> 参数，TTL 将自动设为 N 分钟。
      </div>
    </div>

    <!-- API Key -->
    <div class="input-row">
      <div class="field">
        <label>API 密钥 <span>可选</span></label>
        <input class="input" type="password" id="api-key" placeholder="留空则不发送密钥" autocomplete="off">
      </div>
      <div class="field">
        <label>密钥请求头 <span>默认 X-API-Key</span></label>
        <input class="input" type="text" id="api-key-header" placeholder="X-API-Key">
      </div>
    </div>

    <!-- Result Field -->
    <div class="field">
      <label>JSON 字段路径 <span>可选，留空则使用响应原文</span></label>
      <input class="input" type="text" id="result-field"
        placeholder="例如: data.proxy">
      <div class="hint">点号分隔的 JSON 路径，用于从 JSON 响应中提取代理 URL。
        例如响应为 <code style="font-family:monospace;color:var(--accent-hover)">{"data":{"proxy":"..."}}</code>
        时填写 <code style="font-family:monospace;color:var(--accent-hover)">data.proxy</code>。
      </div>
    </div>

    <div class="divider"></div>

    <!-- TTL + requestsPerIP -->
    <div class="input-row">
      <div class="field">
        <label>TTL 秒数 <span>0 = 从 URL time=N 自动推算</span></label>
        <input class="input" type="number" id="ttl-seconds" min="0" placeholder="0">
      </div>
      <div class="field">
        <label>每 IP 请求数限制 <span>0 = 仅 TTL 轮换</span></label>
        <input class="input" type="number" id="requests-per-ip" min="0" placeholder="0">
        <div class="hint">每个代理 IP 最多服务 N 次请求后自动换新 IP。</div>
      </div>
    </div>

    <!-- Buttons -->
    <div class="btn-row">
      <button class="btn btn-primary" id="btn-save" onclick="saveConfig()">
        💾 保存配置
      </button>
      <button class="btn btn-secondary" id="btn-test" onclick="testProxy()">
        🧪 测试连接
      </button>
      <button class="btn btn-danger" id="btn-clear" onclick="clearConfig()">
        🗑 清空配置
      </button>
    </div>

    <!-- Test Result -->
    <div class="test-result" id="test-result"></div>
  </div>

</div>
</div>

<div id="toast"></div>

<script>
let MGMT_KEY = '';
const BASE = '/v0/management';

// ── Auth ───────────────────────────────────────────────────────────────────

async function doAuth() {
  const key = document.getElementById('mgmt-key').value.trim();
  if (!key) { showAuthError('请输入管理密钥'); return; }

  try {
    const r = await fetch(BASE + '/dynamic-proxy', {
      headers: { 'Authorization': 'Bearer ' + key }
    });
    if (r.status === 401 || r.status === 403) {
      showAuthError('密钥错误或无权限');
      return;
    }
    MGMT_KEY = key;
    document.getElementById('auth-gate').style.display = 'none';
    document.getElementById('main-page').style.display = 'block';
    document.getElementById('status-badge').textContent = '已连接';
    document.getElementById('status-badge').style.color = 'var(--success)';
    const data = await r.json();
    fillForm(data);
    updateStatus(data);
  } catch(e) {
    showAuthError('连接失败: ' + e.message);
  }
}

function showAuthError(msg) {
  const el = document.getElementById('auth-error');
  el.textContent = msg;
  el.style.display = 'block';
}

document.getElementById('mgmt-key').addEventListener('keydown', e => {
  if (e.key === 'Enter') doAuth();
});

// ── Helpers ────────────────────────────────────────────────────────────────

function headers() {
  return {
    'Authorization': 'Bearer ' + MGMT_KEY,
    'Content-Type': 'application/json'
  };
}

function fillForm(d) {
  document.getElementById('enable').checked   = !!d.enable;
  document.getElementById('api-url').value    = d['api-url']          || '';
  document.getElementById('api-key').value    = d['api-key']          || '';
  document.getElementById('api-key-header').value = d['api-key-header'] || '';
  document.getElementById('result-field').value   = d['result-field']   || '';
  document.getElementById('ttl-seconds').value    = d['ttl-seconds']    != null ? d['ttl-seconds'] : 0;
  document.getElementById('requests-per-ip').value= d['requests-per-ip']!= null ? d['requests-per-ip'] : 0;
}

function collectForm() {
  return {
    enable:           document.getElementById('enable').checked,
    'api-url':        document.getElementById('api-url').value.trim(),
    'api-key':        document.getElementById('api-key').value.trim(),
    'api-key-header': document.getElementById('api-key-header').value.trim(),
    'result-field':   document.getElementById('result-field').value.trim(),
    'ttl-seconds':    parseInt(document.getElementById('ttl-seconds').value) || 0,
    'requests-per-ip':parseInt(document.getElementById('requests-per-ip').value) || 0,
  };
}

function updateStatus(d) {
  const row  = document.getElementById('status-row');
  const dot  = row.querySelector('.status-dot');
  const text = document.getElementById('status-text');
  if (d.enable && d['api-url']) {
    dot.className = 'status-dot green';
    const rpi = d['requests-per-ip'] > 0 ? '，每 IP 限 ' + d['requests-per-ip'] + ' 次请求' : '';
    const ttl = d['ttl-seconds'] > 0  ? '，TTL ' + d['ttl-seconds'] + 's' : '（TTL 自动推算）';
    text.textContent = '已启用' + ttl + rpi;
  } else if (d.enable && !d['api-url']) {
    dot.className = 'status-dot red';
    text.textContent = '已启用，但未配置 API 地址';
  } else {
    dot.className = 'status-dot gray';
    dot.style.animation = 'none';
    text.textContent = '已禁用（使用静态 proxy-url）';
  }
}

// ── Toast ──────────────────────────────────────────────────────────────────

let toastTimer;
function showToast(msg, type='info') {
  const t = document.getElementById('toast');
  t.textContent = msg;
  t.className = 'toast-' + type;
  t.style.display = 'block';
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => t.style.display = 'none', 3500);
}

// ── Actions ────────────────────────────────────────────────────────────────

async function saveConfig() {
  const btn = document.getElementById('btn-save');
  btn.disabled = true;
  btn.innerHTML = '<span class="spinner"></span> 保存中...';

  const body = collectForm();
  try {
    const r = await fetch(BASE + '/dynamic-proxy', {
      method: 'PUT',
      headers: headers(),
      body: JSON.stringify(body)
    });
    const j = await r.json();
    if (r.ok) {
      showToast('✅ 配置已保存', 'success');
      updateStatus(body);
    } else {
      showToast('❌ 保存失败: ' + (j.error || j.message || r.status), 'error');
    }
  } catch(e) {
    showToast('❌ 网络错误: ' + e.message, 'error');
  } finally {
    btn.disabled = false;
    btn.innerHTML = '💾 保存配置';
  }
}

async function testProxy() {
  const apiUrl = document.getElementById('api-url').value.trim();
  if (!apiUrl) { showToast('⚠️ 请先填写 API 地址', 'error'); return; }

  const btn = document.getElementById('btn-test');
  btn.disabled = true;
  btn.innerHTML = '<span class="spinner"></span> 测试中...';

  const resEl = document.getElementById('test-result');
  resEl.style.display = 'none';

  try {
    const r = await fetch(BASE + '/dynamic-proxy/test', {
      method: 'POST',
      headers: headers(),
      body: JSON.stringify({
        'api-url':        apiUrl,
        'api-key':        document.getElementById('api-key').value.trim(),
        'api-key-header': document.getElementById('api-key-header').value.trim(),
        'result-field':   document.getElementById('result-field').value.trim(),
      })
    });
    const j = await r.json();

    if (j.ok) {
      resEl.className = 'test-result success';
      resEl.innerHTML =
        '<strong>✅ 测试成功</strong><br>' +
        '获取到代理 URL：<code class="proxy-url">' + escHtml(j.proxy_url) + '</code>';
    } else {
      resEl.className = 'test-result error';
      resEl.innerHTML =
        '<strong>❌ 测试失败</strong>：' + escHtml(j.message || j.error) +
        (j.raw ? '<code class="proxy-url">' + escHtml(j.raw) + '</code>' : '');
    }
    resEl.style.display = 'block';
  } catch(e) {
    resEl.className = 'test-result error';
    resEl.innerHTML = '<strong>❌ 请求失败</strong>：' + escHtml(e.message);
    resEl.style.display = 'block';
  } finally {
    btn.disabled = false;
    btn.innerHTML = '🧪 测试连接';
  }
}

async function clearConfig() {
  if (!confirm('确认要清空并禁用动态代理配置吗？')) return;
  const btn = document.getElementById('btn-clear');
  btn.disabled = true;

  try {
    const r = await fetch(BASE + '/dynamic-proxy', {
      method: 'DELETE',
      headers: headers()
    });
    if (r.ok) {
      fillForm({});
      updateStatus({});
      showToast('🗑 配置已清空', 'info');
    } else {
      showToast('操作失败', 'error');
    }
  } catch(e) {
    showToast('网络错误: ' + e.message, 'error');
  } finally {
    btn.disabled = false;
  }
}

function escHtml(s) {
  return String(s)
    .replace(/&/g,'&amp;').replace(/</g,'&lt;')
    .replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}
</script>
</body>
</html>`

// ServeDynamicProxyPage 渲染动态代理管理页面（不需要认证，JS 层自行提示 key）。
func ServeDynamicProxyPage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Cache-Control", "no-store")
	c.String(http.StatusOK, dynamicProxyPageHTML)
}
