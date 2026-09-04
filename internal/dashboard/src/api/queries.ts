import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import { fetchSnapshot, refreshSnapshot, stopApplication, terminateAgent } from "./client"
import { fetchResourceHistory } from "./resources"

const REFRESH_INTERVAL = 30_000
const SNAPSHOT_STALE_TIME = 15_000

export const snapshotKeys = {
  all: ["snapshot"] as const,
  resources: ["resources"] as const,
}

export function useSnapshot() {
  return useQuery({
    queryKey: snapshotKeys.all,
    queryFn: fetchSnapshot,
    refetchInterval: REFRESH_INTERVAL,
    staleTime: SNAPSHOT_STALE_TIME,
    retry: 1,
  })
}

export function useResourceHistory() {
  return useQuery({
    queryKey: snapshotKeys.resources,
    queryFn: fetchResourceHistory,
    refetchInterval: REFRESH_INTERVAL,
    staleTime: SNAPSHOT_STALE_TIME,
    retry: 1,
  })
}

export function useRefreshSnapshot() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: refreshSnapshot,
    onSuccess: (snapshot) => {
      queryClient.setQueryData(snapshotKeys.all, snapshot)
      void queryClient.invalidateQueries({ queryKey: snapshotKeys.resources })
    },
  })
}

export function useStopApplication() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: stopApplication,
    onSuccess: (snapshot) => {
      queryClient.setQueryData(snapshotKeys.all, snapshot)
      void queryClient.invalidateQueries({ queryKey: snapshotKeys.resources })
    },
  })
}

export function useTerminateAgent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: terminateAgent,
    onSuccess: async () => {
      await new Promise((resolve) => window.setTimeout(resolve, 500))
      await queryClient.invalidateQueries({ queryKey: snapshotKeys.all })
      await queryClient.invalidateQueries({ queryKey: snapshotKeys.resources })
    },
  })
}
