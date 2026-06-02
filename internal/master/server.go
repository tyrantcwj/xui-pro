package master

import (
	"encoding/json"
	"net/http"
	"strings"

	"xui-next/internal/domain"
	"xui-next/internal/reality"
	"xui-next/internal/version"
	"xui-next/internal/xray"
)

type Server struct {
	store   *Store
	reality *reality.Library
	xray    *xray.Manager
}

func NewServer(store *Store, reality *reality.Library, xrayManager *xray.Manager) *Server {
	return &Server{store: store, reality: reality, xray: xrayManager}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.index)
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/version", s.version)
	mux.HandleFunc("GET /api/nodes", s.nodes)
	mux.HandleFunc("POST /api/nodes/register", s.registerNode)
	mux.HandleFunc("POST /api/nodes/", s.nodeAction)
	mux.HandleFunc("GET /api/inbounds", s.inbounds)
	mux.HandleFunc("POST /api/inbounds", s.saveInbound)
	mux.HandleFunc("DELETE /api/inbounds/", s.deleteInbound)
	mux.HandleFunc("GET /api/xray/status", s.xrayStatus)
	mux.HandleFunc("GET /api/xray/config", s.xrayConfig)
	mux.HandleFunc("POST /api/xray/apply", s.xrayApply)
	mux.HandleFunc("GET /api/reality/domains", s.realityDomains)
	mux.HandleFunc("POST /api/reality/recommend", s.realityRecommend)
	return mux
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": version.String()})
}

func (s *Server) nodes(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Nodes())
}

func (s *Server) inbounds(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Inbounds())
}

func (s *Server) saveInbound(w http.ResponseWriter, r *http.Request) {
	var inbound domain.Inbound
	if !decodeJSON(w, r, &inbound) {
		return
	}
	if inbound.Port <= 0 {
		http.Error(w, "port is required", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, s.store.UpsertInbound(inbound))
}

func (s *Server) deleteInbound(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/inbounds/")
	if id == "" || !s.store.DeleteInbound(id) {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (s *Server) xrayStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.xray.Status())
}

func (s *Server) xrayConfig(w http.ResponseWriter, _ *http.Request) {
	data, err := s.xray.Render(s.store.Inbounds())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(data)
}

func (s *Server) xrayApply(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.xray.Apply(s.store.Inbounds()))
}

func (s *Server) registerNode(w http.ResponseWriter, r *http.Request) {
	var n domain.Node
	if !decodeJSON(w, r, &n) {
		return
	}
	if n.ID == "" {
		http.Error(w, "node id is required", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, s.store.UpsertNode(n))
}

func (s *Server) nodeAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/nodes/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	nodeID, action := parts[0], parts[1]
	switch action {
	case "heartbeat":
		var req domain.HeartbeatRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		req.Node.ID = nodeID
		writeJSON(w, http.StatusOK, s.store.SaveHeartbeat(req))
	case "desired-config":
		cfg, err := s.store.DesiredConfig(nodeID)
		if err != nil {
			writeJSON(w, http.StatusOK, domain.DesiredConfig{Version: "empty"})
			return
		}
		writeJSON(w, http.StatusOK, cfg)
	case "update":
		var patch domain.Node
		if !decodeJSON(w, r, &patch) {
			return
		}
		node, ok := s.store.UpdateNode(nodeID, patch)
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, node)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) realityDomains(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.reality.Domains(r.URL.Query().Get("region")))
}

