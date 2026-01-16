import Link from 'next/link';
import { ArrowLeft, Eye, Clock, Bell, Shield } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export default function AboutPage() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-sky-50 to-white">
      <div className="container mx-auto px-4 py-8 max-w-4xl">
        {/* Back Button */}
        <Button asChild variant="ghost" className="mb-8">
          <Link href="/">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Voltar
          </Link>
        </Button>

        {/* Header */}
        <div className="text-center mb-12">
          <div className="flex items-center justify-center gap-2 mb-4">
            <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary">
              <span className="text-2xl font-bold text-primary-foreground">V</span>
            </div>
            <h1 className="text-3xl font-bold text-primary">VitalConnect</h1>
          </div>
          <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
            Sistema de Captacao de Corneas para Centrais de Transplantes e Bancos de Olhos
          </p>
        </div>

        {/* Problem Statement */}
        <Card className="mb-8">
          <CardHeader>
            <CardTitle>O Problema</CardTitle>
          </CardHeader>
          <CardContent className="prose prose-gray max-w-none">
            <p className="text-muted-foreground">
              A captacao de corneas para transplante enfrenta um desafio critico: existe uma
              <strong className="text-foreground"> janela de apenas 6 horas</strong> apos o obito
              para que a equipe de captacao seja notificada e realize o procedimento.
            </p>
            <p className="text-muted-foreground">
              Atualmente, muitos obitos elegiveis nao sao detectados a tempo devido a:
            </p>
            <ul className="text-muted-foreground">
              <li>Comunicacao manual e lenta entre hospitais e centrais de transplante</li>
              <li>Falta de integracao entre sistemas hospitalares</li>
              <li>Ausencia de alertas automaticos para a equipe de plantao</li>
            </ul>
          </CardContent>
        </Card>

        {/* Solution */}
        <Card className="mb-8">
          <CardHeader>
            <CardTitle>A Solucao VitalConnect</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-6 md:grid-cols-2">
              <div className="flex items-start gap-3">
                <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10">
                  <Eye className="h-5 w-5 text-primary" />
                </div>
                <div>
                  <h3 className="font-semibold">Deteccao Automatica</h3>
                  <p className="text-sm text-muted-foreground">
                    Monitoramento continuo dos sistemas hospitalares para deteccao imediata de obitos
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10">
                  <Shield className="h-5 w-5 text-primary" />
                </div>
                <div>
                  <h3 className="font-semibold">Triagem Inteligente</h3>
                  <p className="text-sm text-muted-foreground">
                    Aplicacao automatica de regras de elegibilidade configuraveis para doacao de corneas
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10">
                  <Bell className="h-5 w-5 text-primary" />
                </div>
                <div>
                  <h3 className="font-semibold">Notificacao em Tempo Real</h3>
                  <p className="text-sm text-muted-foreground">
                    Alertas visuais e sonoros instantaneos para a equipe de captacao via dashboard web
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10">
                  <Clock className="h-5 w-5 text-primary" />
                </div>
                <div>
                  <h3 className="font-semibold">Gestao de Janela Critica</h3>
                  <p className="text-sm text-muted-foreground">
                    Acompanhamento em tempo real do tempo restante para cada ocorrencia elegivel
                  </p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Technical Stack */}
        <Card className="mb-8">
          <CardHeader>
            <CardTitle>Stack Tecnologica</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2">
              <div>
                <h3 className="font-semibold mb-2">Backend</h3>
                <ul className="text-sm text-muted-foreground space-y-1">
                  <li>Go (Golang) com Gin Framework</li>
                  <li>PostgreSQL 15+</li>
                  <li>Redis 7+ (Streams e Pub/Sub)</li>
                  <li>JWT para autenticacao</li>
                </ul>
              </div>
              <div>
                <h3 className="font-semibold mb-2">Frontend</h3>
                <ul className="text-sm text-muted-foreground space-y-1">
                  <li>Next.js 14+ com App Router</li>
                  <li>React 18+ com TypeScript</li>
                  <li>Tailwind CSS + Shadcn/UI</li>
                  <li>TanStack Query para data fetching</li>
                </ul>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Footer */}
        <div className="text-center text-sm text-muted-foreground">
          <p>Secretaria de Estado da Saude de Goias</p>
          <p>Central Estadual de Transplantes / Banco de Olhos</p>
          <p className="mt-4">Versao 1.0.0 - MVP</p>
        </div>
      </div>
    </div>
  );
}
