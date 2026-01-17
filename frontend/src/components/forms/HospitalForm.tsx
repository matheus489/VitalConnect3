'use client';

import { useState, useEffect, useCallback } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Loader2, MapPin, Search } from 'lucide-react';
import dynamic from 'next/dynamic';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { useGeocoding } from '@/hooks/useGeocoding';
import { parseCoordinates, reverseGeocode } from '@/services/nominatim';
import type { Hospital, NominatimResult } from '@/types';

const LocationPickerMap = dynamic(
  () => import('@/components/map/LocationPickerMap'),
  {
    ssr: false,
    loading: () => (
      <div className="w-full h-[250px] rounded-lg border bg-muted flex items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    ),
  }
);

const hospitalFormSchema = z.object({
  nome: z
    .string()
    .min(3, 'Nome deve ter no minimo 3 caracteres')
    .max(255, 'Nome muito longo'),
  codigo: z
    .string()
    .min(2, 'Codigo deve ter no minimo 2 caracteres')
    .max(50, 'Codigo muito longo')
    .regex(/^[a-zA-Z0-9]+$/, 'Codigo deve conter apenas letras e numeros'),
  endereco: z
    .string()
    .min(1, 'Endereco e obrigatorio')
    .max(500, 'Endereco muito longo'),
  telefone: z.string().optional(),
  latitude: z.number().min(-90).max(90).optional(),
  longitude: z.number().min(-180).max(180).optional(),
  ativo: z.boolean(),
});

type HospitalFormValues = z.infer<typeof hospitalFormSchema>;

interface HospitalFormProps {
  hospital?: Hospital;
  onSubmit: (data: HospitalFormValues) => Promise<void>;
  onCancel: () => void;
  isLoading?: boolean;
}

function formatPhoneNumber(value: string): string {
  const cleaned = value.replace(/\D/g, '');
  if (cleaned.length <= 2) {
    return cleaned;
  }
  if (cleaned.length <= 6) {
    return `(${cleaned.slice(0, 2)}) ${cleaned.slice(2)}`;
  }
  if (cleaned.length <= 10) {
    return `(${cleaned.slice(0, 2)}) ${cleaned.slice(2, 6)}-${cleaned.slice(6)}`;
  }
  return `(${cleaned.slice(0, 2)}) ${cleaned.slice(2, 7)}-${cleaned.slice(7, 11)}`;
}

export function HospitalForm({
  hospital,
  onSubmit,
  onCancel,
  isLoading = false,
}: HospitalFormProps) {
  const [addressQuery, setAddressQuery] = useState('');
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [isReverseGeocoding, setIsReverseGeocoding] = useState(false);
  const [markerPosition, setMarkerPosition] = useState<{ lat: number; lng: number } | null>(
    hospital?.latitude && hospital?.longitude
      ? { lat: hospital.latitude, lng: hospital.longitude }
      : null
  );
  const { suggestions, isLoading: isSearching } = useGeocoding(addressQuery);

  const isEditMode = !!hospital;

  const form = useForm<HospitalFormValues>({
    resolver: zodResolver(hospitalFormSchema),
    defaultValues: {
      nome: hospital?.nome ?? '',
      codigo: hospital?.codigo ?? '',
      endereco: hospital?.endereco ?? '',
      telefone: hospital?.telefone ?? '',
      latitude: hospital?.latitude,
      longitude: hospital?.longitude,
      ativo: hospital?.ativo ?? true,
    },
  });

  const hasCoordinates = markerPosition !== null;

  const handleLocationChange = useCallback(
    async (lat: number, lng: number) => {
      // Update marker position immediately
      setMarkerPosition({ lat, lng });
      form.setValue('latitude', lat, { shouldValidate: true });
      form.setValue('longitude', lng, { shouldValidate: true });

      // Reverse geocode to update address field
      setIsReverseGeocoding(true);
      try {
        const address = await reverseGeocode(lat, lng);
        if (address) {
          form.setValue('endereco', address, { shouldValidate: true });
          setAddressQuery('');
        }
      } finally {
        setIsReverseGeocoding(false);
      }
    },
    [form]
  );

  const handleSuggestionSelect = useCallback(
    (suggestion: NominatimResult) => {
      const coords = parseCoordinates(suggestion);
      // Update marker position immediately
      setMarkerPosition({ lat: coords.latitude, lng: coords.longitude });
      form.setValue('endereco', suggestion.display_name, { shouldValidate: true });
      form.setValue('latitude', coords.latitude, { shouldValidate: true });
      form.setValue('longitude', coords.longitude, { shouldValidate: true });
      setAddressQuery('');
      setShowSuggestions(false);
    },
    [form]
  );

  const handleAddressChange = useCallback(
    (value: string) => {
      setAddressQuery(value);
      form.setValue('endereco', value);
      setShowSuggestions(true);
    },
    [form]
  );

  const handleFormSubmit = async (data: HospitalFormValues) => {
    await onSubmit(data);
  };

  useEffect(() => {
    const handleClickOutside = () => {
      setShowSuggestions(false);
    };
    document.addEventListener('click', handleClickOutside);
    return () => document.removeEventListener('click', handleClickOutside);
  }, []);

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleFormSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="nome"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Nome do Hospital</FormLabel>
              <FormControl>
                <Input
                  placeholder="Ex: Hospital das Clinicas"
                  disabled={isLoading}
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="codigo"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Codigo</FormLabel>
              <FormControl>
                <Input
                  placeholder="Ex: HC001"
                  disabled={isLoading}
                  {...field}
                />
              </FormControl>
              <FormDescription>
                Identificador unico do hospital (apenas letras e numeros)
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="endereco"
          render={({ field }) => (
            <FormItem className="relative">
              <FormLabel>Endereco</FormLabel>
              <FormControl>
                <div className="relative">
                  <Input
                    placeholder="Digite o endereco para buscar..."
                    disabled={isLoading}
                    value={addressQuery || field.value}
                    onChange={(e) => handleAddressChange(e.target.value)}
                    onFocus={() => setShowSuggestions(true)}
                    className="pr-10"
                  />
                  <div className="absolute right-3 top-1/2 -translate-y-1/2">
                    {isSearching || isReverseGeocoding ? (
                      <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                    ) : (
                      <Search className="h-4 w-4 text-muted-foreground" />
                    )}
                  </div>
                </div>
              </FormControl>
              {showSuggestions && addressQuery.length >= 3 && (
                <div
                  className="absolute top-full left-0 z-50 w-full mt-1 bg-background border rounded-md shadow-lg max-h-60 overflow-auto"
                  onClick={(e) => e.stopPropagation()}
                >
                  {isSearching ? (
                    <div className="px-3 py-4 text-center text-sm text-muted-foreground">
                      <Loader2 className="h-4 w-4 animate-spin mx-auto mb-2" />
                      Buscando enderecos...
                    </div>
                  ) : suggestions.length > 0 ? (
                    suggestions.map((suggestion) => (
                      <button
                        key={suggestion.place_id}
                        type="button"
                        className="w-full px-3 py-2 text-left text-sm hover:bg-muted transition-colors border-b last:border-b-0"
                        onClick={() => handleSuggestionSelect(suggestion)}
                      >
                        <div className="flex items-start gap-2">
                          <MapPin className="h-4 w-4 mt-0.5 shrink-0 text-muted-foreground" />
                          <span className="line-clamp-2">{suggestion.display_name}</span>
                        </div>
                      </button>
                    ))
                  ) : (
                    <div className="px-3 py-4 text-center text-sm text-muted-foreground">
                      Nenhum endereco encontrado. Tente um termo mais generico ou clique no mapa.
                    </div>
                  )}
                </div>
              )}
              <FormDescription>
                Busque o endereco ou clique no mapa para definir a localizacao
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="telefone"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Telefone (opcional)</FormLabel>
              <FormControl>
                <Input
                  placeholder="(62) 3333-4444"
                  disabled={isLoading}
                  value={field.value ?? ''}
                  onChange={(e) => {
                    const formatted = formatPhoneNumber(e.target.value);
                    field.onChange(formatted);
                  }}
                  maxLength={15}
                />
              </FormControl>
              <FormDescription>
                Telefone para contato (recepcao ou plantao)
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="space-y-2">
          <FormLabel className="flex items-center gap-2">
            <MapPin className="h-4 w-4" />
            Localizacao no Mapa
          </FormLabel>
          <LocationPickerMap
            initialPosition={
              hospital?.latitude && hospital?.longitude
                ? { lat: hospital.latitude, lng: hospital.longitude }
                : undefined
            }
            markerPosition={markerPosition}
            onLocationChange={handleLocationChange}
            disabled={isLoading}
          />
          {!hasCoordinates && (
            <p className="text-sm text-muted-foreground">
              Clique no mapa ou selecione um endereco para definir a localizacao
            </p>
          )}
          {hasCoordinates && (
            <p className="text-sm text-muted-foreground">
              Localizacao definida. Arraste o marcador para ajustar.
            </p>
          )}
        </div>

        <FormField
          control={form.control}
          name="ativo"
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3">
              <div className="space-y-0.5">
                <FormLabel>Hospital Ativo</FormLabel>
                <FormDescription>
                  Hospitais inativos nao aparecem no mapa
                </FormDescription>
              </div>
              <FormControl>
                <Switch
                  checked={field.value}
                  onCheckedChange={field.onChange}
                  disabled={isLoading}
                />
              </FormControl>
            </FormItem>
          )}
        />

        <div className="flex gap-3 pt-4">
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            disabled={isLoading}
            className="flex-1"
          >
            Cancelar
          </Button>
          <Button
            type="submit"
            disabled={isLoading || !hasCoordinates}
            className="flex-1"
          >
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                {isEditMode ? 'Salvando...' : 'Cadastrando...'}
              </>
            ) : (
              <>{isEditMode ? 'Salvar Alteracoes' : 'Cadastrar Hospital'}</>
            )}
          </Button>
        </div>
      </form>
    </Form>
  );
}
