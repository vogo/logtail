import axios from "axios"
import { onMounted, ref } from "vue"

export function useServerTypes() {
  const loading = ref(false)
  const serverTypes = ref([] as any[])
  function getServerTypes() {
    loading.value = true
    axios.get('/manage/server/types')
      .then(({ data }) => {
        if (!Array.isArray(data)) return []
        serverTypes.value = data
      })
      .finally(() => loading.value = false)
  }
  onMounted(getServerTypes)
  return {
    loading,
    serverTypes,
    getServerTypes
  }
}

export function useServers() {
  const loading = ref(false)
  const servers = ref([])
  function getServers() {
    loading.value = true
    axios.get('/manage/server/list')
      .then(({ data }) => {
        if (!data) return
        servers.value = Object.values(data)
      })
      .finally(() => loading.value = false)
  }
  onMounted(getServers)
  return {
    loading,
    servers,
    getServers
  }
}