'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { TriagemRule, CreateTriagemRuleInput, UpdateTriagemRuleInput } from '@/types';

interface TriagemRulesResponse {
  data: TriagemRule[];
  total: number;
}

// Fetch all triagem rules
export function useTriagemRules() {
  return useQuery<TriagemRulesResponse>({
    queryKey: ['triagem-rules'],
    queryFn: async () => {
      const { data } = await api.get<TriagemRulesResponse>('/triagem-rules');
      return data;
    },
  });
}

// Create a new triagem rule
export function useCreateTriagemRule() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: CreateTriagemRuleInput) => {
      const { data } = await api.post<TriagemRule>('/triagem-rules', input);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['triagem-rules'] });
    },
  });
}

// Update a triagem rule
export function useUpdateTriagemRule() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, input }: { id: string; input: UpdateTriagemRuleInput }) => {
      const { data } = await api.patch<TriagemRule>(`/triagem-rules/${id}`, input);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['triagem-rules'] });
    },
  });
}

// Toggle rule active status
export function useToggleTriagemRule() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, ativo }: { id: string; ativo: boolean }) => {
      const { data } = await api.patch<TriagemRule>(`/triagem-rules/${id}`, { ativo });
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['triagem-rules'] });
    },
  });
}

// Delete (soft delete) a triagem rule
export function useDeleteTriagemRule() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/triagem-rules/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['triagem-rules'] });
    },
  });
}

export default useTriagemRules;
