import type { ListEntry } from '~/composables/useList'

export interface UserSummary {
  id: string
  username: string
  avatar_url: string | null
}

export const useUsers = () => {
  const config = useRuntimeConfig()

  const publicFetch = $fetch.create({
    baseURL: `${config.public.apiBase}/api/`,
  })

  const searchUsers = async (query: string): Promise<UserSummary[]> => {
    if (query.length < 2) return []
    try {
      return await publicFetch<UserSummary[]>(
        `users/search?q=${encodeURIComponent(query)}`
      )
    } catch {
      return []
    }
  }

  const fetchUserList = async (
    username: string
  ): Promise<{ entries: ListEntry[]; notFound: boolean }> => {
    try {
      const entries = await publicFetch<ListEntry[]>(
        `users/${encodeURIComponent(username)}/list`
      )
      return { entries: entries ?? [], notFound: false }
    } catch (err: unknown) {
      const status = (err as { response?: { status?: number } })?.response?.status
      if (status === 404) {
        return { entries: [], notFound: true }
      }
      throw err
    }
  }

  return { searchUsers, fetchUserList }
}
