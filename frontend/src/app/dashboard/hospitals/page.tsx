'use client';

import { Building2 } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useHospitals } from '@/hooks/useHospitals';
import { Badge } from '@/components/ui/badge';

export default function HospitalsPage() {
  const { data: hospitals, isLoading, isError } = useHospitals();

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Hospitais</h1>
        <p className="text-muted-foreground">
          Hospitais cadastrados no sistema
        </p>
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
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {hospitals?.map((hospital) => (
            <Card key={hospital.id}>
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
                  <p className="text-muted-foreground">{hospital.endereco}</p>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
