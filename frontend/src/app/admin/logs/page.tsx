'use client';

import { useState, useEffect, useCallback } from 'react';
import {
  Search,
  ScrollText,
  Download,
  Filter,
  Calendar,
  X,
  Info,
  AlertTriangle,
  AlertCircle,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  fetchAdminAuditLogs,
  fetchAdminTenants,
  exportAdminAuditLogsCSV,
  type AdminAuditLog,
  type AdminTenant,
  type PaginatedAdminResponse,
} from '@/lib/api/admin';
import { toast } from 'sonner';

// Pagination Component
interface PaginationProps {
  currentPage: number;
  totalPages: number;
  totalItems: number;
  perPage: number;
  onPageChange: (page: number) => void;
}

function Pagination({
  currentPage,
  totalPages,
  totalItems,
  perPage,
  onPageChange,
}: PaginationProps) {
  const startItem = (currentPage - 1) * perPage + 1;
  const endItem = Math.min(currentPage * perPage, totalItems);

  return (
    <div className="flex items-center justify-between px-2 py-4">
      <p className="text-sm text-slate-400">
        Mostrando {startItem} a {endItem} de {totalItems} resultados
      </p>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(currentPage - 1)}
          disabled={currentPage <= 1}
          className="bg-slate-800 border-slate-700 text-slate-300 hover:bg-slate-700 disabled:opacity-50"
        >
          Anterior
        </Button>
        <span className="text-sm text-slate-400">
          Pagina {currentPage} de {totalPages}
        </span>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(currentPage + 1)}
          disabled={currentPage >= totalPages}
          className="bg-slate-800 border-slate-700 text-slate-300 hover:bg-slate-700 disabled:opacity-50"
        >
          Proxima
        </Button>
      </div>
    </div>
  );
}

// Log Details Dialog
interface LogDetailsDialogProps {
  open: boolean;
  onClose: () => void;
  log: AdminAuditLog | null;
}

