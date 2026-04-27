export interface ListEntry {
  media_id: string
  external_id: string
  source: string
  type: string
  title: string
  poster_path: string | null
  release_year: number | null
  status: 'watched' | 'plan_to_watch'
  rating: number | null
  review: string | null
  added_at: string
}

export interface AddOrUpdatePayload {
  external_id: string
  source: string
  type: string
  status: 'watched' | 'plan_to_watch'
  rating?: number | null
  review?: string | null
}

export interface UpdatePayload {
  status: 'watched' | 'plan_to_watch'
  rating?: number | null
  review?: string | null
}

export const useList = () => {
  const { apiFetch } = useAuth()

  const list = useState<ListEntry[]>('user-list', () => [])
  const listLoading = useState<boolean>('user-list-loading', () => false)

  const fetchList = async () => {
    listLoading.value = true
    try {
      const data = await apiFetch<ListEntry[]>('list')
      list.value = data ?? []
    } finally {
      listLoading.value = false
    }
  }

  const addOrUpdate = async (payload: AddOrUpdatePayload) => {
    await apiFetch('list', { method: 'POST', body: payload })
    await fetchList()
  }

  const update = async (mediaID: string, payload: UpdatePayload) => {
    await apiFetch(`list/${mediaID}`, { method: 'PUT', body: payload })
    await fetchList()
  }

  const remove = async (mediaID: string) => {
    await apiFetch(`list/${mediaID}`, { method: 'DELETE' })
    list.value = list.value.filter((e) => e.media_id !== mediaID)
  }

  const fetchStatus = async (
    externalID: string,
    source = 'tmdb'
  ): Promise<ListEntry | null> => {
    try {
      return await apiFetch<ListEntry>(
        `list/status/${externalID}?source=${source}`
      )
    } catch {
      return null
    }
  }

  return {
    list,
    listLoading,
    fetchList,
    addOrUpdate,
    update,
    remove,
    fetchStatus,
  }
}
