import axios from "axios"
import { onMounted, ref } from "vue"

export function useRouters() {
  const loading = ref(false)
  const routers = ref([])
  function getRouters() {
    loading.value = true
    axios.get('/manage/router/list')
      .then(({ data }) => {
        if (!data) return
        routers.value = Object.values(data)
      })
      .finally(() => loading.value = false)
  }
  onMounted(getRouters)
  return {
    loading,
    routers,
    getRouters
  }
}