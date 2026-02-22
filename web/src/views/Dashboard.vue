<template>
  <div class="dashboard">
    <div class="sys-banner">
      <div class="sys-grid">
        <div class="sys-item" v-for="item in sysItems" :key="item.label">
          <div class="si-label">{{ item.label }}</div>
          <div class="si-val" :title="item.value">{{ item.value || '—' }}</div>
        </div>
      </div>
    </div>
    <div class="cards-row">
      <MetricCard :title="t('cpu')" :value="snap.cpu?.usage_percent" emoji="⚡" accent="#6366f1"
        :sub="`${snap.cpu?.cpu_threads||0} threads · Load ${snap.cpu?.load_avg_1?.toFixed(2)||'—'}`" />
      <MetricCard :title="t('memory')" :value="snap.memory?.used_percent" emoji="💾" accent="#06b6d4"
        :sub="`${fmt(snap.memory?.used)} / ${fmt(snap.memory?.total)}`" />
      <MetricCard :title="t('disk')" :value="maxDisk" emoji="💽" accent="#10b981"
        :sub="`${diskParts} ${t('partitions')}`" />
      <MetricCard :title="t('network_speed')" :value="netPct" emoji="🌐" accent="#f59e0b"
        :sub="`Up ${fmtSpeed(totalUp)} · Dn ${fmtSpeed(totalDown)}`" />
    </div>
    <div class="mid-row">
      <div class="card">
        <div class="ch"><span class="ct">{{ t('memory') }}</span></div>
        <div style="display:flex;flex-direction:column;gap:12px">
          <div class="mb-row">
            <div class="mb-hd"><span>RAM</span><span style="color:#6366f1;font-weight:700">{{ snap.memory?.used_percent?.toFixed(1) }}%</span></div>
            <div class="mb-sub">{{ fmt(snap.memory?.used) }} / {{ fmt(snap.memory?.total) }}</div>
            <div class="bar"><div class="bf" :style="`width:${snap.memory?.used_percent||0}%;background:#6366f1`"></div></div>
          </div>
          <div class="mb-row" v-if="snap.memory?.swap_total>0">
            <div class="mb-hd"><span>Swap</span><span style="color:#06b6d4;font-weight:700">{{ snap.memory?.swap_percent?.toFixed(1) }}%</span></div>
            <div class="mb-sub">{{ fmt(snap.memory?.swap_used) }} / {{ fmt(snap.memory?.swap_total) }}</div>
            <div class="bar"><div class="bf" :style="`width:${snap.memory?.swap_percent||0}%;background:#06b6d4`"></div></div>
          </div>
          <div style="display:flex;border-top:1px solid rgba(99,102,241,0.08);margin-top:4px">
            <div style="flex:1;text-align:center;padding:8px 0">
              <div style="font-size:10px;color:#9ca3af;margin-bottom:3px">{{ t('cached') }}</div>
              <div style="font-size:12px;color:#4f46e5;font-weight:600">{{ fmt(snap.memory?.cached) }}</div>
            </div>
            <div style="flex:1;text-align:center;padding:8px 0">
              <div style="font-size:10px;color:#9ca3af;margin-bottom:3px">{{ t('buffers') }}</div>
              <div style="font-size:12px;color:#4f46e5;font-weight:600">{{ fmt(snap.memory?.buffers) }}</div>
            </div>
            <div style="flex:1;text-align:center;padding:8px 0">
              <div style="font-size:10px;color:#9ca3af;margin-bottom:3px">{{ t('available') }}</div>
              <div style="font-size:12px;color:#4f46e5;font-weight:600">{{ fmt(snap.memory?.available) }}</div>
            </div>
          </div>
        </div>
      </div>
      <div class="card">
        <div class="ch"><span class="ct">CPU · {{ snap.cpu?.frequency_mhz?.toFixed(0)||'—' }} MHz</span></div>
        <div style="display:flex;flex-direction:column;gap:5px;overflow-y:auto;max-height:200px">
          <div v-for="(pct,i) in snap.cpu?.per_core_usage" :key="i" style="display:grid;grid-template-columns:28px 1fr 34px;align-items:center;gap:6px">
            <span style="font-size:11px;color:#9ca3af">C{{ i }}</span>
            <div class="bar"><div class="bf" :style="`width:${pct||0}%;background:${pct>80?'#f43f5e':pct>50?'#f59e0b':'#6366f1'}`"></div></div>
            <span style="font-size:11px;color:#4b5563;text-align:right">{{ pct?.toFixed(0) }}%</span>
          </div>
        </div>
      </div>
      <div class="card">
        <div class="ch"><span class="ct">{{ t('temperature') }}</span></div>
        <div v-if="snap.temps?.length" style="display:flex;flex-direction:column;gap:9px">
          <div v-for="tp in snap.temps" :key="tp.sensor" style="display:flex;align-items:center;gap:8px">
            <span style="font-size:12px;color:#4b5563;width:72px;flex-shrink:0;overflow:hidden;text-overflow:ellipsis">{{ tp.sensor }}</span>
            <div class="bar" style="flex:1"><div class="bf" :style="`width:${Math.min(tp.temperature/1.2,100)}%;background:${tp.temperature>80?'#f43f5e':tp.temperature>60?'#f59e0b':'#10b981'}`"></div></div>
            <span style="font-size:13px;font-weight:700;width:48px;text-align:right;font-family:monospace" :style="tp.temperature>80?'color:#f43f5e':tp.temperature>60?'color:#f59e0b':'color:#10b981'">{{ tp.temperature?.toFixed(1) }}°C</span>
          </div>
        </div>
        <div v-else style="text-align:center;padding:28px 0">
          <div style="font-size:32px">🌡️</div>
          <div style="font-size:12px;color:#9ca3af;margin-top:8px">{{ t('no_temp') }}</div>
        </div>
      </div>
    </div>
    <div class="card">
      <div class="ch"><span class="ct">{{ t('disk') }}</span></div>
      <div class="disk-grid">
        <div class="dk" v-for="p in snap.disk?.partitions" :key="p.mountpoint">
          <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:4px">
            <span style="font-size:14px;font-weight:600;color:#1e1b4b">{{ p.mountpoint }}</span>
            <span class="tag" :class="p.used_percent>85?'tag-red':p.used_percent>70?'tag-yellow':'tag-green'">{{ p.used_percent?.toFixed(1) }}%</span>
          </div>
          <div style="font-size:11px;color:#9ca3af;margin-bottom:8px">{{ p.device }} · {{ p.fstype }}</div>
          <div class="bar"><div class="bf" :style="`width:${p.used_percent||0}%;background:${p.used_percent>85?'#f43f5e':p.used_percent>70?'#f59e0b':'#10b981'}`"></div></div>
          <div style="display:flex;justify-content:space-between;font-size:11px;color:#6b7280;margin-top:6px">
            <span>{{ fmt(p.used) }} used</span><span>{{ fmt(p.free) }} free / {{ fmt(p.total) }}</span>
          </div>
        </div>
      </div>
    </div>
    <div class="card">
      <div class="ch">
        <span class="ct">{{ t('history') }}</span>
        <div style="display:flex;gap:4px">
          <button class="tab" :class="{active:histH===1}"  @click="setH(1)">1H</button>
          <button class="tab" :class="{active:histH===6}"  @click="setH(6)">6H</button>
          <button class="tab" :class="{active:histH===24}" @click="setH(24)">24H</button>
        </div>
      </div>
      <div ref="chartEl" style="height:200px"></div>
    </div>
    <div class="bot-row">
      <div class="card">
        <div class="ch"><span class="ct">{{ t('load') }}</span></div>
        <div style="display:flex">
          <div v-for="(v,k) in loads" :key="k" style="flex:1;text-align:center;padding:12px 0">
            <div style="font-size:28px;font-weight:700;font-family:monospace;background:linear-gradient(135deg,#6366f1,#06b6d4);-webkit-background-clip:text;-webkit-text-fill-color:transparent">{{ v }}</div>
            <div style="font-size:12px;color:#9ca3af;margin-top:4px">{{ k }}</div>
          </div>
        </div>
      </div>
      <div class="card">
        <div class="ch"><span class="ct">{{ t('network') }}</span><span style="font-size:11px;color:#9ca3af">{{ t('connections') }}: <b style="color:#4f46e5">{{ snap.network?.connections }}</b></span></div>
        <div style="display:flex;flex-direction:column;gap:7px">
          <div v-for="iface in snap.network?.interfaces?.slice(0,8)" :key="iface.name" style="display:flex;align-items:center;gap:8px;font-size:13px">
            <span style="font-weight:600;color:#1e1b4b;min-width:72px;overflow:hidden;text-overflow:ellipsis">{{ iface.name }}</span>
            <span style="color:#10b981;font-family:monospace;font-size:12px;flex:1">↑{{ fmtSpeed(iface.speed_up) }}</span>
            <span style="color:#6366f1;font-family:monospace;font-size:12px;flex:1">↓{{ fmtSpeed(iface.speed_down) }}</span>

          </div>
        </div>
      </div>
    </div>

    <!-- Network Monitoring Section -->
    <div class="card">
      <div class="ch">
        <span class="ct">🌐 {{ t('network_speed') }}</span>
        <div style="display:flex;align-items:center;gap:16px;font-size:12px;color:#9ca3af">
          <span>↑ <b style="color:#10b981">{{ fmtSpeed(totalUp) }}</b></span>
          <span>↓ <b style="color:#6366f1">{{ fmtSpeed(totalDown) }}</b></span>
          <span>{{ t('connections') }}: <b style="color:#4f46e5">{{ snap.network?.connections }}</b></span>
        </div>
      </div>
      <div ref="netChartEl" style="height:180px;margin-bottom:12px"></div>
      <div class="iface-grid">
        <div class="iface-card" v-for="iface in snap.network?.interfaces" :key="iface.name">
          <div style="display:flex;align-items:center;gap:6px;margin-bottom:8px">
            <div class="iface-dot" :class="(iface.speed_up||iface.speed_down)?'active':'idle'"></div>
            <span style="font-weight:700;color:#1e1b4b;font-size:13px">{{ iface.name }}</span>
            <div style="display:flex;gap:3px;flex-wrap:wrap;flex:1">
              <span class="tag tag-blue" v-for="addr in (iface.addrs||[]).slice(0,2)" :key="addr" style="font-size:9px">{{ addr }}</span>
            </div>
          </div>
          <div style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:6px;font-size:11px">
            <div style="text-align:center">
              <div style="color:#10b981;font-weight:700;font-family:monospace">{{ fmtSpeed(iface.speed_up) }}</div>
              <div style="color:#9ca3af">↑ {{ fmt(iface.bytes_sent) }}</div>
            </div>
            <div style="text-align:center">
              <div style="color:#6366f1;font-weight:700;font-family:monospace">{{ fmtSpeed(iface.speed_down) }}</div>
              <div style="color:#9ca3af">↓ {{ fmt(iface.bytes_recv) }}</div>
            </div>
            <div style="text-align:center">
              <div style="color:#f59e0b;font-weight:700;font-family:monospace">{{ (iface.packets_sent||0)+(iface.packets_recv||0) }}</div>
              <div style="color:#9ca3af">pkts</div>
            </div>
          </div>
          <div style="margin-top:7px;display:flex;flex-direction:column;gap:3px">
            <div style="display:flex;align-items:center;gap:6px">
              <span style="font-size:10px;color:#10b981;width:10px">↑</span>
              <div class="bar" style="flex:1"><div class="bf" :style="`width:${Math.min((iface.speed_up||0)/1048576*20,100)}%;background:#10b981`"></div></div>
            </div>
            <div style="display:flex;align-items:center;gap:6px">
              <span style="font-size:10px;color:#6366f1;width:10px">↓</span>
              <div class="bar" style="flex:1"><div class="bf" :style="`width:${Math.min((iface.speed_down||0)/1048576*20,100)}%;background:#6366f1`"></div></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import axios from 'axios'
