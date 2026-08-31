import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import { fetchSettings, fetchTrustedIPStatus, revokeCurrentTrustedIP, saveSettings, trustCurrentIP } from "./settings"

const settingsKeys = {
  configuration: ["settings", "configuration"] as const,
  trustedIP: ["settings", "trusted-ip"] as const,
}

export function useSettings() {
  return useQuery({ queryKey: settingsKeys.configuration, queryFn: fetchSettings })
}

export function useSaveSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: saveSettings,
    onSuccess: (settings) => queryClient.setQueryData(settingsKeys.configuration, settings),
  })
}

export function useTrustedIPStatus() {
  return useQuery({ queryKey: settingsKeys.trustedIP, queryFn: fetchTrustedIPStatus })
}

export function useRevokeTrustedIP() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: revokeCurrentTrustedIP,
    onSuccess: (status) => queryClient.setQueryData(settingsKeys.trustedIP, status),
  })
}

export function useTrustCurrentIP() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: trustCurrentIP,
    onSuccess: (status) => queryClient.setQueryData(settingsKeys.trustedIP, status),
  })
}
