import { defineStore } from 'pinia'

export interface Node {
  id: string
  name: string
  region: string
  endpoint: string
  version: string
  status: 'pending' | 'online' | 'offline'
  lastSeen: string
}

export const useNodesStore = defineStore('nodes', {
  state: () => ({
    nodes: [] as Node[],
    loading: false,
  }),
  actions: {
    async refresh() {
      this.loading = true
      try {
        const res = await fetch('/api/nodes')
        this.nodes = await res.json()
      } finally {
        this.loading = false
      }
    },
  },
})
