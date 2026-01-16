'use client';

import { useState } from 'react';
import { Calendar, Clock, Users, AlertTriangle, Sun, Moon } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useHospitals, useShifts, useTodayShifts, useCoverageGaps } from '@/hooks';
import { useAuth } from '@/hooks/useAuth';
import type { Shift, TodayShift, DayOfWeek } from '@/types';
import { DayNames } from '@/types';

function ShiftCard({ shift, isToday = false }: { shift: Shift | TodayShift; isToday?: boolean }) {
  const isActive = 'is_active' in shift ? shift.is_active : false;

  return (
    <div
      className={`flex items-center justify-between p-3 rounded-lg border ${
        isActive ? 'bg-green-50 border-green-200' : 'bg-background'
      }`}
    >
      <div className="flex items-center gap-3">
        <div className={`p-2 rounded-full ${shift.is_night ? 'bg-indigo-100' : 'bg-yellow-100'}`}>
          {shift.is_night ? (
            <Moon className="h-4 w-4 text-indigo-600" />
          ) : (
            <Sun className="h-4 w-4 text-yellow-600" />
          )}
        </div>
        <div>
          <p className="font-medium">{shift.user?.nome || 'Usuario'}</p>
          <p className="text-sm text-muted-foreground">
            {shift.start_time} - {shift.end_time}
          </p>
        </div>
      </div>
      <div className="flex items-center gap-2">
        {isActive && <Badge variant="default">Ativo</Badge>}
        {shift.is_night && (
          <Badge variant="secondary">Noturno</Badge>
        )}
      </div>
    </div>
  );
}

function CoverageCard({ hospitalId }: { hospitalId: string }) {
  const { data: coverage, isLoading } = useCoverageGaps(hospitalId);

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <AlertTriangle className="h-5 w-5" />
            Analise de Cobertura
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground">Carregando...</p>
        </CardContent>
      </Card>
    );
  }

  if (!coverage) return null;

  return (
    <Card className={coverage.has_gaps ? 'border-orange-200 bg-orange-50' : 'border-green-200 bg-green-50'}>
      <CardHeader>
        <CardTitle className="text-lg flex items-center gap-2">
          <AlertTriangle className={`h-5 w-5 ${coverage.has_gaps ? 'text-orange-500' : 'text-green-500'}`} />
          Analise de Cobertura
        </CardTitle>
        <CardDescription>
          {coverage.has_gaps
            ? `${coverage.gaps.length} periodo(s) sem cobertura encontrado(s)`
            : 'Cobertura completa para todos os dias'}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          <p className="text-sm font-medium">Total de escalas: {coverage.total_shifts}</p>
          {coverage.has_gaps && (
            <div className="mt-4 space-y-2">
              <p className="text-sm font-medium text-orange-700">Gaps encontrados:</p>
              {coverage.gaps.map((gap, idx) => (
                <div key={idx} className="text-sm p-2 bg-orange-100 rounded">
                  <span className="font-medium">{gap.day_name}:</span> {gap.start_time} - {gap.end_time}
                </div>
              ))}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

export default function ShiftsPage() {
  const { user } = useAuth();
  const { data: hospitals = [] } = useHospitals();
  const [selectedHospitalId, setSelectedHospitalId] = useState<string>('');

  // Set default hospital for non-admin users
  const effectiveHospitalId = selectedHospitalId || (user?.hospital_id ?? '');

  const { data: shifts = [], isLoading: shiftsLoading } = useShifts(effectiveHospitalId);
  const { data: todayShifts = [], isLoading: todayLoading } = useTodayShifts(effectiveHospitalId);

  // Group shifts by day
  const shiftsByDay = shifts.reduce((acc, shift) => {
    const day = shift.day_of_week;
    if (!acc[day]) acc[day] = [];
    acc[day].push(shift);
    return acc;
  }, {} as Record<DayOfWeek, Shift[]>);

  const canSelectHospital = user?.role === 'admin';

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Escalas de Plantao</h1>
          <p className="text-muted-foreground">
            Visualize e gerencie as escalas dos operadores
          </p>
        </div>

        {canSelectHospital && (
          <Select value={selectedHospitalId} onValueChange={setSelectedHospitalId}>
            <SelectTrigger className="w-[250px]">
              <SelectValue placeholder="Selecione um hospital" />
            </SelectTrigger>
            <SelectContent>
              {hospitals.map((hospital) => (
                <SelectItem key={hospital.id} value={hospital.id}>
                  {hospital.nome}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      </div>

      {!effectiveHospitalId ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Calendar className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground">Selecione um hospital para ver as escalas</p>
          </CardContent>
        </Card>
      ) : (
        <>
          {/* Today's Shifts */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Clock className="h-5 w-5" />
                Plantao de Hoje
              </CardTitle>
              <CardDescription>
                Operadores escalados para o dia de hoje
              </CardDescription>
            </CardHeader>
            <CardContent>
              {todayLoading ? (
                <p className="text-muted-foreground">Carregando...</p>
              ) : todayShifts.length === 0 ? (
                <p className="text-muted-foreground">Nenhuma escala para hoje</p>
              ) : (
                <div className="space-y-2">
                  {todayShifts.map((shift) => (
                    <ShiftCard key={shift.id} shift={shift} isToday />
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Coverage Analysis */}
          <CoverageCard hospitalId={effectiveHospitalId} />

          {/* Weekly Schedule */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Calendar className="h-5 w-5" />
                Escala Semanal
              </CardTitle>
              <CardDescription>
                Visao geral das escalas por dia da semana
              </CardDescription>
            </CardHeader>
            <CardContent>
              {shiftsLoading ? (
                <p className="text-muted-foreground">Carregando...</p>
              ) : shifts.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-8">
                  <Users className="h-12 w-12 text-muted-foreground mb-4" />
                  <p className="text-muted-foreground">Nenhuma escala cadastrada</p>
                </div>
              ) : (
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                  {([0, 1, 2, 3, 4, 5, 6] as DayOfWeek[]).map((day) => (
                    <div key={day} className="border rounded-lg p-3">
                      <h4 className="font-medium mb-2">{DayNames[day]}</h4>
                      {shiftsByDay[day]?.length > 0 ? (
                        <div className="space-y-2">
                          {shiftsByDay[day].map((shift) => (
                            <div
                              key={shift.id}
                              className="text-sm p-2 bg-muted rounded flex items-center justify-between"
                            >
                              <span>{shift.user?.nome || 'Usuario'}</span>
                              <span className="text-muted-foreground">
                                {shift.start_time}-{shift.end_time}
                              </span>
                            </div>
                          ))}
                        </div>
                      ) : (
                        <p className="text-sm text-muted-foreground">Sem escala</p>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
