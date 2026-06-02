package master

import (
	"encoding/json"
	"net/http"
	"strings"

	"xui-next/internal/domain"
	"xui-next/internal/reality"
	"xui-next/internal/version"
)

type Server struct { store *Store; reality *reality.Library }

func NewServer(store *Store, reality *reality.Library) *Server { return &Server{store: store, reality: reality} }

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
	mux.HandleFunc("GET /api/reality/domains", s.realityDomains)
	mux.HandleFunc("POST /api/reality/recommend", s.realityRecommend)
	return mux
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/" { http.NotFound(w, r); return }; w.Header().Set("Content-Type", "text/html; charset=utf-8"); _, _ = w.Write([]byte(indexHTML)) }
func (s *Server) health(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, map[string]string{"status":"ok"}) }
func (s *Server) version(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, map[string]string{"version":version.String()}) }
func (s *Server) nodes(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, s.store.Nodes()) }
func (s *Server) inbounds(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, s.store.Inbounds()) }

func (s *Server) saveInbound(w http.ResponseWriter, r *http.Request) { var inbound domain.Inbound; if !decodeJSON(w,r,&inbound){return}; if inbound.Port<=0{http.Error(w,"port is required",http.StatusBadRequest);return}; writeJSON(w,http.StatusOK,s.store.UpsertInbound(inbound)) }
func (s *Server) deleteInbound(w http.ResponseWriter, r *http.Request) { id:=strings.TrimPrefix(r.URL.Path,"/api/inbounds/"); if id==""||!s.store.DeleteInbound(id){http.NotFound(w,r);return}; writeJSON(w,http.StatusOK,map[string]bool{"deleted":true}) }
func (s *Server) registerNode(w http.ResponseWriter, r *http.Request) { var n domain.Node; if !decodeJSON(w,r,&n){return}; if n.ID==""{http.Error(w,"node id is required",http.StatusBadRequest);return}; writeJSON(w,http.StatusCreated,s.store.UpsertNode(n)) }

func (s *Server) nodeAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/nodes/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 { http.NotFound(w, r); return }
	nodeID, action := parts[0], parts[1]
	switch action {
	case "heartbeat": var req domain.HeartbeatRequest; if !decodeJSON(w,r,&req){return}; req.Node.ID=nodeID; writeJSON(w,http.StatusOK,s.store.SaveHeartbeat(req))
	case "desired-config": cfg, err := s.store.DesiredConfig(nodeID); if err != nil { writeJSON(w,http.StatusOK,domain.DesiredConfig{Version:"empty"}); return }; writeJSON(w,http.StatusOK,cfg)
	case "update": var patch domain.Node; if !decodeJSON(w,r,&patch){return}; node, ok := s.store.UpdateNode(nodeID, patch); if !ok { http.NotFound(w,r); return }; writeJSON(w,http.StatusOK,node)
	default: http.NotFound(w,r)
	}
}

func (s *Server) realityDomains(w http.ResponseWriter, r *http.Request) { writeJSON(w,http.StatusOK,s.reality.Domains(r.URL.Query().Get("region"))) }
func (s *Server) realityRecommend(w http.ResponseWriter, r *http.Request) { var req struct{Region string `json:"region"`; Limit int `json:"limit"`}; if !decodeJSON(w,r,&req){return}; writeJSON(w,http.StatusOK,s.reality.Recommend(req.Region,req.Limit)) }
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool { defer r.Body.Close(); if err:=json.NewDecoder(r.Body).Decode(v); err!=nil{http.Error(w,err.Error(),http.StatusBadRequest);return false}; return true }
func writeJSON(w http.ResponseWriter, status int, v any) { w.Header().Set("Content-Type","application/json; charset=utf-8"); w.WriteHeader(status); _=json.NewEncoder(w).Encode(v) }

