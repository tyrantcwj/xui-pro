package master

import (
	"encoding/json"
	"net/http"
	"strings"

	"xui-next/internal/domain"
	"xui-next/internal/reality"
)

type Server struct {
	store   *Store
	reality *reality.Library
}

func NewServer(store *Store, reality *reality.Library) *Server {
	return &Server{store: store, reality: reality}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.index)
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/nodes", s.nodes)
	mux.HandleFunc("POST /api/nodes/register", s.registerNode)
	mux.HandleFunc("POST /api/nodes/", s.nodeAction)
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

func (s *Server) nodes(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Nodes())
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
<html lang="zh-CN">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>XUI Pro</title>
<style>:root{color-scheme:dark;font-family:Inter,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;background:#101316;color:#edf5f2}body{margin:0}.shell{max-width:1180px;margin:0 auto;padding:40px 20px}.top{display:flex;justify-content:space-between;gap:16px;align-items:center;margin-bottom:24px}h1{margin:0;font-size:30px}p{color:#94a6a1}.badge{border-radius:999px;padding:6px 10px;background:#18382c;color:#65e6ad;font-weight:700}.grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:14px}.card,.panel{border:1px solid #29343c;background:#171d22;border-radius:8px;padding:18px}.panel{margin-top:16px;overflow-x:auto}.card span{display:block;color:#94a6a1;margin-bottom:8px}.card strong{font-size:28px}table{width:100%;border-collapse:collapse}th,td{text-align:left;padding:10px 8px;border-bottom:1px solid #27323a;white-space:nowrap}th{color:#94a6a1;font-size:13px}a,.linkbtn{color:#65e6ad}.linkbtn{border:0;background:transparent;cursor:pointer;font:inherit;padding:0}.actions{display:flex;gap:12px}code{background:#222b31;padding:2px 6px;border-radius:6px}@media(max-width:760px){.top{align-items:flex-start;flex-direction:column}.grid{grid-template-columns:1fr}}</style></head>
<body><main class="shell"><div class="top"><div><h1>XUI Pro</h1><p>主控服务已启动。当前是早期内置面板，完整 Vue 面板会继续完善。</p></div><span class="badge" id="health">检测中</span></div><section class="grid"><div class="card"><span>节点数量</span><strong id="nodeCount">-</strong></div><div class="card"><span>在线节点</span><strong id="onlineCount">-</strong></div><div class="card"><span>主控域名</span><strong>xui.ityc.cc</strong></div></section><section class="panel"><h2>节点列表</h2><table><thead><tr><th>节点名</th><th>国家</th><th>状态</th><th>CPU</th><th>内存</th><th>磁盘</th><th>版本</th><th>最后心跳</th><th>操作</th></tr></thead><tbody id="nodes"><tr><td colspan="9">暂无节点，安装 Agent 后会显示在这里。</td></tr></tbody></table></section><section class="panel"><h2>下一步</h2><p>健康检查：<a href="/api/health">/api/health</a>，节点 API：<a href="/api/nodes">/api/nodes</a></p><p>Agent 示例：<code>bash &lt;(curl -Ls https://raw.githubusercontent.com/tyrantcwj/xui-pro/main/install.sh) agent --master http://xui.ityc.cc:8080 --country China --token test</code></p></section></main>
<script>
async function load(){const h=await fetch('/api/health').then(r=>r.json()).catch(()=>({status:'error'}));document.getElementById('health').textContent=h.status==='ok'?'正常':'异常';const ns=await fetch('/api/nodes').then(r=>r.json()).catch(()=>[]);document.getElementById('nodeCount').textContent=ns.length;document.getElementById('onlineCount').textContent=ns.filter(n=>n.status==='online').length;if(ns.length){document.getElementById('nodes').innerHTML=ns.map(n=>{const m=n.metrics||{};const target=n.endpoint||n.publicIp||n.name||n.id;const user=n.sshUser||'root';const ssh='ssh '+user+'@'+target;const safe=esc(n.id);return '<tr><td>'+(n.name||n.id)+'</td><td>'+(n.country||n.region||'-')+'</td><td>'+statusText(n.status)+'</td><td>'+pct(m.cpu)+'</td><td>'+pct(m.memory)+'</td><td>'+pct(m.disk)+'</td><td>'+(n.version||'-')+'</td><td>'+formatTime(n.lastSeen)+'</td><td><span class="actions"><button class="linkbtn" onclick="editNode(\''+safe+'\')">编辑</button><a href="ssh://'+user+'@'+target+'">SSH</a><a href="vscode://vscode-remote/ssh-remote+'+user+'@'+target+'">VSCode</a><button class="linkbtn" onclick="copyText(\''+esc(ssh)+'\')">复制</button></span></td></tr>'}).join('')}}
function statusText(v){return v==='online'?'在线':(v||'-')}function pct(v){return typeof v==='number'&&v>0?v.toFixed(1)+'%':'-'}function formatTime(v){return v?new Date(v).toLocaleString():'-'}function esc(v){return String(v||'').replace(/\\/g,'\\\\').replace(/'/g,"\\'")}async function copyText(v){try{await navigator.clipboard.writeText(v);alert('已复制：'+v)}catch(e){prompt('复制 SSH 命令',v)}}async function editNode(id){const ns=await fetch('/api/nodes').then(r=>r.json());const n=ns.find(x=>x.id===id);if(!n)return;const name=prompt('节点名',n.name||n.id);if(name===null)return;const country=prompt('国家',n.country||n.region||'');if(country===null)return;const endpoint=prompt('SSH 地址（公网 IP 或域名）',n.endpoint||n.publicIp||'');if(endpoint===null)return;const sshUser=prompt('SSH 用户',n.sshUser||'root');if(sshUser===null)return;await fetch('/api/nodes/'+encodeURIComponent(id)+'/update',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({name,country,endpoint,sshUser})});load()}load();setInterval(load,15000);
</script></body></html>`
