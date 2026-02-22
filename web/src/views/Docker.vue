<template>
  <div>
    <div class="toolbar">
      <input class="inp" style="width:200px" v-model="search" :placeholder="`🔍 ${t('search')}...`" />
      <button class="btn btn-ghost btn-sm" @click="load()">🔄 {{ t('refresh') }}</button>
      <span style="margin-left:auto;font-size:12px;color:#9ca3af">{{ filtered.length }} {{ t('container') }}</span>
    </div>

    <div v-if="filtered.length" class="container-grid">
      <template v-for="(c, idx) in filtered" :key="c.id">
        <!-- 不同状态之间插入占满整行的分隔，强制换行 -->
        <div v-if="idx > 0 && filtered[idx-1].state !== c.state" class="row-break"></div>
        <div class="ccard">
        <div class="cc-head">
          <div class="cc-dot" :class="c.state==='running'?'run':c.state==='paused'?'pause':'stop'"></div>
          <div class="cc-name">{{ c.name }}</div>
          <span class="tag" :class="stateTag(c.state)">{{ c.state }}</span>
        </div>
        <div class="cc-img">🐳 {{ c.image }}</div>
        <!-- 端口区域固定两行高度，保持卡片布局整齐 -->
        <div class="cc-ports-wrap">
          <div class="cc-ports-row">
            <span
              v-if="parsedPorts(c.ports)[0]"
              class="tag tag-blue port-tag"
              @click="openPort(parsedPorts(c.ports)[0])"
              :title="isClickablePort(parsedPorts(c.ports)[0]) ? `点击打开 ${parsedPorts(c.ports)[0]}` : parsedPorts(c.ports)[0]"
              :style="isClickablePort(parsedPorts(c.ports)[0]) ? '' : 'cursor:default;opacity:0.75'"
            >{{ parsedPorts(c.ports)[0] }}</span>
          </div>
          <div class="cc-ports-row">
            <span
              v-if="parsedPorts(c.ports)[1]"
              class="tag tag-blue port-tag"
              @click="openPort(parsedPorts(c.ports)[1])"
              :title="isClickablePort(parsedPorts(c.ports)[1]) ? `点击打开 ${parsedPorts(c.ports)[1]}` : parsedPorts(c.ports)[1]"
              :style="isClickablePort(parsedPorts(c.ports)[1]) ? '' : 'cursor:default;opacity:0.75'"
            >{{ parsedPorts(c.ports)[1] }}</span>
          </div>
        </div>
        <div class="cc-metrics" v-if="c.state==='running'">
          <div class="cm">
            <span style="font-size:10px;color:#9ca3af">CPU</span>
            <div class="mini-bar"><div class="mini-fill" :style="`width:${c.cpu_percent||0}%;background:#6366f1`"></div></div>
            <span style="font-size:11px;color:#6366f1;font-weight:600">{{ c.cpu_percent?.toFixed(1) }}%</span>
          </div>
          <div class="cm">
            <span style="font-size:10px;color:#9ca3af">MEM</span>
            <div class="mini-bar"><div class="mini-fill" :style="`width:${c.mem_percent||0}%;background:#06b6d4`"></div></div>
            <span style="font-size:11px;color:#06b6d4;font-weight:600">{{ c.mem_percent?.toFixed(1) }}%</span>
          </div>
        </div>
        <div class="cc-actions">
          <button class="btn btn-icon btn-cyan"  v-if="c.state!=='running'" @click="action(c,'start')" title="启动">▶</button>
          <button class="btn btn-icon btn-ghost" v-if="c.state==='running'" @click="action(c,'stop')" title="停止">⏹</button>
          <button class="btn btn-icon btn-ghost" @click="action(c,'restart')" title="重启">↺</button>
          <button class="btn btn-icon btn-ghost" @click="showInspect(c)" title="参数">⚙️</button>
          <button class="btn btn-icon btn-ghost" @click="pullUpdate(c)" :disabled="updating===c.id" title="更新镜像">{{ updating===c.id ? '⏳' : '⬆️' }}</button>
          <button class="btn btn-icon btn-ghost" @click="showLogs(c)" title="日志">📋</button>
        </div>
        </div>
      </template>
    </div>

    <div class="card empty-state" v-else>
      <div style="font-size:48px;margin-bottom:16px">🐳</div>
      <div style="font-size:16px;font-weight:600;color:#1e1b4b;margin-bottom:6px">{{ t('no_docker') }}</div>
      <div style="font-size:13px;color:#9ca3af">Make sure Docker is running</div>
    </div>

    <!-- Logs modal -->
    <div class="modal-overlay" v-if="logModal" @click.self="logModal=null">
      <div class="modal" style="width:720px;max-width:96vw">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:16px">
          <h3 style="color:#1e1b4b;font-size:16px">📋 {{ logModal.name }} {{ t('logs') }}</h3>
          <button class="btn btn-ghost btn-sm" @click="logModal=null">✕ {{ t('close') }}</button>
        </div>
        <pre class="log-box">{{ logContent || t('loading') }}</pre>
      </div>
    </div>

    <!-- 容器参数 + Compose 编辑 modal -->
    <div class="modal-overlay" v-if="inspectModal" @click.self="closeInspect">
      <div class="modal" style="width:760px;max-width:96vw">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:16px">
          <div>
            <h3 style="color:#1e1b4b;font-size:15px;font-weight:700">⚙️ {{ inspectModal.name }} 参数</h3>
            <p v-if="inspectData && inspectData.compose_file" style="font-size:11px;color:#10b981;margin-top:2px">📦 docker-compose 容器</p>
          </div>
          <button class="btn btn-ghost btn-sm" @click="closeInspect">✕</button>
        </div>

        <div class="tab-bar" v-if="inspectData && inspectData.compose_file">
          <button class="tab-btn" :class="{active: inspectTab==='info'}" @click="inspectTab='info'">容器信息</button>
          <button class="tab-btn" :class="{active: inspectTab==='compose'}" @click="inspectTab='compose'">编辑 Compose</button>
        </div>

        <div v-if="inspectTab==='info' || !(inspectData && inspectData.compose_file)">
          <div v-if="!inspectData" style="text-align:center;padding:30px;color:#9ca3af">加载中...</div>
          <div v-else class="inspect-grid">
            <div class="inspect-item"><span class="inspect-label">镜像</span><span class="inspect-val mono">{{ inspectData.image }}</span></div>
            <div class="inspect-item"><span class="inspect-label">状态</span><span class="inspect-val">{{ inspectData.status }}</span></div>
            <div class="inspect-item"><span class="inspect-label">创建时间</span><span class="inspect-val mono">{{ inspectData.created }}</span></div>
            <div class="inspect-item"><span class="inspect-label">重启策略</span><span class="inspect-val">{{ inspectData.restart_policy }}</span></div>
            <div class="inspect-item" style="grid-column:1/-1"><span class="inspect-label">环境变量</span><pre class="mini-pre">{{ (inspectData.env||[]).join('\n') || '—' }}</pre></div>
            <div class="inspect-item" style="grid-column:1/-1"><span class="inspect-label">挂载卷</span><pre class="mini-pre">{{ (inspectData.mounts||[]).join('\n') || '—' }}</pre></div>
            <div class="inspect-item" style="grid-column:1/-1"><span class="inspect-label">网络</span><span class="inspect-val">{{ (inspectData.networks||[]).join(', ') || '—' }}</span></div>
            <div class="inspect-item" style="grid-column:1/-1"><span class="inspect-label">端口映射</span><span class="inspect-val mono">{{ inspectData.ports || '—' }}</span></div>
            <div class="inspect-item" style="grid-column:1/-1"><span class="inspect-label">启动命令</span><pre class="mini-pre">{{ (inspectData.cmd||[]).join(' ') || '—' }}</pre></div>
          </div>
        </div>

        <div v-if="inspectTab==='compose' && inspectData && inspectData.compose_file">
          <div style="font-size:11px;color:#9ca3af;font-family:monospace;margin-bottom:8px">{{ inspectData.compose_file }}</div>
          <div v-if="composeErr" class="alert-err">{{ composeErr }}</div>
          <div v-if="composeOk" class="alert-ok">✓ {{ composeOk }}</div>
          <textarea class="editor-box" v-model="composeContent" spellcheck="false" placeholder="加载中..."></textarea>
          <div style="display:flex;gap:8px;margin-top:12px;align-items:center;flex-wrap:wrap">
            <button class="btn btn-danger btn-sm" @click="applyCompose" :disabled="composeSaving">{{ composeSaving ? '处理中...' : '💥 销毁并重建容器' }}</button>
            <span style="font-size:11px;color:#9ca3af">⚠️ 将销毁当前容器并用新 compose 文件重建</span>
          </div>
        </div>

        <div style="margin-top:16px;border-top:1px solid rgba(99,102,241,0.1);padding-top:12px">
          <button class="btn btn-ghost btn-sm" @click="pullUpdate(inspectModal)" :disabled="updating===inspectModal.id">
            {{ updating===inspectModal.id ? '⏳ 更新中...' : '⬆️ 一键更新镜像' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 更新结果 modal -->
    <div class="modal-overlay" v-if="updateLog" @click.self="updateLog=null">
      <div class="modal" style="width:640px;max-width:96vw">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:16px">
          <h3 style="color:#1e1b4b;font-size:15px">⬆️ 镜像更新结果</h3>
          <button class="btn btn-ghost btn-sm" @click="updateLog=null">✕</button>
        </div>
        <pre class="log-box">{{ updateLog }}</pre>
        <button class="btn btn-ghost btn-sm" style="margin-top:12px" @click="updateLog=null;load(true)">关闭并刷新</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from 'axios'
import { useI18n } from '../stores/i18n.js'
const i18n = useI18n(); const t = k => i18n.t(k)

const containers = ref([])
const search = ref('')
const logModal = ref(null)
const logContent = ref('')
const inspectModal = ref(null)
const inspectData = ref(null)
const inspectTab = ref('info')
const composeContent = ref('')
const composeErr = ref('')
const composeOk = ref('')
const composeSaving = ref(false)
const updating = ref(null)
const updateLog = ref(null)

const filtered = computed(() => {
  const stateOrder = s => s === 'running' ? 0 : s === 'exited' ? 1 : 2
  return containers.value
    .filter(c =>
      c.name?.toLowerCase().includes(search.value.toLowerCase()) ||
      c.image?.toLowerCase().includes(search.value.toLowerCase())
    )
    .sort((a, b) => stateOrder(a.state) - stateOrder(b.state))
})

const groupedContainers = computed(() => {
  const order = ['running', 'exited']
  const groups = {}
  for (const c of filtered.value) {
    const key = order.includes(c.state) ? c.state : 'other'
    if (!groups[key]) groups[key] = []
    groups[key].push(c)
  }
  const result = []
  for (const state of [...order, 'other']) {
    if (groups[state]?.length) result.push({ state, items: groups[state] })
  }
  return result
})

function stateTag(s) { return s==='running'?'tag tag-green':s==='paused'?'tag tag-yellow':'tag tag-gray' }

async function load() {
  try {
    const { data } = await axios.get('/api/docker/containers')
    containers.value = data || []
  } catch { if (!containers.value.length) containers.value = [] }
}

async function action(c, act) {
  try { await axios.post(`/api/docker/containers/${c.id}/${act}`); setTimeout(() => load(), 1000) }
  catch(e) { alert(e.response?.data?.error || e.message) }
}

async function showLogs(c) {
  logModal.value = c; logContent.value = ''
  const { data } = await axios.get(`/api/docker/containers/${c.id}/logs`)
  logContent.value = data.logs || ''
}

// Parse raw docker ports string into individual entries, keeping both IPv4 and IPv6
// Properly splits concatenated strings like "0.0.0.0:4444->4444/tcp[::]:4444->4444/tcp"
// Result row0 = IPv4 mapping, row1 = IPv6 mapping (or exposed-only port)
function parsedPorts(portsStr) {
  if (!portsStr) return []
  let s = portsStr
  // Insert comma before [ that follows /tcp /udp /sctp
  s = s.replace(/(\/(?:tcp|udp|sctp))(\[)/g, '$1,$2')
  // Insert comma before a digit that follows /tcp /udp /sctp (handles "443/tcp0.0.0.0:...")
  s = s.replace(/(\/(?:tcp|udp|sctp))(\d)/g, '$1,$2')
  // Split on comma and whitespace
  const tokens = s.split(/[,\s]+/).map(t => t.trim()).filter(Boolean)
  // Deduplicate while preserving order
  const seen = new Set()
  const result = []
  for (const tok of tokens) {
    if (tok && !seen.has(tok)) { seen.add(tok); result.push(tok) }
  }
  return result
}

function isClickablePort(portStr) {
  // Clickable if it has a host port mapping (IPv4 or IPv6)
  return /(?:^[0-9.]+:|^\[::]:)\d+->/.test(portStr)
}

function openPort(portStr) {
  // Extract host port from IPv4 binding like 0.0.0.0:4444->...
  const m = portStr.match(/^[0-9.]+:(\d+)->/)
  if (m) { window.open(`http://${window.location.hostname}:${m[1]}`, '_blank') }
}

async function showInspect(c) {
  inspectModal.value = c
  inspectData.value = null
  inspectTab.value = 'info'
  composeContent.value = ''
  composeErr.value = ''
  composeOk.value = ''
  try {
    const { data } = await axios.get(`/api/docker/containers/${c.id}/inspect`)
    inspectData.value = data
    if (data.compose_file) {
      try {
        const r2 = await axios.get('/api/docker/compose/file', { params: { path: data.compose_file } })
        composeContent.value = r2.data.content || ''
      } catch { composeContent.value = '' }
    }
  } catch(e) { inspectData.value = { image: '获取失败', status: e.response?.data?.error || e.message, env:[], mounts:[], networks:[], cmd:[] } }
}

function closeInspect() { inspectModal.value = null; inspectData.value = null }

async function applyCompose() {
  if (!confirm('确定要销毁当前容器并用修改后的 compose 文件重建吗？')) return
  composeSaving.value = true; composeErr.value = ''; composeOk.value = ''
  try {
    const { data } = await axios.post('/api/docker/compose/apply', {
      path: inspectData.value.compose_file,
      content: composeContent.value,
      container_id: inspectModal.value.id
    })
    composeOk.value = data.message || '重建成功'
    setTimeout(() => load(true), 2000)
  } catch(e) { composeErr.value = e.response?.data?.error || e.message }
  finally { composeSaving.value = false }
}

async function pullUpdate(c) {
  if (!c) return
  if (!confirm(`确定要拉取最新镜像并重建容器 ${c.name} 吗？`)) return
  updating.value = c.id
  try {
    const { data } = await axios.post(`/api/docker/containers/${c.id}/update`)
    updateLog.value = data.log || '更新完成'
    load(true)
  } catch(e) { updateLog.value = '错误: ' + (e.response?.data?.error || e.message) }
  finally { updating.value = null }
}

onMounted(load)
</script>

<style scoped>
.toolbar { display:flex;align-items:center;gap:8px;flex-wrap:wrap;margin-bottom:14px; }
.inp { background:#f8faff;border:1.5px solid rgba(99,102,241,0.15);color:#1e1b4b;border-radius:8px;padding:8px 12px;font-size:13px;font-family:inherit;outline:none; }
.inp:focus { border-color:#6366f1; }
.container-grid { display:flex;flex-wrap:wrap;gap:12px;align-content:flex-start; }
.ccard { flex:0 0 280px;background:#fff;border:1px solid rgba(99,102,241,0.1);border-radius:14px;padding:14px;box-shadow:0 2px 12px rgba(99,102,241,0.06);transition:transform 0.2s;box-sizing:border-box; }
.ccard:hover { transform:translateY(-2px); }
.cc-head { display:flex;align-items:center;gap:8px;margin-bottom:10px; }
.cc-dot { width:9px;height:9px;border-radius:50%;flex-shrink:0; }
.cc-dot.run { background:#10b981;box-shadow:0 0 8px rgba(16,185,129,0.5);animation:pulse 2s infinite; }
.cc-dot.pause { background:#f59e0b; }
.cc-dot.stop { background:#9ca3af; }
.cc-name { font-size:15px;font-weight:700;color:#1e1b4b;flex:1;overflow:hidden;text-overflow:ellipsis; }
.cc-img { font-size:12px;color:#9ca3af;margin-bottom:8px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap; }
.cc-ports-wrap { margin-bottom:8px; }
.cc-ports-row { height:22px;display:flex;align-items:center;margin-bottom:2px; }
.cc-ports-row:last-child { margin-bottom:0; }
.port-tag { cursor:pointer;transition:all 0.15s; }
.port-tag:hover { background:rgba(6,182,212,0.2);color:#0e7490;transform:scale(1.05); }
.cc-metrics { display:flex;flex-direction:column;gap:6px;margin-bottom:12px;background:rgba(99,102,241,0.04);border-radius:8px;padding:10px; }
.cm { display:flex;align-items:center;gap:6px; }
.mini-bar { flex:1;height:4px;background:#f0f4ff;border-radius:2px;overflow:hidden; }
.mini-fill { height:100%;border-radius:2px;transition:width 0.5s; }
.cc-actions { display:flex;gap:5px;align-items:center;flex-wrap:nowrap; }
.card { background:#fff;border:1px solid rgba(99,102,241,0.1);border-radius:14px;padding:18px;box-shadow:0 2px 12px rgba(99,102,241,0.06); }
.empty-state { text-align:center;padding:60px 20px; }
.modal-overlay { position:fixed;inset:0;background:rgba(30,27,75,0.4);backdrop-filter:blur(6px);display:flex;align-items:center;justify-content:center;z-index:1000; }
.modal { background:#fff;border:1px solid rgba(99,102,241,0.15);border-radius:16px;padding:24px;box-shadow:0 20px 60px rgba(99,102,241,0.15);max-height:90vh;overflow-y:auto; }
.log-box { background:#0f172a;color:#e2e8f0;border-radius:10px;padding:16px;font-family:'JetBrains Mono',monospace;font-size:12px;max-height:50vh;overflow-y:auto;white-space:pre-wrap;word-break:break-all; }
.btn { display:inline-flex;align-items:center;gap:4px;padding:7px 14px;border-radius:8px;font-size:13px;font-weight:500;cursor:pointer;border:none;font-family:inherit;transition:all 0.2s; }
.btn:disabled { opacity:0.5;cursor:not-allowed; }
.btn-sm { padding:5px 11px;font-size:12px; }
.btn-icon { padding:5px 8px;font-size:13px;border-radius:8px; }
.btn-cyan { background:linear-gradient(135deg,#06b6d4,#10b981);color:#fff;box-shadow:0 2px 8px rgba(6,182,212,0.3); }
.btn-ghost { background:#fff;color:#6b7280;border:1px solid rgba(99,102,241,0.15); }
.btn-ghost:hover { background:rgba(99,102,241,0.06); }
.btn-danger { background:rgba(244,63,94,0.1);color:#f43f5e;border:1px solid rgba(244,63,94,0.2); }
.tag { display:inline-flex;align-items:center;padding:2px 8px;border-radius:100px;font-size:11px;font-weight:600; }
.tag-green { background:rgba(16,185,129,0.1);color:#10b981; }
.tag-yellow { background:rgba(245,158,11,0.1);color:#f59e0b; }
.tag-gray { background:rgba(107,114,128,0.1);color:#6b7280; }
.tag-blue { background:rgba(99,102,241,0.1);color:#6366f1; }
.tab-bar { display:flex;gap:4px;margin-bottom:14px;border-bottom:1px solid rgba(99,102,241,0.12);padding-bottom:8px; }
.tab-btn { padding:5px 14px;border-radius:8px 8px 0 0;font-size:13px;font-weight:500;cursor:pointer;border:none;background:transparent;color:#9ca3af;transition:all 0.2s; }
.tab-btn.active { background:rgba(99,102,241,0.1);color:#6366f1;font-weight:600; }
.inspect-grid { display:grid;grid-template-columns:1fr 1fr;gap:10px; }
.inspect-item { background:#f8faff;border:1px solid rgba(99,102,241,0.08);border-radius:10px;padding:10px 14px;display:flex;flex-direction:column;gap:4px; }
.inspect-label { font-size:10px;font-weight:700;text-transform:uppercase;letter-spacing:0.06em;color:#9ca3af; }
.inspect-val { font-size:12px;color:#1e1b4b;font-weight:500;word-break:break-all; }
.mono { font-family:'JetBrains Mono',monospace; }
.mini-pre { margin:4px 0 0;font-family:'JetBrains Mono',monospace;font-size:11px;color:#374151;background:#f0f4ff;border-radius:6px;padding:6px 8px;max-height:80px;overflow-y:auto;white-space:pre-wrap;word-break:break-all; }
.editor-box { width:100%;height:45vh;background:#0f172a;color:#e2e8f0;border:1px solid rgba(99,102,241,0.2);border-radius:10px;padding:14px;font-family:'JetBrains Mono',monospace;font-size:12px;line-height:1.6;resize:vertical;outline:none;box-sizing:border-box; }
.alert-err { background:rgba(244,63,94,0.1);border:1px solid rgba(244,63,94,0.3);border-radius:8px;padding:10px;margin-bottom:10px;font-size:12px;color:#f43f5e; }
.alert-ok { background:rgba(16,185,129,0.1);border:1px solid rgba(16,185,129,0.3);border-radius:8px;padding:10px;margin-bottom:10px;font-size:12px;color:#10b981; }
@keyframes pulse { 0%,100%{opacity:1}50%{opacity:.5} }
.row-break { flex-basis:100%;width:100%;height:0;margin:0; }
@media (max-width:600px) {
  .ccard { flex:0 0 calc(50% - 6px); }
  .inspect-grid { grid-template-columns:1fr; }
}
@media (max-width:400px) {
  .ccard { flex:0 0 100%; }
}
</style>
