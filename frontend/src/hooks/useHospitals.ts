'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { Hospital } from '@/types';

interface HospitalsResponse {
  data: Hospital[];
  total: number;
}

export function useHospitals() {
  return useQuery({
    queryKey: ['hospitals'],
    queryFn: async () => {
      const response = await api.get<HospitalsResponse>('/hospitals');
      return response.data.data;
    },
  });
}
