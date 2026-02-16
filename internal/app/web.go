package app

const indexHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width,initial-scale=1" />
  <title>翻鸡缺一门 - 大厅</title>
  <style>
    body { font-family: -apple-system,BlinkMacSystemFont,Segoe UI,Roboto,Helvetica,Arial,sans-serif; margin: 0; background: #f5f6fa; }
    .wrap { max-width: 1024px; margin: 0 auto; padding: 16px; }
    .card { background: white; border-radius: 10px; padding: 14px; margin-bottom: 12px; box-shadow: 0 1px 3px rgba(0,0,0,.06); }
    input, button, select { padding: 8px; margin: 4px; }
    table { border-collapse: collapse; width: 100%; }
    th, td { border-bottom: 1px solid #eee; padding: 8px; text-align: left; }
    .row { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
    .hidden { display: none; }
  </style>
</head>
<body>
<div class="wrap">
  <div class="card" id="loginCard">
    <h3>登录 / 注册（演示）</h3>
    <div class="row">
      <input id="username" placeholder="用户名" />
      <input id="password" placeholder="密码" type="password" />
      <button onclick="login()">登录</button>
    </div>
    <div id="loginMsg"></div>
  </div>

  <div id="app" class="hidden">
    <div class="card">
      <div class="row"><strong id="who"></strong><button onclick="logout()">退出</button></div>
    </div>

    <div class="card">
      <h3>大厅（参考腾讯麻将流程）</h3>
      <button onclick="loadHall()">刷新大厅</button>
      <h4>房间列表</h4>
      <table><thead><tr><th>房间</th><th>模式</th><th>人数</th></tr></thead><tbody id="rooms"></tbody></table>
      <h4>在线好友</h4>
      <table><thead><tr><th>ID</th><th>昵称</th></tr></thead><tbody id="friends"></tbody></table>
    </div>

    <div class="card">
      <h3>历史记录</h3>
      <button onclick="loadHistory()">刷新历史</button>
      <table><thead><tr><th>局ID</th><th>房间</th><th>赢家</th><th>时间</th></tr></thead><tbody id="history"></tbody></table>
    </div>

    <div class="card">
      <h3>咱俩胜率</h3>
      <div class="row">
        <input id="oppId" placeholder="对手 userId（例如 u-2）" />
        <button onclick="loadH2H()">查询</button>
      </div>
      <pre id="h2h"></pre>
    </div>
  </div>
</div>
<script>
let me = null;
async function login() {
  const username = document.getElementById('username').value;
  const password = document.getElementById('password').value;
  const r = await fetch('/api/login', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({username,password})});
  const msg = document.getElementById('loginMsg');
  if (!r.ok) { msg.innerText = await r.text(); return; }
  const data = await r.json();
  me = data.user;
  msg.innerText = '登录成功';
  document.getElementById('loginCard').classList.add('hidden');
  document.getElementById('app').classList.remove('hidden');
  document.getElementById('who').innerText = '当前用户：' + me.username + ' (' + me.id + ')';
  loadHall(); loadHistory();
}
async function logout() {
  await fetch('/api/logout', {method:'POST'});
  location.reload();
}
async function loadHall() {
  const r = await fetch('/api/hall');
  if (!r.ok) return;
  const d = await r.json();
  document.getElementById('rooms').innerHTML = d.rooms.map(function(x){ return '<tr><td>'+x.name+'</td><td>'+x.mode+'</td><td>'+x.currentPeople+'/'+x.capacity+'</td></tr>'; }).join('');
  document.getElementById('friends').innerHTML = d.onlineFriends.map(function(x){ return '<tr><td>'+x.id+'</td><td>'+x.username+'</td></tr>'; }).join('');
}
async function loadHistory() {
  const r = await fetch('/api/history?limit=20');
  if (!r.ok) return;
  const d = await r.json();
  document.getElementById('history').innerHTML = d.records.map(function(x){ return '<tr><td>'+x.id+'</td><td>'+x.roomId+'</td><td>'+((x.winners||[]).join(','))+'</td><td>'+x.playedAt+'</td></tr>'; }).join('');
}
async function loadH2H() {
  const opp = document.getElementById('oppId').value;
  const r = await fetch('/api/stats/headtohead?opponent=' + encodeURIComponent(opp));
  if (!r.ok) { document.getElementById('h2h').innerText = await r.text(); return; }
  const d = await r.json();
  document.getElementById('h2h').innerText = JSON.stringify(d, null, 2);
}
</script>
</body>
</html>`
