'use client';

import { Building2 } from 'lucide-react';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet';
import { HospitalForm } from '@/components/forms/HospitalForm';
import { useHospitals } from '@/hooks/useHospitals';
import { toast } from 'sonner';
import type { Hospital, CreateHospitalInput, UpdateHospitalInput } from '@/types';

interface HospitalFormDrawerProps {
  open: boolean;
  onClose: () => void;
  hospital?: Hospital;
  onSuccess?: () => void;
}

export function HospitalFormDrawer({
  open,
  onClose,
  hospital,
  onSuccess,
}: HospitalFormDrawerProps) {
  const { createHospital, updateHospital, isCreating, isUpdating } = useHospitals();

  const isEditMode = !!hospital;
  const isLoading = isCreating || isUpdating;

  const handleSubmit = async (data: {
    nome: string;
    codigo: string;
    endereco: string;
    telefone?: string;
    latitude?: number;
    longitude?: number;
    ativo: boolean;
  }) => {
    if (data.latitude === undefined || data.longitude === undefined) {
      toast.error('Por favor, selecione uma localizacao no mapa');
      return;
    }

    try {
      if (isEditMode && hospital) {
        const updateInput: UpdateHospitalInput = {
          nome: data.nome,
          codigo: data.codigo,
          endereco: data.endereco,
          telefone: data.telefone || undefined,
          latitude: data.latitude,
          longitude: data.longitude,
          ativo: data.ativo,
        };
        await updateHospital({ id: hospital.id, input: updateInput });
        toast.success('Hospital atualizado com sucesso!');
      } else {
        const createInput: CreateHospitalInput = {
          nome: data.nome,
          codigo: data.codigo,
          endereco: data.endereco,
          telefone: data.telefone || undefined,
          latitude: data.latitude,
          longitude: data.longitude,
          ativo: data.ativo,
        };
        await createHospital(createInput);
        toast.success('Hospital cadastrado com sucesso!');
      }
      onSuccess?.();
      onClose();
    } catch (error) {
      console.error('Error saving hospital:', error);
      const errorMessage =
        error instanceof Error ? error.message : 'Erro desconhecido';

      if (errorMessage.includes('409') || errorMessage.includes('already exists')) {
        toast.error('Ja existe um hospital com este codigo');
      } else if (errorMessage.includes('400')) {
        toast.error('Dados invalidos. Verifique os campos e tente novamente.');
      } else {
        toast.error(
          isEditMode
            ? 'Erro ao atualizar hospital. Tente novamente.'
            : 'Erro ao cadastrar hospital. Tente novamente.'
        );
      }
    }
  };

  const handleCancel = () => {
    onClose();
  };

  return (
    <Sheet open={open} onOpenChange={(isOpen) => !isOpen && onClose()}>
      <SheetContent side="right" className="w-full sm:max-w-md overflow-y-auto">
        <SheetHeader>
          <SheetTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5 text-primary" />
            {isEditMode ? 'Editar Hospital' : 'Novo Hospital'}
          </SheetTitle>
          <SheetDescription>
            {isEditMode
              ? 'Atualize as informacoes do hospital. A localizacao pode ser ajustada no mapa.'
              : 'Preencha as informacoes do hospital e defina sua localizacao no mapa.'}
          </SheetDescription>
        </SheetHeader>

        {open && (
          <div className="mt-6">
            <HospitalForm
              hospital={hospital}
              onSubmit={handleSubmit}
              onCancel={handleCancel}
              isLoading={isLoading}
            />
          </div>
        )}
      </SheetContent>
    </Sheet>
  );
}