const indexHTML = `<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>XUI Pro</title>
<style>:root{color-scheme:dark;font-family:Inter,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;background:#101316;color:#edf5f2}body{margin:0}.shell{max-width:1180px;margin:0 auto;padding:40px 20px}.top{display:flex;justify-content:space-between;gap:16px;align-items:center;margin-bottom:24px}h1{margin:0;font-size:30px}h2{margin:0 0 16px;font-size:24px}p{color:#94a6a1}.badge{border-radius:999px;padding:6px 10px;background:#18382c;color:#65e6ad;font-weight:700}.grid{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:14px}.card,.panel{border:1px solid #29343c;background:#171d22;border-radius:8px;padding:18px}.panel{margin-top:16px;overflow-x:auto}.card span{display:block;color:#94a6a1;margin-bottom:8px}.card strong{font-size:24px;word-break:break-word}table{width:100%;border-collapse:collapse}th,td{text-align:left;padding:10px 8px;border-bottom:1px solid #27323a;white-space:nowrap}th{color:#94a6a1;font-size:13px}a,.linkbtn{color:#65e6ad}.linkbtn{border:0;background:transparent;cursor:pointer;font:inherit;padding:0}.actions{display:flex;gap:12px;align-items:center}code{background:#222b31;padding:2px 6px;border-radius:6px}.toolbar{display:flex;gap:12px;margin-bottom:12px}.primary{border:1px solid #65e6ad;background:#18382c;color:#65e6ad;border-radius:6px;padding:8px 12px;cursor:pointer}@media(max-width:900px){.grid{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(max-width:600px){.top{align-items:flex-start;flex-direction:column}.grid{grid-template-columns:1fr}}</style></head>
<body><main class="shell"><div class="top"><div><h1>XUI Pro</h1><p>&#20027;&#25511;&#26381;&#21153;&#24050;&#21551;&#21160;&#12290;&#24403;&#21069;&#20869;&#32622;&#38754;&#26495;&#29992;&#20110;&#35797;&#35013;&#21644;&#26089;&#26399;&#31649;&#29702;&#12290;</p></div><span class="badge" id="health">&#26816;&#27979;&#20013;</span></div>
<section class="grid"><div class="card"><span>&#33410;&#28857;&#25968;&#37327;</span><strong id="nodeCount">-</strong></div><div class="card"><span>&#22312;&#32447;&#33410;&#28857;</span><strong id="onlineCount">-</strong></div><div class="card"><span>&#20027;&#25511;&#29256;&#26412;</span><strong id="masterVersion">-</strong></div><div class="card"><span>&#20027;&#25511;&#22495;&#21517;</span><strong>xui.ityc.cc</strong></div></section>
<section class="panel"><h2>&#33410;&#28857;&#21015;&#34920;</h2><table><thead><tr><th>&#33410;&#28857;&#21517;</th><th>&#22269;&#23478;</th><th>&#29366;&#24577;</th><th>CPU</th><th>&#20869;&#23384;</th><th>&#30913;&#30424;</th><th>Agent &#29256;&#26412;</th><th>&#26368;&#21518;&#24515;&#36339;</th><th>&#25805;&#20316;</th></tr></thead><tbody id="nodes"><tr><td colspan="9">&#26242;&#26080;&#33410;&#28857;&#65292;&#23433;&#35013; Agent &#21518;&#20250;&#26174;&#31034;&#22312;&#36825;&#37324;&#12290;</td></tr></tbody></table></section>
<section class="panel"><div class="toolbar"><h2 style="flex:1">&#20837;&#31449;&#31649;&#29702;</h2><button class="primary" onclick="addInbound()">&#26032;&#22686;&#20837;&#31449;</button></div><table><thead><tr><th>&#22791;&#27880;</th><th>&#21327;&#35758;</th><th>&#31471;&#21475;</th><th>&#21551;&#29992;</th><th>Settings</th><th>Stream</th><th>&#25805;&#20316;</th></tr></thead><tbody id="inbounds"><tr><td colspan="7">&#26242;&#26080;&#20837;&#31449;&#12290;</td></tr></tbody></table></section>
<section class="panel"><h2>&#19979;&#19968;&#27493;</h2><p>&#20581;&#24247;&#26816;&#26597;&#65306;<a href="/api/health">/api/health</a>&#65292;&#33410;&#28857; API&#65306;<a href="/api/nodes">/api/nodes</a>&#65292;&#20837;&#31449; API&#65306;<a href="/api/inbounds">/api/inbounds</a></p><p>Agent &#31034;&#20363;&#65306;<code>bash &lt;(curl -Ls https://raw.githubusercontent.com/tyrantcwj/xui-pro/main/install.sh) agent --master http://xui.ityc.cc:8080 --token test</code></p></section></main>
<script>
async function load(){const h=await fetch('/api/health').then(r=>r.json()).catch(()=>({status:'error'}));document.getElementById('health').textContent=h.status==='ok'?'正常':'异常';const v=await fetch('/api/version').then(r=>r.json()).catch(()=>({version:'-'}));document.getElementById('masterVersion').textContent=v.version||'-';const ns=await fetch('/api/nodes').then(r=>r.json()).catch(()=>[]);document.getElementById('nodeCount').textContent=ns.length;document.getElementById('onlineCount').textContent=ns.filter(n=>n.status==='online').length;if(ns.length){document.getElementById('nodes').innerHTML=ns.map(nodeRow).join('')}const ins=await fetch('/api/inbounds').then(r=>r.json()).catch(()=>[]);document.getElementById('inbounds').innerHTML=ins.length?ins.map(inboundRow).join(''):'<tr><td colspan="7">暂无入站。</td></tr>'}
function nodeRow(n){const m=n.metrics||{};const target=n.endpoint||n.publicIp||n.name||n.id;const user=n.sshUser||'root';const ssh='ssh '+user+'@'+target;const safe=esc(n.id);return '<tr><td>'+html(n.name||n.id)+'</td><td>'+html(countryText(n.country||n.region))+'</td><td>'+statusText(n.status)+'</td><td>'+pct(m.cpu)+'</td><td>'+pct(m.memory)+'</td><td>'+pct(m.disk)+'</td><td>'+html(n.version||'-')+'</td><td>'+formatTime(n.lastSeen)+'</td><td><span class="actions"><button class="linkbtn" onclick="editNode(\''+safe+'\')">编辑</button><a href="ssh://'+html(user)+'@'+html(target)+'">SSH</a><a href="vscode://vscode-remote/ssh-remote+'+html(user)+'@'+html(target)+'">VSCode</a><button class="linkbtn" onclick="copyText(\''+esc(ssh)+'\')">复制</button></span></td></tr>'}
function inboundRow(i){return '<tr><td>'+html(i.remark||'-')+'</td><td>'+html(i.protocol||'-')+'</td><td>'+html(i.port||'-')+'</td><td>'+(i.enabled?'是':'否')+'</td><td><code>'+html(short(i.settings))+'</code></td><td><code>'+html(short(i.stream))+'</code></td><td><button class="linkbtn" onclick="deleteInbound(\''+esc(i.id)+'\')">删除</button></td></tr>'}
function statusText(v){return v==='online'?'在线':(v||'-')}function countryText(v){return !v||v==='asia'||v==='unknown'?'-':v}function pct(v){return typeof v==='number'&&v>0?v.toFixed(1)+'%':'-'}function formatTime(v){return v?new Date(v).toLocaleString():'-'}function short(v){v=String(v||'');return v.length>60?v.slice(0,60)+'...':v}function esc(v){return String(v||'').replace(/\\/g,'\\\\').replace(/'/g,"\\'")}function html(v){return String(v||'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]))}
async function copyText(v){try{await navigator.clipboard.writeText(v);alert('已复制：'+v)}catch(e){prompt('复制 SSH 命令',v)}}async function editNode(id){const ns=await fetch('/api/nodes').then(r=>r.json());const n=ns.find(x=>x.id===id);if(!n)return;const name=prompt('节点名',n.name||n.id);if(name===null)return;const country=prompt('国家',countryText(n.country||n.region));if(country===null)return;const endpoint=prompt('SSH 地址（公网 IP 或域名）',n.endpoint||n.publicIp||'');if(endpoint===null)return;const sshUser=prompt('SSH 用户',n.sshUser||'root');if(sshUser===null)return;await fetch('/api/nodes/'+encodeURIComponent(id)+'/update',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({name,country,endpoint,sshUser})});load()}
async function addInbound(){const remark=prompt('备注','vless-reality');if(remark===null)return;const protocol=prompt('协议','vless');if(protocol===null)return;const port=Number(prompt('端口','443'));if(!port)return;const settings=prompt('Settings JSON','{}');if(settings===null)return;const stream=prompt('Stream JSON','{}');if(stream===null)return;await fetch('/api/inbounds',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({remark,protocol,port,enabled:true,settings,stream})});load()}async function deleteInbound(id){if(!confirm('删除这个入站？'))return;await fetch('/api/inbounds/'+encodeURIComponent(id),{method:'DELETE'});load()}load();setInterval(load,15000);
</script></body></html>`
