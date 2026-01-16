'use client';

import { useState } from 'react';
import { FileText, Download, Calendar, Filter } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
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
import { useHospitals } from '@/hooks';
import { useAuth } from '@/hooks/useAuth';
import { getAccessToken } from '@/lib/api';
import { toast } from 'sonner';

const DESFECHO_OPTIONS = [
  { value: 'Captado', label: 'Captacao Realizada' },
  { value: 'Recusa Familiar', label: 'Familia Recusou' },
  { value: 'Contraindicacao Medica', label: 'Contraindicacao Medica' },
  { value: 'Expirado', label: 'Tempo Excedido' },
];

export default function ReportsPage() {
  const { user } = useAuth();
  const { data: hospitals = [] } = useHospitals();

  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');
  const [hospitalId, setHospitalId] = useState('');
  const [desfechos, setDesfechos] = useState<string[]>([]);
  const [isExporting, setIsExporting] = useState(false);

  const canSelectHospital = user?.role === 'admin';

  const buildQueryString = () => {
    const params = new URLSearchParams();
    if (dateFrom) params.append('date_from', dateFrom);
    if (dateTo) params.append('date_to', dateTo);
    if (hospitalId && hospitalId !== 'all') params.append('hospital_id', hospitalId);
    desfechos.forEach((d) => params.append('desfecho[]', d));
    return params.toString();
  };

  const handleExport = async (format: 'csv' | 'pdf') => {
    setIsExporting(true);
    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
      const queryString = buildQueryString();
      const url = `${apiUrl}/reports/${format}${queryString ? `?${queryString}` : ''}`;

      const token = getAccessToken();
      if (!token) {
        throw new Error('Sessão expirada. Faça login novamente.');
      }

      const response = await fetch(url, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || errorData.details || `Erro ${response.status} ao exportar relatorio`);
      }

      const blob = await response.blob();
      const contentDisposition = response.headers.get('content-disposition');
      let filename = `relatorio.${format}`;
      if (contentDisposition) {
        const match = contentDisposition.match(/filename="(.+)"/);
        if (match) filename = match[1];
      }

      // Create download link
      const downloadUrl = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = downloadUrl;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(downloadUrl);

      toast.success(`Relatorio ${format.toUpperCase()} exportado com sucesso!`);
    } catch (error) {
      console.error('Export error:', error);
      const message = error instanceof Error ? error.message : 'Erro ao exportar relatorio';
      toast.error(message);
    } finally {
      setIsExporting(false);
    }
  };

  const toggleDesfecho = (value: string) => {
    setDesfechos((prev) =>
      prev.includes(value) ? prev.filter((d) => d !== value) : [...prev, value]
    );
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Relatorios</h1>
        <p className="text-muted-foreground">
          Exporte relatorios de ocorrencias em PDF ou CSV
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {/* Filters Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Filter className="h-5 w-5" />
              Filtros do Relatorio
            </CardTitle>
            <CardDescription>
              Configure os filtros para personalizar seu relatorio
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="date-from">Data Inicio</Label>
                <Input
                  id="date-from"
                  type="date"
                  value={dateFrom}
                  onChange={(e) => setDateFrom(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="date-to">Data Fim</Label>
                <Input
                  id="date-to"
                  type="date"
                  value={dateTo}
                  onChange={(e) => setDateTo(e.target.value)}
                />
              </div>
            </div>

            {canSelectHospital && (
              <div className="space-y-2">
                <Label htmlFor="hospital">Hospital</Label>
                <Select value={hospitalId} onValueChange={setHospitalId}>
                  <SelectTrigger id="hospital">
                    <SelectValue placeholder="Todos os hospitais" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">Todos os hospitais</SelectItem>
                    {hospitals.map((hospital) => (
                      <SelectItem key={hospital.id} value={hospital.id}>
                        {hospital.nome}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}

            <div className="space-y-2">
              <Label>Desfecho</Label>
              <div className="flex flex-wrap gap-2">
                {DESFECHO_OPTIONS.map((option) => (
                  <Button
                    key={option.value}
                    variant={desfechos.includes(option.value) ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => toggleDesfecho(option.value)}
                  >
                    {option.label}
                  </Button>
                ))}
              </div>
              <p className="text-xs text-muted-foreground">
                Selecione um ou mais desfechos para filtrar (vazio = todos)
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Export Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Download className="h-5 w-5" />
              Exportar Relatorio
            </CardTitle>
            <CardDescription>
              Escolha o formato do relatorio para download
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4">
              <Button
                onClick={() => handleExport('csv')}
                disabled={isExporting}
                variant="outline"
                className="w-full justify-start h-auto py-4"
              >
                <div className="flex items-center gap-3">
                  <FileText className="h-8 w-8 text-green-600" />
                  <div className="text-left">
                    <p className="font-medium">Exportar CSV</p>
                    <p className="text-sm text-muted-foreground">
                      Planilha compativel com Excel e Google Sheets
                    </p>
                  </div>
                </div>
              </Button>

              <Button
                onClick={() => handleExport('pdf')}
                disabled={isExporting}
                variant="outline"
                className="w-full justify-start h-auto py-4"
              >
                <div className="flex items-center gap-3">
                  <FileText className="h-8 w-8 text-red-600" />
                  <div className="text-left">
                    <p className="font-medium">Exportar PDF</p>
                    <p className="text-sm text-muted-foreground">
                      Documento formatado para impressao
                    </p>
                  </div>
                </div>
              </Button>
            </div>

            {isExporting && (
              <p className="text-sm text-center text-muted-foreground">
                Gerando relatorio...
              </p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Info Card */}
      <Card className="bg-muted/50">
        <CardContent className="flex items-start gap-3 pt-6">
          <Calendar className="h-5 w-5 text-primary mt-0.5" />
          <div className="text-sm">
            <p className="font-medium">Sobre os relatorios</p>
            <ul className="mt-2 space-y-1 text-muted-foreground">
              <li>Os relatorios incluem dados de ocorrencias conforme os filtros selecionados</li>
              <li>CSV: Ideal para analise em planilhas e integracao com outros sistemas</li>
              <li>PDF: Formatado para apresentacao e impressao, inclui metricas resumidas</li>
              <li>Todas as exportacoes sao registradas para conformidade com LGPD</li>
            </ul>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
