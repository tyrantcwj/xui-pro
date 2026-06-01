<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useNodesStore } from './stores/nodes'

const nodes = useNodesStore()
const region = ref('global')
const domains = ref<any[]>([])

const onlineCount = computed(() => nodes.nodes.filter((n) => n.status === 'online').length)

async function loadDomains() {
  const res = await fetch('/api/reality/recommend', {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ region: region.value, limit: 5 }),
  })
  domains.value = await res.json()
}

onMounted(async () => {
  await Promise.all([nodes.refresh(), loadDomains()])
})
</script>

<template>
  <main class="shell">
    <aside class="sidebar">
      <div class="brand">XUI Next</div>
      <button class="nav active">Dashboard</button>
      <button class="nav">Nodes</button>
      <button class="nav">Inbounds</button>
      <button class="nav">Reality</button>
      <button class="nav">Audit</button>
    </aside>

    <section class="content">
      <header class="topbar">
        <div>
          <h1>Unified Xray Control Plane</h1>
          <p>{{ onlineCount }} online nodes · {{ nodes.nodes.length }} total nodes</p>
        </div>
        <button class="primary" @click="nodes.refresh">Refresh</button>
      </header>

      <div class="stats">
        <article>
          <span>Nodes</span>
          <strong>{{ nodes.nodes.length }}</strong>
        </article>
        <article>
          <span>Online</span>
          <strong>{{ onlineCount }}</strong>
        </article>
        <article>
          <span>Reality Pool</span>
          <strong>{{ domains.length }}</strong>
        </article>
      </div>

      <section class="grid">
        <div class="panel">
          <div class="panel-head">
            <h2>Node Fleet</h2>
            <span v-if="nodes.loading">Loading</span>
          </div>
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>Region</th>
                <th>Status</th>
                <th>Version</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="node in nodes.nodes" :key="node.id">
                <td>{{ node.name }}</td>
                <td>{{ node.region }}</td>
                <td><span class="badge">{{ node.status }}</span></td>
                <td>{{ node.version }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="panel">
          <div class="panel-head">
            <h2>Reality Recommendations</h2>
            <select v-model="region" @change="loadDomains">
              <option value="global">Global</option>
              <option value="asia">Asia</option>
              <option value="north-america">North America</option>
            </select>
          </div>
          <ul class="domains">
            <li v-for="item in domains" :key="item.domain">
              <span>{{ item.domain }}</span>
              <small>{{ item.region }} · {{ item.latencyMs || '-' }}ms</small>
            </li>
          </ul>
        </div>
      </section>
    </section>
  </main>
</template>
