import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import { fetchSkillsStatus, fetchStatusReport, runSkillsSync } from "./status"

const REFRESH_INTERVAL = 30_000

export const statusKeys = {
  skills: ["status", "skills-sync"] as const,
  environment: ["status", "environment"] as const,
}

export function useSkillsStatus() {
  return useQuery({
    queryKey: statusKeys.skills,
    queryFn: fetchSkillsStatus,
    refetchInterval: REFRESH_INTERVAL,
    staleTime: 10_000,
    retry: 1,
  })
}

export function useRunSkillsSync() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: runSkillsSync,
    onSuccess: (payload) => {
      queryClient.setQueryData(statusKeys.skills, payload.status)
      void queryClient.invalidateQueries({ queryKey: statusKeys.environment })
    },
  })
}

export function useStatusReport() {
  return useQuery({
    queryKey: statusKeys.environment,
    queryFn: fetchStatusReport,
    refetchInterval: REFRESH_INTERVAL,
    staleTime: 10_000,
    retry: 1,
  })
}
