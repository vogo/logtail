import axios from "axios"
import { onMounted, ref } from "vue"

export function useTransfers() {
  const loading = ref(false)
  const transfers = ref([])
  function getTransfers() {
    loading.value = true
    axios.get('/manage/transfer/list')
      .then(({ data }) => {
        if (!data) return
        transfers.value = Object.values(data)
      })
      .finally(() => loading.value = false)
  }
  onMounted(getTransfers)
  return {
    loading,
    transfers,
    getTransfers
  }
}