func (s *Server) realityRecommend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Region string `json:"region"`
		Limit  int    `json:"limit"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	writeJSON(w, http.StatusOK, s.reality.Recommend(req.Region, req.Limit))
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

const indexHTML = `<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>XUI Pro</title>
<style>:root{color-scheme:dark;font-family:Inter,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;background:#101316;color:#edf5f2}body{margin:0}.shell{max-width:1180px;margin:0 auto;padding:40px 20px}.top{display:flex;justify-content:space-between;gap:16px;align-items:center;margin-bottom:24px}h1{margin:0;font-size:30px}h2{margin:0 0 16px;font-size:24px}p{color:#94a6a1}.badge{border-radius:999px;padding:6px 10px;background:#18382c;color:#65e6ad;font-weight:700}.grid{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:14px}.card,.panel{border:1px solid #29343c;background:#171d22;border-radius:8px;padding:18px}.panel{margin-top:16px;overflow-x:auto}.card span{display:block;color:#94a6a1;margin-bottom:8px}.card strong{font-size:24px;word-break:break-word}table{width:100%;border-collapse:collapse}th,td{text-align:left;padding:10px 8px;border-bottom:1px solid #27323a;white-space:nowrap}th{color:#94a6a1;font-size:13px}a,.linkbtn{color:#65e6ad}.linkbtn{border:0;background:transparent;cursor:pointer;font:inherit;padding:0}.actions{display:flex;gap:12px;align-items:center}.toolbar{display:flex;gap:12px;margin-bottom:12px}.primary{border:1px solid #65e6ad;background:#18382c;color:#65e6ad;border-radius:6px;padding:8px 12px;cursor:pointer;text-decoration:none}code{background:#222b31;padding:2px 6px;border-radius:6px}@media(max-width:900px){.grid{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(max-width:600px){.top{align-items:flex-start;flex-direction:column}.grid{grid-template-columns:1fr}.toolbar{flex-wrap:wrap}}</style></head>
<body><main class="shell"><div class="top"><div><h1>XUI Pro</h1><p>&#20027;&#25511;&#26381;&#21153;&#24050;&#21551;&#21160;&#12290;&#20837;&#31449;&#20250;&#29983;&#25104; Xray &#37197;&#32622;&#24182;&#21487;&#24212;&#29992;&#21040;&#26381;&#21153;&#12290;</p></div><span class="badge" id="health">&#26816;&#27979;&#20013;</span></div>
<section class="grid"><div class="card"><span>&#33410;&#28857;&#25968;&#37327;</span><strong id="nodeCount">-</strong></div><div class="card"><span>&#22312;&#32447;&#33410;&#28857;</span><strong id="onlineCount">-</strong></div><div class="card"><span>&#20027;&#25511;&#29256;&#26412;</span><strong id="masterVersion">-</strong></div><div class="card"><span>Xray</span><strong id="xrayStatus">-</strong></div></section>
<section class="panel"><h2>&#33410;&#28857;&#21015;&#34920;</h2><table><thead><tr><th>&#33410;&#28857;&#21517;</th><th>&#22269;&#23478;</th><th>&#29366;&#24577;</th><th>CPU</th><th>&#20869;&#23384;</th><th>&#30913;&#30424;</th><th>Agent &#29256;&#26412;</th><th>&#26368;&#21518;&#24515;&#36339;</th><th>&#25805;&#20316;</th></tr></thead><tbody id="nodes"><tr><td colspan="9">&#26242;&#26080;&#33410;&#28857;&#12290;</td></tr></tbody></table></section>
<section class="panel"><div class="toolbar"><h2 style="flex:1">&#20837;&#31449;&#31649;&#29702;</h2><button class="primary" onclick="addInbound()">&#26032;&#22686;&#20837;&#31449;</button><button class="primary" onclick="applyXray()">&#24212;&#29992;&#21040; Xray</button><a class="primary" href="/api/xray/config" target="_blank">&#39044;&#35272;&#37197;&#32622;</a></div><table><thead><tr><th>&#22791;&#27880;</th><th>&#21327;&#35758;</th><th>&#31471;&#21475;</th><th>&#21551;&#29992;</th><th>&#24635;&#27969;&#37327;</th><th>&#36807;&#26399;</th><th>Tag</th><th>&#25805;&#20316;</th></tr></thead><tbody id="inbounds"><tr><td colspan="8">&#26242;&#26080;&#20837;&#31449;&#12290;</td></tr></tbody></table></section>
<section class="panel"><h2>&#25509;&#21475;</h2><p><a href="/api/health">/api/health</a> <a href="/api/nodes">/api/nodes</a> <a href="/api/inbounds">/api/inbounds</a> <a href="/api/xray/status">/api/xray/status</a></p></section></main>
<script>
async function load(){const h=await fetch('/api/health').then(r=>r.json()).catch(()=>({status:'error'}));document.getElementById('health').textContent=h.status==='ok'?'\u6b63\u5e38':'\u5f02\u5e38';const v=await fetch('/api/version').then(r=>r.json()).catch(()=>({version:'-'}));document.getElementById('masterVersion').textContent=v.version||'-';const xs=await fetch('/api/xray/status').then(r=>r.json()).catch(()=>({running:false}));document.getElementById('xrayStatus').textContent=xs.running?'\u8fd0\u884c':'\u672a\u8fd0\u884c';const ns=await fetch('/api/nodes').then(r=>r.json()).catch(()=>[]);document.getElementById('nodeCount').textContent=ns.length;document.getElementById('onlineCount').textContent=ns.filter(n=>n.status==='online').length;document.getElementById('nodes').innerHTML=ns.length?ns.map(nodeRow).join(''):'<tr><td colspan="9">&#26242;&#26080;&#33410;&#28857;&#12290;</td></tr>';const ins=await fetch('/api/inbounds').then(r=>r.json()).catch(()=>[]);document.getElementById('inbounds').innerHTML=ins.length?ins.map(inboundRow).join(''):'<tr><td colspan="8">&#26242;&#26080;&#20837;&#31449;&#12290;</td></tr>'}
function nodeRow(n){const m=n.metrics||{};const target=n.endpoint||n.publicIp||n.name||n.id;const user=n.sshUser||'root';const ssh='ssh '+user+'@'+target;const safe=esc(n.id);return '<tr><td>'+html(n.name||n.id)+'</td><td>'+html(countryText(n.country||n.region))+'</td><td>'+statusText(n.status)+'</td><td>'+pct(m.cpu)+'</td><td>'+pct(m.memory)+'</td><td>'+pct(m.disk)+'</td><td>'+html(n.version||'-')+'</td><td>'+formatTime(n.lastSeen)+'</td><td><span class="actions"><button class="linkbtn" onclick="editNode(\''+safe+'\')">&#32534;&#36753;</button><button class="linkbtn" onclick="copyText(\''+esc(ssh)+'\')">&#22797;&#21046;SSH</button></span></td></tr>'}
function inboundRow(i){return '<tr><td>'+html(i.remark||'-')+'</td><td>'+html(i.protocol||'-')+'</td><td>'+html(i.port||'-')+'</td><td>'+(i.enabled?'\u662f':'\u5426')+'</td><td>'+bytes(i.total)+'</td><td>'+expiry(i.expiryTime)+'</td><td>'+html(i.tag||i.id)+'</td><td><span class="actions"><button class="linkbtn" onclick="editInbound(\''+esc(i.id)+'\')">&#32534;&#36753;</button><button class="linkbtn" onclick="deleteInbound(\''+esc(i.id)+'\')">&#21024;&#38500;</button></span></td></tr>'}
function statusText(v){return v==='online'?'\u5728\u7ebf':(v||'-')}function countryText(v){return !v||v==='asia'||v==='unknown'?'-':v}function pct(v){return typeof v==='number'&&v>0?v.toFixed(1)+'%':'-'}function bytes(v){return v?Math.round(v/1024/1024)+' MB':'\u4e0d\u9650'}function expiry(v){return v?new Date(v).toLocaleString():'\u4e0d\u9650'}function formatTime(v){return v?new Date(v).toLocaleString():'-'}function esc(v){return String(v||'').replace(/\\/g,'\\\\').replace(/'/g,"\\'")}function html(v){return String(v||'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]))}
async function copyText(v){try{await navigator.clipboard.writeText(v);alert('\u5df2\u590d\u5236\uff1a'+v)}catch(e){prompt('\u590d\u5236 SSH \u547d\u4ee4',v)}}async function editNode(id){const ns=await fetch('/api/nodes').then(r=>r.json());const n=ns.find(x=>x.id===id);if(!n)return;const name=prompt('\u8282\u70b9\u540d',n.name||n.id);if(name===null)return;const country=prompt('\u56fd\u5bb6',countryText(n.country||n.region));if(country===null)return;const endpoint=prompt('SSH \u5730\u5740',n.endpoint||n.publicIp||'');if(endpoint===null)return;const sshUser=prompt('SSH \u7528\u6237',n.sshUser||'root');if(sshUser===null)return;await fetch('/api/nodes/'+encodeURIComponent(id)+'/update',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({name,country,endpoint,sshUser})});load()}
async function inboundPrompt(old){old=old||{};const remark=prompt('\u5907\u6ce8',old.remark||'vless-reality');if(remark===null)return null;const protocol=prompt('\u534f\u8bae',old.protocol||'vless');if(protocol===null)return null;const port=Number(prompt('\u7aef\u53e3',old.port||443));if(!port)return null;const settings=prompt('Settings JSON',old.settings||'{"clients":[]}');if(settings===null)return null;const streamSettings=prompt('StreamSettings JSON',old.streamSettings||'{}');if(streamSettings===null)return null;return {...old,remark,protocol,port,enabled:true,settings,streamSettings,sniffing:old.sniffing||'{"enabled":true,"destOverride":["http","tls","quic"]}'}}async function addInbound(){const v=await inboundPrompt();if(!v)return;await fetch('/api/inbounds',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify(v)});load()}async function editInbound(id){const ins=await fetch('/api/inbounds').then(r=>r.json());const old=ins.find(x=>x.id===id);const v=await inboundPrompt(old);if(!v)return;await fetch('/api/inbounds',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify(v)});load()}async function deleteInbound(id){if(!confirm('\u5220\u9664\u8fd9\u4e2a\u5165\u7ad9\uff1f'))return;await fetch('/api/inbounds/'+encodeURIComponent(id),{method:'DELETE'});load()}async function applyXray(){const r=await fetch('/api/xray/apply',{method:'POST'}).then(r=>r.json());alert(r.lastError||'\u5df2\u5e94\u7528\u5e76\u91cd\u542f Xray');load()}load();setInterval(load,15000);
</script></body></html>`
