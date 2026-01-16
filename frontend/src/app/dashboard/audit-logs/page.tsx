'use client';

import { useState } from 'react';
import { History, Filter, AlertCircle, Info, AlertTriangle } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { useAuditLogs, useHospitals } from '@/hooks';
import { useAuth } from '@/hooks/useAuth';
import type { AuditLog, AuditLogFilters, Severity } from '@/types';

function SeverityBadge({ severity }: { severity: Severity }) {
  const config = {
    INFO: { variant: 'secondary' as const, icon: Info, className: 'text-blue-600' },
    WARN: { variant: 'default' as const, icon: AlertTriangle, className: 'text-yellow-600' },
    CRITICAL: { variant: 'destructive' as const, icon: AlertCircle, className: 'text-red-600' },
  };

  const { variant, icon: Icon, className } = config[severity];

  return (
    <Badge variant={variant} className="gap-1">
      <Icon className={`h-3 w-3 ${className}`} />
      {severity}
    </Badge>
  );
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleString('pt-BR', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export default function AuditLogsPage() {
  const { user } = useAuth();
  const { data: hospitals = [] } = useHospitals();

  const [filters, setFilters] = useState<AuditLogFilters>({
    page: 1,
    page_size: 20,
  });

  const [tempFilters, setTempFilters] = useState({
    data_inicio: '',
    data_fim: '',
    entidade_tipo: '',
    severity: '',
    hospital_id: '',
  });

  const { data, isLoading } = useAuditLogs(filters);
  const logs = data?.data ?? [];
  const meta = data?.meta;

  const canSelectHospital = user?.role === 'admin';

  const handleApplyFilters = () => {
    const newFilters: AuditLogFilters = {
      page: 1,
      page_size: 20,
    };

    if (tempFilters.data_inicio) newFilters.data_inicio = tempFilters.data_inicio;
    if (tempFilters.data_fim) newFilters.data_fim = tempFilters.data_fim;
    if (tempFilters.entidade_tipo && tempFilters.entidade_tipo !== 'all') newFilters.entidade_tipo = tempFilters.entidade_tipo;
    if (tempFilters.severity && tempFilters.severity !== 'all') newFilters.severity = tempFilters.severity as Severity;
    if (tempFilters.hospital_id && tempFilters.hospital_id !== 'all') newFilters.hospital_id = tempFilters.hospital_id;

    setFilters(newFilters);
  };

  const handleClearFilters = () => {
    setTempFilters({
      data_inicio: '',
      data_fim: '',
      entidade_tipo: '',
      severity: '',
      hospital_id: '',
    });
    setFilters({ page: 1, page_size: 20 });
  };

  const handlePageChange = (newPage: number) => {
    setFilters((prev) => ({ ...prev, page: newPage }));
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Logs de Auditoria</h1>
        <p className="text-muted-foreground">
          Historico de acoes realizadas no sistema
        </p>
      </div>

      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <Filter className="h-5 w-5" />
            Filtros
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <div className="space-y-2">
              <Label htmlFor="data-inicio">Data Inicio</Label>
              <Input
                id="data-inicio"
                type="date"
                value={tempFilters.data_inicio}
                onChange={(e) =>
                  setTempFilters((prev) => ({ ...prev, data_inicio: e.target.value }))
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="data-fim">Data Fim</Label>
              <Input
                id="data-fim"
                type="date"
                value={tempFilters.data_fim}
                onChange={(e) =>
                  setTempFilters((prev) => ({ ...prev, data_fim: e.target.value }))
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="entidade">Tipo de Entidade</Label>
              <Select
                value={tempFilters.entidade_tipo}
                onValueChange={(v) =>
                  setTempFilters((prev) => ({ ...prev, entidade_tipo: v }))
                }
              >
                <SelectTrigger id="entidade">
                  <SelectValue placeholder="Todas" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Todas</SelectItem>
                  <SelectItem value="Ocorrencia">Ocorrencia</SelectItem>
                  <SelectItem value="Regra">Regra</SelectItem>
                  <SelectItem value="Usuario">Usuario</SelectItem>
                  <SelectItem value="Hospital">Hospital</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="severity">Severidade</Label>
              <Select
                value={tempFilters.severity}
                onValueChange={(v) =>
                  setTempFilters((prev) => ({ ...prev, severity: v }))
                }
              >
                <SelectTrigger id="severity">
                  <SelectValue placeholder="Todas" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Todas</SelectItem>
                  <SelectItem value="INFO">INFO</SelectItem>
                  <SelectItem value="WARN">WARN</SelectItem>
                  <SelectItem value="CRITICAL">CRITICAL</SelectItem>
                </SelectContent>
              </Select>
            </div>
            {canSelectHospital && (
              <div className="space-y-2">
                <Label htmlFor="hospital">Hospital</Label>
                <Select
                  value={tempFilters.hospital_id}
                  onValueChange={(v) =>
                    setTempFilters((prev) => ({ ...prev, hospital_id: v }))
                  }
                >
                  <SelectTrigger id="hospital">
                    <SelectValue placeholder="Todos" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">Todos</SelectItem>
                    {hospitals.map((hospital) => (
                      <SelectItem key={hospital.id} value={hospital.id}>
                        {hospital.nome}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}
          </div>
          <div className="flex gap-2 mt-4">
            <Button onClick={handleApplyFilters}>Aplicar Filtros</Button>
            <Button variant="outline" onClick={handleClearFilters}>
              Limpar
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Logs Table */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <History className="h-5 w-5" />
            Registros
          </CardTitle>
          {meta && (
            <CardDescription>
              Mostrando {logs.length} de {meta.total} registros
            </CardDescription>
          )}
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-muted-foreground text-center py-8">Carregando...</p>
          ) : logs.length === 0 ? (
            <p className="text-muted-foreground text-center py-8">
              Nenhum registro encontrado
            </p>
          ) : (
            <>
              <div className="rounded-md border overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Data/Hora</TableHead>
                      <TableHead>Usuario</TableHead>
                      <TableHead>Acao</TableHead>
                      <TableHead>Entidade</TableHead>
                      <TableHead>Severidade</TableHead>
                      {canSelectHospital && <TableHead>Hospital</TableHead>}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {logs.map((log) => (
                      <TableRow key={log.id}>
                        <TableCell className="whitespace-nowrap">
                          {formatDate(log.timestamp)}
                        </TableCell>
                        <TableCell>{log.actor_name}</TableCell>
                        <TableCell className="font-mono text-sm">{log.acao}</TableCell>
                        <TableCell>
                          <span className="font-medium">{log.entidade_tipo}</span>
                          <br />
                          <span className="text-xs text-muted-foreground">
                            {log.entidade_id.slice(0, 8)}...
                          </span>
                        </TableCell>
                        <TableCell>
                          <SeverityBadge severity={log.severity} />
                        </TableCell>
                        {canSelectHospital && (
                          <TableCell>{log.hospital_name || '-'}</TableCell>
                        )}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>

              {/* Pagination */}
              {meta && meta.total_pages > 1 && (
                <div className="flex items-center justify-between mt-4">
                  <p className="text-sm text-muted-foreground">
                    Pagina {meta.page} de {meta.total_pages}
                  </p>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handlePageChange(meta.page - 1)}
                      disabled={meta.page <= 1}
                    >
                      Anterior
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handlePageChange(meta.page + 1)}
                      disabled={meta.page >= meta.total_pages}
                    >
                      Proximo
                    </Button>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
