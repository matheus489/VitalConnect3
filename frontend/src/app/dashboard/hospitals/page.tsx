'use client';

import { useState } from 'react';
import { Building2, Plus, Phone, MapPin } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useHospitals } from '@/hooks/useHospitals';
import { useAuth } from '@/hooks/useAuth';
import { Badge } from '@/components/ui/badge';
import { HospitalFormDrawer } from '@/components/hospitals/HospitalFormDrawer';
import type { Hospital } from '@/types';

/**
 * Hospitals management page
 *
 * Features:
 * - List all hospitals in card grid
 * - "Novo Hospital" button for Admin and Gestor roles
 * - Click on hospital card to edit (Admin and Gestor only)
 * - Shows telefone and coordinates when available
 */
export default function HospitalsPage() {
  const { data: hospitals, isLoading, isError } = useHospitals();
  const { user } = useAuth();

  const [drawerOpen, setDrawerOpen] = useState(false);
  const [selectedHospital, setSelectedHospital] = useState<Hospital | undefined>();

  const canManageHospitals = user?.role === 'admin' || user?.role === 'gestor';

  const handleOpenCreate = () => {
    setSelectedHospital(undefined);
    setDrawerOpen(true);
  };

  const handleOpenEdit = (hospital: Hospital) => {
    if (!canManageHospitals) return;
    setSelectedHospital(hospital);
    setDrawerOpen(true);
  };

  const handleCloseDrawer = () => {
    setDrawerOpen(false);
    setSelectedHospital(undefined);
  };

  const handleSuccess = () => {
    setDrawerOpen(false);
    setSelectedHospital(undefined);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Hospitais</h1>
          <p className="text-muted-foreground">
            Hospitais cadastrados no sistema
          </p>
        </div>
        {canManageHospitals && (
          <Button onClick={handleOpenCreate}>
            <Plus className="mr-2 h-4 w-4" />
            Novo Hospital
          </Button>
        )}
      </div>

      {isLoading ? (
        <div className="grid gap-4 md:grid-cols-2">
          {[1, 2].map((i) => (
            <Card key={i} className="animate-pulse">
              <CardHeader>
                <div className="h-6 w-48 bg-muted rounded" />
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  <div className="h-4 w-32 bg-muted rounded" />
                  <div className="h-4 w-64 bg-muted rounded" />
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : isError ? (
        <div className="text-center py-12">
          <p className="text-muted-foreground">Erro ao carregar hospitais</p>
        </div>
      ) : hospitals?.length === 0 ? (
        <div className="text-center py-12">
          <Building2 className="h-12 w-12 mx-auto text-muted-foreground/50 mb-4" />
          <p className="text-muted-foreground">
            Nenhum hospital cadastrado
          </p>
          {canManageHospitals && (
            <Button onClick={handleOpenCreate} className="mt-4">
              <Plus className="mr-2 h-4 w-4" />
              Cadastrar Hospital
            </Button>
          )}
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {hospitals?.map((hospital) => (
            <Card
              key={hospital.id}
              className={
                canManageHospitals
                  ? 'cursor-pointer transition-colors hover:bg-muted/50'
                  : ''
              }
              onClick={() => handleOpenEdit(hospital)}
            >
              <CardHeader className="flex flex-row items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
                    <Building2 className="h-5 w-5 text-primary" />
                  </div>
                  <div>
                    <CardTitle className="text-lg">{hospital.nome}</CardTitle>
                    <p className="text-sm text-muted-foreground">{hospital.codigo}</p>
                  </div>
                </div>
                <Badge variant={hospital.ativo ? 'default' : 'secondary'}>
                  {hospital.ativo ? 'Ativo' : 'Inativo'}
                </Badge>
              </CardHeader>
              <CardContent>
                <div className="space-y-2 text-sm">
                  {hospital.endereco && (
                    <div className="flex items-start gap-2 text-muted-foreground">
                      <MapPin className="h-4 w-4 mt-0.5 shrink-0" />
                      <span className="line-clamp-2">{hospital.endereco}</span>
                    </div>
                  )}
                  {hospital.telefone && (
                    <div className="flex items-center gap-2 text-muted-foreground">
                      <Phone className="h-4 w-4 shrink-0" />
                      <span>{hospital.telefone}</span>
                    </div>
                  )}
                  {hospital.latitude && hospital.longitude && (
                    <div className="text-xs text-muted-foreground/70">
                      Coordenadas: {hospital.latitude.toFixed(6)}, {hospital.longitude.toFixed(6)}
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <HospitalFormDrawer
        open={drawerOpen}
        onClose={handleCloseDrawer}
        hospital={selectedHospital}
        onSuccess={handleSuccess}
      />
    </div>
  );
}