function LogDetailsDialog({ open, onClose, log }: LogDetailsDialogProps) {
  if (!log) return null;

  const getSeverityIcon = () => {
    switch (log.severity) {
      case 'CRITICAL':
        return <AlertCircle className="h-5 w-5 text-red-400" />;
      case 'WARN':
        return <AlertTriangle className="h-5 w-5 text-amber-400" />;
      default:
        return <Info className="h-5 w-5 text-blue-400" />;
    }
  };

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="bg-slate-800 border-slate-700 text-white sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {getSeverityIcon()}
            Detalhes do Log
          </DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="text-xs text-slate-500">Timestamp</label>
              <p className="text-sm text-white">
                {new Date(log.timestamp).toLocaleString('pt-BR')}
              </p>
            </div>
            <div>
              <label className="text-xs text-slate-500">Severity</label>
              <div className="mt-1">
                <Badge
                  className={
                    log.severity === 'CRITICAL'
                      ? 'bg-red-400/10 text-red-400 border-red-400/20'
                      : log.severity === 'WARN'
                      ? 'bg-amber-400/10 text-amber-400 border-amber-400/20'
                      : 'bg-blue-400/10 text-blue-400 border-blue-400/20'
                  }
                >
                  {log.severity}
                </Badge>
              </div>
            </div>
            <div>
              <label className="text-xs text-slate-500">Usuario</label>
              <p className="text-sm text-white">{log.actor_name || '-'}</p>
            </div>
            <div>
              <label className="text-xs text-slate-500">Tenant</label>
              <p className="text-sm text-white">{log.tenant_name || 'Sistema'}</p>
            </div>
            <div>
              <label className="text-xs text-slate-500">Acao</label>
              <p className="text-sm text-white">{log.acao}</p>
            </div>
            <div>
              <label className="text-xs text-slate-500">Entidade</label>
              <p className="text-sm text-white">
                {log.entidade_tipo} ({log.entidade_id})
              </p>
            </div>
            {log.ip_address && (
              <div className="col-span-2">
                <label className="text-xs text-slate-500">IP Address</label>
                <p className="text-sm text-white">{log.ip_address}</p>
              </div>
            )}
          </div>

          {log.detalhes && Object.keys(log.detalhes).length > 0 && (
            <div className="space-y-2">
              <label className="text-xs text-slate-500 uppercase tracking-wide">
                Detalhes
              </label>
              <pre className="p-3 bg-slate-900 rounded-lg overflow-x-auto text-xs text-slate-300 font-mono">
                {JSON.stringify(log.detalhes, null, 2)}
              </pre>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

// Active Filters Display
interface ActiveFiltersProps {
  filters: {
    tenant_id?: string;
    severity?: string;
    acao?: string;
    data_inicio?: string;
    data_fim?: string;
  };
  tenants: AdminTenant[];
  onClear: (key: string) => void;
  onClearAll: () => void;
}

function ActiveFilters({ filters, tenants, onClear, onClearAll }: ActiveFiltersProps) {
  const activeFilters: { key: string; label: string; value: string }[] = [];

  if (filters.tenant_id) {
    const tenant = tenants.find((t) => t.id === filters.tenant_id);
    activeFilters.push({
      key: 'tenant_id',
      label: 'Tenant',
      value: tenant?.name || filters.tenant_id,
    });
  }
  if (filters.severity) {
    activeFilters.push({ key: 'severity', label: 'Severity', value: filters.severity });
  }
  if (filters.acao) {
    activeFilters.push({ key: 'acao', label: 'Acao', value: filters.acao });
  }
  if (filters.data_inicio) {
    activeFilters.push({ key: 'data_inicio', label: 'Data Inicio', value: filters.data_inicio });
  }
  if (filters.data_fim) {
    activeFilters.push({ key: 'data_fim', label: 'Data Fim', value: filters.data_fim });
  }

  if (activeFilters.length === 0) return null;

  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="text-xs text-slate-500">Filtros ativos:</span>
      {activeFilters.map((filter) => (
        <Badge
          key={filter.key}
          variant="outline"
          className="bg-slate-700/50 border-slate-600 text-slate-300 gap-1 pr-1"
        >
          <span className="text-slate-500">{filter.label}:</span> {filter.value}
          <button
            onClick={() => onClear(filter.key)}
            className="ml-1 p-0.5 rounded hover:bg-slate-600"
          >
            <X className="h-3 w-3" />
          </button>
        </Badge>
      ))}
      <button
        onClick={onClearAll}
        className="text-xs text-slate-400 hover:text-slate-300 underline"
      >
        Limpar todos
      </button>
    </div>
  );
}

export default function AuditLogsPage() {
  const [logs, setLogs] = useState<AdminAuditLog[]>([]);
  const [tenants, setTenants] = useState<AdminTenant[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isExporting, setIsExporting] = useState(false);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const [perPage] = useState(20);

  // Filters
  const [tenantFilter, setTenantFilter] = useState<string>('');
  const [severityFilter, setSeverityFilter] = useState<string>('');
  const [acaoFilter, setAcaoFilter] = useState<string>('');
  const [dataInicio, setDataInicio] = useState<string>('');
  const [dataFim, setDataFim] = useState<string>('');
  const [showFilters, setShowFilters] = useState(false);

  // Dialog state
  const [selectedLog, setSelectedLog] = useState<AdminAuditLog | null>(null);
  const [detailsDialogOpen, setDetailsDialogOpen] = useState(false);

  // Load tenants on mount
  useEffect(() => {
    async function loadTenants() {
      try {
        const response = await fetchAdminTenants({ per_page: 100 });
        setTenants(response.data);
      } catch (err) {
        console.error('Failed to fetch tenants:', err);
      }
    }
    loadTenants();
  }, []);

  const loadLogs = useCallback(async () => {
    try {
      setIsLoading(true);
      const response: PaginatedAdminResponse<AdminAuditLog> = await fetchAdminAuditLogs({
        page,
        per_page: perPage,
        tenant_id: tenantFilter || undefined,
        severity: severityFilter as 'INFO' | 'WARN' | 'CRITICAL' | undefined,
        acao: acaoFilter || undefined,
        data_inicio: dataInicio || undefined,
        data_fim: dataFim || undefined,
      });
      setLogs(response.data);
      setTotalPages(response.meta.total_pages);
      setTotalItems(response.meta.total);
    } catch (err) {
      console.error('Failed to fetch logs:', err);
      toast.error('Falha ao carregar logs');
      setLogs([]);
      setTotalPages(1);
      setTotalItems(0);
    } finally {
      setIsLoading(false);
    }
  }, [page, perPage, tenantFilter, severityFilter, acaoFilter, dataInicio, dataFim]);

  useEffect(() => {
    loadLogs();
  }, [loadLogs]);

  const handleExportCSV = async () => {
    try {
      setIsExporting(true);
      const blob = await exportAdminAuditLogsCSV({
        tenant_id: tenantFilter || undefined,
        severity: severityFilter as 'INFO' | 'WARN' | 'CRITICAL' | undefined,
        acao: acaoFilter || undefined,
        data_inicio: dataInicio || undefined,
        data_fim: dataFim || undefined,
      });

      // Create download link
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `audit-logs-${new Date().toISOString().split('T')[0]}.csv`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      toast.success('Exportacao concluida');
    } catch (err) {
      console.error('Failed to export logs:', err);
      toast.error('Falha ao exportar logs');
    } finally {
      setIsExporting(false);
    }
  };

  const clearFilter = (key: string) => {
    switch (key) {
      case 'tenant_id':
        setTenantFilter('');
        break;
      case 'severity':
        setSeverityFilter('');
        break;
      case 'acao':
        setAcaoFilter('');
        break;
      case 'data_inicio':
        setDataInicio('');
        break;
      case 'data_fim':
        setDataFim('');
        break;
    }
    setPage(1);
  };

  const clearAllFilters = () => {
    setTenantFilter('');
    setSeverityFilter('');
    setAcaoFilter('');
    setDataInicio('');
    setDataFim('');
    setPage(1);
  };

  const handleViewDetails = (log: AdminAuditLog) => {
    setSelectedLog(log);
    setDetailsDialogOpen(true);
  };

  const getSeverityBadge = (severity: string) => {
    switch (severity) {
      case 'CRITICAL':
        return (
          <Badge className="bg-red-400/10 text-red-400 border-red-400/20">
            <AlertCircle className="h-3 w-3 mr-1" />
            CRITICAL
          </Badge>
        );
      case 'WARN':
        return (
          <Badge className="bg-amber-400/10 text-amber-400 border-amber-400/20">
            <AlertTriangle className="h-3 w-3 mr-1" />
            WARN
          </Badge>
        );
      default:
        return (
          <Badge className="bg-blue-400/10 text-blue-400 border-blue-400/20">
            <Info className="h-3 w-3 mr-1" />
            INFO
          </Badge>
        );
    }
  };

  const filters = {
    tenant_id: tenantFilter,
    severity: severityFilter,
    acao: acaoFilter,
    data_inicio: dataInicio,
    data_fim: dataFim,
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-4">
          <div className="h-12 w-12 rounded-lg bg-slate-800 flex items-center justify-center">
            <ScrollText className="h-6 w-6 text-violet-400" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-white">
              Logs de Auditoria
            </h1>
            <p className="text-slate-400">
              Visualize todas as acoes realizadas no sistema
            </p>
          </div>
        </div>
        <Button
          onClick={handleExportCSV}
          disabled={isExporting}
          className="gap-2 bg-violet-600 hover:bg-violet-700"
        >
          {isExporting ? (
            <>
              <div className="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent" />
              Exportando...
            </>
          ) : (
            <>
              <Download className="h-4 w-4" />
              Exportar CSV
            </>
          )}
        </Button>
      </div>

      {/* Filters */}
      <Card className="bg-slate-800 border-slate-700">
        <CardContent className="pt-6">
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowFilters(!showFilters)}
                className="gap-2 text-slate-400 hover:text-white hover:bg-slate-700"
              >
                <Filter className="h-4 w-4" />
                {showFilters ? 'Ocultar Filtros' : 'Mostrar Filtros'}
              </Button>
              <ActiveFilters
                filters={filters}
                tenants={tenants}
                onClear={clearFilter}
                onClearAll={clearAllFilters}
              />
            </div>

            {showFilters && (
              <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5 pt-4 border-t border-slate-700">
                <div className="space-y-2">
                  <label className="text-xs text-slate-400">Tenant</label>
                  <Select value={tenantFilter} onValueChange={(v) => { setTenantFilter(v); setPage(1); }}>
                    <SelectTrigger className="bg-slate-900 border-slate-700 text-white">
                      <SelectValue placeholder="Todos" />
                    </SelectTrigger>
                    <SelectContent className="bg-slate-800 border-slate-700">
                      <SelectItem value="" className="text-white hover:bg-slate-700">
                        Todos
                      </SelectItem>
                      {tenants.map((tenant) => (
                        <SelectItem
                          key={tenant.id}
                          value={tenant.id}
                          className="text-white hover:bg-slate-700"
                        >
                          {tenant.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <label className="text-xs text-slate-400">Severity</label>
                  <Select value={severityFilter} onValueChange={(v) => { setSeverityFilter(v); setPage(1); }}>
                    <SelectTrigger className="bg-slate-900 border-slate-700 text-white">
                      <SelectValue placeholder="Todos" />
                    </SelectTrigger>
                    <SelectContent className="bg-slate-800 border-slate-700">
                      <SelectItem value="" className="text-white hover:bg-slate-700">
                        Todos
                      </SelectItem>
                      <SelectItem value="INFO" className="text-white hover:bg-slate-700">
                        INFO
                      </SelectItem>
                      <SelectItem value="WARN" className="text-white hover:bg-slate-700">
                        WARN
                      </SelectItem>
                      <SelectItem value="CRITICAL" className="text-white hover:bg-slate-700">
                        CRITICAL
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <label className="text-xs text-slate-400">Acao</label>
                  <Select value={acaoFilter} onValueChange={(v) => { setAcaoFilter(v); setPage(1); }}>
                    <SelectTrigger className="bg-slate-900 border-slate-700 text-white">
                      <SelectValue placeholder="Todas" />
                    </SelectTrigger>
                    <SelectContent className="bg-slate-800 border-slate-700">
                      <SelectItem value="" className="text-white hover:bg-slate-700">
                        Todas
                      </SelectItem>
                      <SelectItem value="CREATE" className="text-white hover:bg-slate-700">
                        CREATE
                      </SelectItem>
                      <SelectItem value="UPDATE" className="text-white hover:bg-slate-700">
                        UPDATE
                      </SelectItem>
                      <SelectItem value="DELETE" className="text-white hover:bg-slate-700">
                        DELETE
                      </SelectItem>
                      <SelectItem value="LOGIN" className="text-white hover:bg-slate-700">
                        LOGIN
                      </SelectItem>
                      <SelectItem value="IMPERSONATE" className="text-white hover:bg-slate-700">
                        IMPERSONATE
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <label className="text-xs text-slate-400">Data Inicio</label>
                  <div className="relative">
                    <Calendar className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
                    <Input
                      type="date"
                      value={dataInicio}
                      onChange={(e) => { setDataInicio(e.target.value); setPage(1); }}
                      className="pl-10 bg-slate-900 border-slate-700 text-white"
                    />
                  </div>
                </div>
                <div className="space-y-2">
                  <label className="text-xs text-slate-400">Data Fim</label>
                  <div className="relative">
                    <Calendar className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
                    <Input
                      type="date"
                      value={dataFim}
                      onChange={(e) => { setDataFim(e.target.value); setPage(1); }}
                      className="pl-10 bg-slate-900 border-slate-700 text-white"
                    />
                  </div>
                </div>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Table */}
      <Card className="bg-slate-800 border-slate-700">
        <CardHeader>
          <CardTitle className="text-white">Logs</CardTitle>
          <CardDescription className="text-slate-400">
            {totalItems} registro(s) encontrado(s)
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="h-6 w-6 animate-spin rounded-full border-2 border-violet-600 border-t-transparent" />
            </div>
          ) : logs.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <ScrollText className="h-12 w-12 text-slate-600 mb-4" />
              <h3 className="text-lg font-medium text-slate-300">
                Nenhum log encontrado
              </h3>
              <p className="text-sm text-slate-500 mt-1">
                Ajuste os filtros para encontrar registros
              </p>
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow className="border-slate-700 hover:bg-slate-800">
                      <TableHead className="text-slate-400">Timestamp</TableHead>
                      <TableHead className="text-slate-400">Tenant</TableHead>
                      <TableHead className="text-slate-400">Usuario</TableHead>
                      <TableHead className="text-slate-400">Acao</TableHead>
                      <TableHead className="text-slate-400">Entidade</TableHead>
                      <TableHead className="text-slate-400">Severity</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {logs.map((log) => (
                      <TableRow
                        key={log.id}
                        onClick={() => handleViewDetails(log)}
                        className="border-slate-700 hover:bg-slate-800/50 cursor-pointer"
                      >
                        <TableCell className="text-slate-300 text-sm">
                          {new Date(log.timestamp).toLocaleString('pt-BR')}
                        </TableCell>
                        <TableCell className="text-slate-300">
                          {log.tenant_name || (
                            <span className="text-slate-500 italic">Sistema</span>
                          )}
                        </TableCell>
                        <TableCell className="text-slate-300">
                          {log.actor_name || '-'}
                        </TableCell>
                        <TableCell>
                          <Badge
                            variant="outline"
                            className="bg-slate-700/50 border-slate-600 text-slate-300"
                          >
                            {log.acao}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-slate-400 text-sm">
                          {log.entidade_tipo}
                        </TableCell>
                        <TableCell>{getSeverityBadge(log.severity)}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
              <Pagination
                currentPage={page}
                totalPages={totalPages}
                totalItems={totalItems}
                perPage={perPage}
                onPageChange={setPage}
              />
            </>
          )}
        </CardContent>
      </Card>

      {/* Details Dialog */}
      <LogDetailsDialog
        open={detailsDialogOpen}
        onClose={() => setDetailsDialogOpen(false)}
        log={selectedLog}
      />
    </div>
  );
}
