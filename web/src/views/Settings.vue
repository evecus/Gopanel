<template>
  <div class="settings-page">
    <div class="settings-grid">
      <!-- Change credentials -->
      <div class="card">
        <div class="sh">
          <div class="si">🔐</div>
          <div>
            <div class="st">{{ t('credentials') }}</div>
            <div class="sd">{{ t('credentials_desc') }}</div>
          </div>
        </div>
        <form @submit.prevent="saveCredentials" style="display:flex;flex-direction:column;gap:14px">
          <div class="field">
            <label>{{ t('current_username') }}</label>
            <input class="input" v-model="creds.username" required />
          </div>
          <div class="field">
            <label>{{ t('current_password') }}</label>
            <input class="input" type="password" v-model="creds.password" required />
          </div>
          <div style="height:1px;background:rgba(99,102,241,0.1)"></div>
          <div class="field">
            <label>{{ t('new_username') }} <span style="color:#9ca3af">({{ t('optional') }})</span></label>
            <input class="input" v-model="creds.new_username" :placeholder="t('leave_blank')" />
          </div>
          <div class="field">
            <label>{{ t('new_password') }} <span style="color:#9ca3af">({{ t('optional') }})</span></label>
            <input class="input" type="password" v-model="creds.new_password" :placeholder="t('leave_blank')" />
          </div>
          <div class="alert alert-success" v-if="credMsg==='ok'">✅ {{ t('save_success') }}</div>
          <div class="alert alert-error"   v-if="credMsg==='err'">❌ {{ t('save_fail') }}: {{ credErr }}</div>
          <button class="btn btn-primary" type="submit" :disabled="credLoading">
            <span v-if="credLoading" class="animate-spin" style="display:inline-block">⟳</span>
            {{ t('save') }}
          </button>
        </form>
      </div>

      <!-- Language & Display -->
      <div class="card">
        <div class="sh">
          <div class="si">🌐</div>
          <div>
            <div class="st">{{ t('display') }}</div>
            <div class="sd">{{ t('display_desc') }}</div>
          </div>
        </div>
        <div style="display:flex;flex-direction:column;gap:16px">
          <div class="field">
            <label>{{ t('language') }}</label>
            <div style="display:flex;gap:8px;margin-top:6px">
              <button class="lang-opt" :class="{active:i18n.locale==='zh'}" @click="i18n.locale='zh';saveLang()">🇨🇳 中文</button>
              <button class="lang-opt" :class="{active:i18n.locale==='en'}" @click="i18n.locale='en';saveLang()">🇺🇸 English</button>
            </div>
          </div>
          <div class="info-box">
            <div class="ib-row"><span>{{ t('version') }}</span><strong>GoPanel v1.0</strong></div>
            <div class="ib-row"><span>Go + Vue 3</span><strong>MIT License</strong></div>
            <div class="ib-row"><span>Port</span><strong>1080</strong></div>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>
<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from '../stores/i18n.js'
const i18n = useI18n(); const t = k => i18n.t(k)
const creds = ref({ username:'', password:'', new_username:'', new_password:'' })
const credMsg = ref(''), credErr = ref(''), credLoading = ref(false)
async function saveCredentials() {
  credMsg.value=''; credLoading.value=true
  try {
    await axios.post('/api/settings/credentials', creds.value)
    credMsg.value='ok'
    creds.value = { username:'', password:'', new_username:'', new_password:'' }
  } catch(e) {
    credMsg.value='err'; credErr.value = e.response?.data?.error || e.message
  } finally { credLoading.value=false }
}
function saveLang() { localStorage.setItem('gp_lang', i18n.locale.value) }
onMounted(() => {})
</script>
<style scoped>
.settings-page { }
.settings-grid { display:grid;grid-template-columns:1fr 1fr;gap:14px; }
.card { background:#fff;border:1px solid rgba(99,102,241,0.1);border-radius:14px;padding:24px;box-shadow:0 2px 12px rgba(99,102,241,0.06); }
.sh { display:flex;align-items:flex-start;gap:14px;margin-bottom:20px; }
.si { width:44px;height:44px;background:linear-gradient(135deg,rgba(99,102,241,0.12),rgba(6,182,212,0.08));border:1px solid rgba(99,102,241,0.15);border-radius:12px;display:flex;align-items:center;justify-content:center;font-size:20px;flex-shrink:0; }
.st { font-size:16px;font-weight:700;color:#1e1b4b; }
.sd { font-size:13px;color:#9ca3af;margin-top:3px; }
.field { display:flex;flex-direction:column;gap:6px; }
.field label { font-size:13px;font-weight:500;color:#4f46e5; }
.input { background:#f8faff;border:1.5px solid rgba(99,102,241,0.15);color:#1e1b4b;border-radius:8px;padding:9px 13px;font-size:13px;font-family:inherit;outline:none;transition:border-color 0.2s,box-shadow 0.2s;width:100%; }
.input:focus { border-color:#6366f1;box-shadow:0 0 0 3px rgba(99,102,241,0.1); }
.alert { padding:10px 14px;border-radius:8px;font-size:13px; }
.alert-success { background:rgba(16,185,129,0.08);border:1px solid rgba(16,185,129,0.2);color:#059669; }
.alert-error   { background:rgba(244,63,94,0.08);border:1px solid rgba(244,63,94,0.2);color:#f43f5e; }
.btn { display:inline-flex;align-items:center;justify-content:center;gap:6px;padding:10px 20px;border-radius:8px;font-size:14px;font-weight:600;font-family:inherit;cursor:pointer;border:none;transition:all 0.2s; }
.btn-primary { background:linear-gradient(135deg,#6366f1,#8b5cf6);color:#fff;box-shadow:0 4px 12px rgba(99,102,241,0.3); }
.btn-primary:hover { transform:translateY(-1px);box-shadow:0 6px 18px rgba(99,102,241,0.4); }
.lang-opt { padding:8px 20px;border-radius:100px;font-size:13px;font-weight:500;cursor:pointer;border:1.5px solid rgba(99,102,241,0.2);background:transparent;color:#6b7280;transition:all 0.2s;font-family:inherit; }
.lang-opt.active { background:linear-gradient(135deg,#6366f1,#8b5cf6);color:#fff;border-color:transparent;box-shadow:0 3px 10px rgba(99,102,241,0.3); }
.info-box { background:rgba(99,102,241,0.04);border:1px solid rgba(99,102,241,0.1);border-radius:10px;padding:14px;display:flex;flex-direction:column;gap:10px; }
.ib-row { display:flex;justify-content:space-between;font-size:13px;color:#6b7280; }
.ib-row strong { color:#1e1b4b; }
.info-grid { display:grid;grid-template-columns:repeat(auto-fill,minmax(220px,1fr));gap:12px; }
.ig-item { background:rgba(99,102,241,0.04);border:1px solid rgba(99,102,241,0.08);border-radius:10px;padding:12px; }
.ig-lbl { font-size:11px;color:#9ca3af;text-transform:uppercase;letter-spacing:0.05em;margin-bottom:4px; }
.ig-val { font-size:13px;color:#1e1b4b;font-weight:500;overflow:hidden;text-overflow:ellipsis; }
@media (max-width:768px) { .settings-grid { grid-template-columns:1fr; } }
</style>
