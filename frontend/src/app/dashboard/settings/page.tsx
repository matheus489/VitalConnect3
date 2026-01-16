'use client';

import { Settings, Filter, Clock, AlertTriangle } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

// Mock triagem rules
const triagemRules = [
  {
    id: '1',
    nome: 'Idade Maxima',
    descricao: 'Pacientes acima de 80 anos sao considerados inelegiveis',
    ativo: true,
    config: { idade_maxima: 80 },
  },
  {
    id: '2',
    nome: 'Janela de 6 Horas',
    descricao: 'Obitos com mais de 6 horas sao considerados inelegiveis',
    ativo: true,
    config: { janela_horas: 6 },
  },
  {
    id: '3',
    nome: 'Identificacao Desconhecida',
    descricao: 'Pacientes sem identificacao (indigentes) sao inelegiveis',
    ativo: true,
    config: { identificacao_desconhecida_inelegivel: true },
  },
  {
    id: '4',
    nome: 'Causas Excludentes',
    descricao: 'Lista de causas de morte que tornam o paciente inelegivel',
    ativo: true,
    config: { causas_excludentes: ['Septicemia', 'AIDS', 'Hepatite'] },
  },
];

export default function SettingsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Configuracoes</h1>
        <p className="text-muted-foreground">
          Regras de triagem e configuracoes do sistema
        </p>
      </div>

      {/* Triagem Rules */}
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <Filter className="h-5 w-5 text-primary" />
          <h2 className="text-lg font-semibold">Regras de Triagem</h2>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          {triagemRules.map((rule) => (
            <Card key={rule.id}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base">{rule.nome}</CardTitle>
                  <Badge variant={rule.ativo ? 'default' : 'secondary'}>
                    {rule.ativo ? 'Ativa' : 'Inativa'}
                  </Badge>
                </div>
                <CardDescription>{rule.descricao}</CardDescription>
              </CardHeader>
              <CardContent>
                <pre className="text-xs bg-muted p-2 rounded overflow-x-auto">
                  {JSON.stringify(rule.config, null, 2)}
                </pre>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>

      {/* System Info */}
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <Settings className="h-5 w-5 text-primary" />
          <h2 className="text-lg font-semibold">Informacoes do Sistema</h2>
        </div>

        <div className="grid gap-4 md:grid-cols-3">
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center gap-3">
                <Clock className="h-5 w-5 text-primary" />
                <div>
                  <p className="font-medium">Polling Interval</p>
                  <p className="text-sm text-muted-foreground">3-5 segundos</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center gap-3">
                <AlertTriangle className="h-5 w-5 text-primary" />
                <div>
                  <p className="font-medium">Janela Critica</p>
                  <p className="text-sm text-muted-foreground">6 horas</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center gap-3">
                <Settings className="h-5 w-5 text-primary" />
                <div>
                  <p className="font-medium">Retencao de Dados</p>
                  <p className="text-sm text-muted-foreground">5 anos</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Info Card */}
      <Card className="bg-amber-50 border-amber-200">
        <CardContent className="flex items-start gap-3 pt-6">
          <AlertTriangle className="h-5 w-5 text-amber-600 mt-0.5" />
          <div className="text-sm">
            <p className="font-medium text-amber-800">Ambiente de Demonstracao</p>
            <p className="text-amber-700 mt-1">
              As configuracoes exibidas sao para fins de demonstracao. Em producao,
              as regras de triagem podem ser editadas diretamente via API ou interface administrativa.
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
