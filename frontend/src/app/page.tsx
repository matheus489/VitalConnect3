import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

export default function Home() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-b from-sky-50 to-white">
      <main className="flex flex-col items-center gap-8 px-4 py-16 text-center">
        {/* Logo/Title */}
        <div className="space-y-4">
          <h1 className="text-4xl font-bold tracking-tight text-sky-600 sm:text-5xl">
            SIDOT
          </h1>
          <p className="text-xl text-muted-foreground max-w-2xl">
            Sistema Inteligente de Doação de Órgãos e Tecidos
          </p>
        </div>

        {/* Description Card */}
        <Card className="max-w-xl">
          <CardHeader>
            <CardTitle>Central de Transplantes</CardTitle>
            <CardDescription>
              Middleware GovTech para deteccao automatica de obitos e notificacao de
              equipes de captacao em tempo real
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-2 text-sm text-left">
              <div className="flex items-center gap-2">
                <div className="h-2 w-2 rounded-full bg-emerald-500" />
                <span>Deteccao automatica de obitos em hospitais</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="h-2 w-2 rounded-full bg-emerald-500" />
                <span>Triagem inteligente para elegibilidade</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="h-2 w-2 rounded-full bg-emerald-500" />
                <span>Notificacoes em tempo real (janela de 6h)</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="h-2 w-2 rounded-full bg-emerald-500" />
                <span>Dashboard de metricas e gestao de ocorrencias</span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Action Buttons */}
        <div className="flex flex-col gap-4 sm:flex-row">
          <Button asChild size="lg" className="min-w-[160px]">
            <Link href="/login">Acessar Sistema</Link>
          </Button>
          <Button asChild variant="outline" size="lg" className="min-w-[160px]">
            <Link href="/about">Sobre o Projeto</Link>
          </Button>
        </div>

        {/* Footer */}
        <footer className="mt-8 text-sm text-muted-foreground">
          <p>Secretaria de Estado da Saude de Goias</p>
          <p>Central Estadual de Transplantes / Banco de Olhos</p>
        </footer>
      </main>
    </div>
  );
}