import * as echarts from 'echarts'
import MetricCard from '../components/MetricCard.vue'
import { useI18n } from '../stores/i18n.js'
const i18n = useI18n(); const t = k => i18n.t(k)
const snap = ref({ cpu:{}, memory:{}, disk:{partitions:[]}, network:{interfaces:[]}, temps:[] })
const sysInfo = ref({}); const histH = ref(1); const chartEl = ref(null)
const netChartEl = ref(null)
let chart = null, netChart = null, wsh = null
const netSpeedHistory = ref([])
const sysItems = computed(() => [
  { label: t('hostname'), value: sysInfo.value.hostname },
  { label: t('os'), value: `${sysInfo.value.platform||''} ${sysInfo.value.platform_version||''}`.trim() },
  { label: t('kernel'), value: sysInfo.value.kernel_version },
  { label: t('arch'), value: sysInfo.value.arch },
  { label: t('uptime'), value: sysInfo.value.uptime_str },
  { label: 'CPU', value: sysInfo.value.cpu_model },

])
const maxDisk = computed(() => (snap.value.disk?.partitions||[]).reduce((m,p)=>Math.max(m,p.used_percent||0),0))
const diskParts = computed(() => (snap.value.disk?.partitions||[]).length)
const totalUp = computed(() => (snap.value.network?.interfaces||[]).reduce((s,i)=>s+(i.speed_up||0),0))
const totalDown = computed(() => (snap.value.network?.interfaces||[]).reduce((s,i)=>s+(i.speed_down||0),0))
const netPct = computed(() => Math.min(((totalUp.value+totalDown.value)/1048576)*10,100))
const loads = computed(() => ({'1m':snap.value.cpu?.load_avg_1?.toFixed(2)||'—','5m':snap.value.cpu?.load_avg_5?.toFixed(2)||'—','15m':snap.value.cpu?.load_avg_15?.toFixed(2)||'—'}))
function fmt(b) { if (!b) return '0 B'; const u=['B','KB','MB','GB','TB'],i=Math.min(Math.floor(Math.log(b)/Math.log(1024)),4); return (b/Math.pow(1024,i)).toFixed(1)+' '+u[i] }
function fmtSpeed(b) { if (!b) return '0 B/s'; if (b<1024) return b+'B/s'; if (b<1048576) return (b/1024).toFixed(1)+'KB/s'; return (b/1048576).toFixed(1)+'MB/s' }
function initNetChart() {
  if (!netChartEl.value) return
  netChart = echarts.init(netChartEl.value, null, {renderer:'svg'})
  netChart.setOption({
    backgroundColor:'transparent',
    tooltip:{trigger:'axis',backgroundColor:'rgba(255,255,255,0.95)',borderColor:'rgba(99,102,241,0.2)',textStyle:{color:'#1e1b4b',fontSize:12},formatter:params=>`${params[0].axisValue}<br>↑ ${params[0].data} KB/s<br>↓ ${params[1]?.data} KB/s`},
    legend:{data:['Upload','Download'],right:0,textStyle:{color:'#6b7280',fontSize:12}},
    grid:{left:60,right:16,top:28,bottom:20},
    xAxis:{type:'category',data:[],axisLabel:{color:'#9ca3af',fontSize:10},axisLine:{lineStyle:{color:'rgba(99,102,241,0.1)'}},splitLine:{show:false}},
    yAxis:{type:'value',axisLabel:{color:'#9ca3af',fontSize:11,formatter:'{value}KB/s'},splitLine:{lineStyle:{color:'rgba(99,102,241,0.06)'}}},
    series:[
      {name:'Upload',type:'line',smooth:true,data:[],showSymbol:false,lineStyle:{color:'#10b981',width:2},areaStyle:{color:{type:'linear',x:0,y:0,x2:0,y2:1,colorStops:[{offset:0,color:'rgba(16,185,129,0.18)'},{offset:1,color:'rgba(16,185,129,0)'}]}}},
      {name:'Download',type:'line',smooth:true,data:[],showSymbol:false,lineStyle:{color:'#6366f1',width:2},areaStyle:{color:{type:'linear',x:0,y:0,x2:0,y2:1,colorStops:[{offset:0,color:'rgba(99,102,241,0.15)'},{offset:1,color:'rgba(99,102,241,0)'}]}}},
    ]
  })
}
function updateNetChart() {
  if (!netChart) return
  const now = new Date(); const label = `${String(now.getHours()).padStart(2,'0')}:${String(now.getMinutes()).padStart(2,'0')}:${String(now.getSeconds()).padStart(2,'0')}`
  netSpeedHistory.value.push({ t: label, up: totalUp.value/1024, down: totalDown.value/1024 })
  if (netSpeedHistory.value.length > 60) netSpeedHistory.value.shift()
  netChart.setOption({ xAxis:{data:netSpeedHistory.value.map(p=>p.t)}, series:[{data:netSpeedHistory.value.map(p=>p.up.toFixed(1))},{data:netSpeedHistory.value.map(p=>p.down.toFixed(1))}] })
}
async function load() {
  const [sys,cpu,mem,dsk,net,tmp] = await Promise.all([axios.get('/api/system'),axios.get('/api/cpu'),axios.get('/api/memory'),axios.get('/api/disk'),axios.get('/api/network'),axios.get('/api/temperature')])
  sysInfo.value = sys.data
  snap.value = { cpu:cpu.data, memory:mem.data, disk:dsk.data, network:net.data, temps:tmp.data||[] }
}
function setH(h) { histH.value=h; loadHistory() }
async function loadHistory() {
  if (!chart) return
  const {data} = await axios.get(`/api/metrics/history?hours=${histH.value}`)
  if (!data?.length) return
  const labels = data.map(d => { const dt=new Date(d.timestamp*1000); return `${String(dt.getHours()).padStart(2,'0')}:${String(dt.getMinutes()).padStart(2,'0')}` })
  chart.setOption({ xAxis:{data:labels}, series:[{data:data.map(d=>+d.cpu.toFixed(1))},{data:data.map(d=>+d.memory.toFixed(1))}] })
}
function initChart() {
  chart = echarts.init(chartEl.value, null, {renderer:'svg'})
  chart.setOption({
    backgroundColor:'transparent',
    tooltip:{trigger:'axis',backgroundColor:'rgba(255,255,255,0.95)',borderColor:'rgba(99,102,241,0.2)',textStyle:{color:'#1e1b4b',fontSize:12}},
    legend:{data:['CPU','RAM'],right:0,textStyle:{color:'#6b7280',fontSize:12}},
    grid:{left:44,right:16,top:32,bottom:20},
    xAxis:{type:'category',data:[],axisLabel:{color:'#9ca3af',fontSize:11},axisLine:{lineStyle:{color:'rgba(99,102,241,0.1)'}},splitLine:{show:false}},
    yAxis:{type:'value',max:100,axisLabel:{color:'#9ca3af',fontSize:11,formatter:'{value}%'},splitLine:{lineStyle:{color:'rgba(99,102,241,0.06)'}}},
    series:[
      {name:'CPU',type:'line',smooth:true,data:[],showSymbol:false,lineStyle:{color:'#6366f1',width:2},areaStyle:{color:{type:'linear',x:0,y:0,x2:0,y2:1,colorStops:[{offset:0,color:'rgba(99,102,241,0.2)'},{offset:1,color:'rgba(99,102,241,0)'}]}}},
      {name:'RAM',type:'line',smooth:true,data:[],showSymbol:false,lineStyle:{color:'#06b6d4',width:2},areaStyle:{color:{type:'linear',x:0,y:0,x2:0,y2:1,colorStops:[{offset:0,color:'rgba(6,182,212,0.15)'},{offset:1,color:'rgba(6,182,212,0)'}]}}},
    ]
  })
}
onMounted(async () => {
  await load(); initChart(); loadHistory(); initNetChart()
  wsh = e => { const {event,data}=e.detail; if(event==='metrics') { snap.value={cpu:data.cpu,memory:data.memory,disk:data.disk,network:data.network,temps:data.temperatures||snap.value.temps}; updateNetChart() } }
  window.addEventListener('ws-msg', wsh)
})
onUnmounted(() => { window.removeEventListener('ws-msg', wsh); chart?.dispose(); netChart?.dispose() })
</script>
<style scoped>
.dashboard { display:flex;flex-direction:column;gap:14px; }
.sys-banner { background:linear-gradient(135deg,rgba(99,102,241,0.07),rgba(6,182,212,0.05));border:1px solid rgba(99,102,241,0.14);border-radius:14px;padding:16px 20px; }
.sys-grid { display:grid;grid-template-columns:repeat(6,1fr);gap:10px; }
.si-label { font-size:10px;color:#9ca3af;text-transform:uppercase;letter-spacing:0.06em;margin-bottom:3px; }
.si-val   { font-size:13px;color:#1e1b4b;font-weight:500;overflow:hidden;text-overflow:ellipsis;white-space:nowrap; }
.cards-row { display:grid;grid-template-columns:repeat(4,1fr);gap:14px; }
.mid-row   { display:grid;grid-template-columns:1fr 1fr 1fr;gap:14px; }
.bot-row   { display:grid;grid-template-columns:1fr 1fr;gap:14px; }
.card { background:#fff;border:1px solid rgba(99,102,241,0.1);border-radius:14px;padding:18px;box-shadow:0 2px 12px rgba(99,102,241,0.06); }
.ch { display:flex;align-items:center;justify-content:space-between;margin-bottom:14px; }
.ct { font-size:14px;font-weight:600;color:#1e1b4b; }
.bar { height:6px;background:#f0f4ff;border-radius:3px;overflow:hidden; }
.bf  { height:100%;border-radius:3px;transition:width 0.5s ease; }
.mb-row { display:flex;flex-direction:column;gap:4px; }
.mb-hd  { display:flex;justify-content:space-between;font-size:12px;color:#4b5563;font-weight:500; }
.mb-sub { font-size:11px;color:#9ca3af; }
.disk-grid { display:grid;grid-template-columns:repeat(auto-fill,minmax(200px,1fr));gap:12px; }
.dk { background:rgba(99,102,241,0.03);border:1px solid rgba(99,102,241,0.09);border-radius:10px;padding:12px; }
.tab { padding:4px 12px;border-radius:100px;font-size:12px;font-weight:500;cursor:pointer;border:1px solid rgba(99,102,241,0.2);background:transparent;color:#6b7280;transition:all 0.2s; }
.tab.active { background:linear-gradient(135deg,#6366f1,#8b5cf6);color:#fff;border-color:transparent;box-shadow:0 2px 8px rgba(99,102,241,0.3); }
@media (max-width:1100px) { .cards-row { grid-template-columns:1fr 1fr; } .mid-row { grid-template-columns:1fr 1fr; } .sys-grid { grid-template-columns:repeat(3,1fr); } }
@media (max-width:680px)  { .cards-row,.mid-row,.bot-row { grid-template-columns:1fr; } .sys-grid { grid-template-columns:1fr 1fr; } .disk-grid { grid-template-columns:1fr; } }
.iface-grid { display:grid;grid-template-columns:repeat(auto-fill,minmax(180px,1fr));gap:12px;margin-top:8px; }
.iface-card { background:rgba(99,102,241,0.03);border:1px solid rgba(99,102,241,0.09);border-radius:10px;padding:12px; }
.iface-dot  { width:8px;height:8px;border-radius:50%;flex-shrink:0; }
.iface-dot.active { background:#10b981;box-shadow:0 0 6px rgba(16,185,129,0.5);animation:pulse 2s infinite; }
.iface-dot.idle   { background:#9ca3af; }
.tag-blue { background:rgba(6,182,212,0.1);color:#0891b2; }
@keyframes pulse { 0%,100%{opacity:1}50%{opacity:.5} }
</style>
