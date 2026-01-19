'use client';

import Link from 'next/link';
import { useState, useEffect } from 'react';
import {
  ArrowLeft,
  Eye,
  Clock,
  Bell,
  Shield,
  Zap,
  Brain,
  Heart,
  Activity,
  CheckCircle2,
  TrendingUp,
  Users,
  Building2,
  Database,
  Lock,
  FileCheck,
  Timer,
  AlertTriangle,
  Workflow,
  Server,
  Globe,
  ChevronRight,
  Sparkles,
  Target,
  Lightbulb,
  LineChart,
  Layers,
  ArrowRight,
  Play,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { cn } from '@/lib/utils';

// Counter animation hook
function useCounter(end: number, duration: number = 2000) {
  const [count, setCount] = useState(0);
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    if (!isVisible) return;

    let startTime: number;
    let animationFrame: number;

    const animate = (timestamp: number) => {
      if (!startTime) startTime = timestamp;
      const progress = Math.min((timestamp - startTime) / duration, 1);
      setCount(Math.floor(progress * end));

      if (progress < 1) {
        animationFrame = requestAnimationFrame(animate);
      }
    };

    animationFrame = requestAnimationFrame(animate);
    return () => cancelAnimationFrame(animationFrame);
  }, [end, duration, isVisible]);

  return { count, setIsVisible };
}

// Stat Card Component
function StatCard({ value, label, suffix = '', icon: Icon }: { value: number; label: string; suffix?: string; icon: React.ElementType }) {
  const { count, setIsVisible } = useCounter(value);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => { if (entry.isIntersecting) setIsVisible(true); },
      { threshold: 0.1 }
    );
    const element = document.getElementById(`stat-${label.replace(/\s/g, '-')}`);
    if (element) observer.observe(element);
    return () => observer.disconnect();
  }, [label, setIsVisible]);

  return (
    <div id={`stat-${label.replace(/\s/g, '-')}`} className="text-center p-6 rounded-2xl bg-white/80 backdrop-blur-sm shadow-lg border border-sky-100 hover:shadow-xl transition-all duration-300 hover:-translate-y-1">
      <Icon className="h-8 w-8 mx-auto mb-3 text-primary" />
      <div className="text-4xl font-bold text-primary mb-1">
        {count}{suffix}
      </div>
      <div className="text-sm text-muted-foreground font-medium">{label}</div>
    </div>
  );
}

// Pipeline Step Component
function PipelineStep({ step, title, description, icon: Icon, isLast = false }: { step: number; title: string; description: string; icon: React.ElementType; isLast?: boolean }) {
  return (
    <div className="flex items-start gap-4 group">
      <div className="flex flex-col items-center">
        <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-primary text-white font-bold shadow-lg group-hover:scale-110 transition-transform">
          {step}
        </div>
        {!isLast && <div className="w-0.5 h-full min-h-[60px] bg-gradient-to-b from-primary to-primary/20 mt-2" />}
      </div>
      <div className="pb-8">
        <div className="flex items-center gap-2 mb-1">
          <Icon className="h-5 w-5 text-primary" />
          <h4 className="font-semibold text-lg">{title}</h4>
        </div>
        <p className="text-muted-foreground">{description}</p>
      </div>
    </div>
  );
}

// TRL Indicator Component
function TRLIndicator() {
  const levels = [
    { level: 1, name: 'Pesquisa Basica', completed: true },
    { level: 2, name: 'Conceito Formulado', completed: true },
    { level: 3, name: 'Prova de Conceito', completed: true },
    { level: 4, name: 'Validacao em Lab', completed: true },
    { level: 5, name: 'Validacao Ambiente Relevante', completed: true },
    { level: 6, name: 'Demonstracao Ambiente Relevante', completed: true },
    { level: 7, name: 'Demonstracao Ambiente Operacional', completed: true },
    { level: 8, name: 'Sistema Completo Qualificado', completed: false, current: true },
    { level: 9, name: 'Sistema Operacional', completed: false },
  ];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm font-medium">Nivel de Maturidade Tecnologica (TRL)</span>
        <Badge variant="default" className="bg-primary">TRL 7-8</Badge>
      </div>
      <Progress value={78} className="h-3" />
      <div className="grid grid-cols-9 gap-1 mt-4">
        {levels.map((l) => (
          <div key={l.level} className="text-center">
            <div className={cn(
              "w-full h-8 rounded-lg flex items-center justify-center text-xs font-bold transition-all",
              l.completed ? "bg-primary text-white" : l.current ? "bg-primary/60 text-white animate-pulse" : "bg-gray-200 text-gray-500"
            )}>
              {l.level}
            </div>
          </div>
        ))}
      </div>
      <p className="text-sm text-muted-foreground text-center mt-2">
        Sistema em estagio de <strong>demonstracao em ambiente operacional</strong>, pronto para piloto em producao
      </p>
    </div>
  );
}

export default function AboutPage() {
  const [activeTab, setActiveTab] = useState('problema');

  return (
    <div className="min-h-screen bg-gradient-to-b from-sky-50 via-white to-sky-50">
      {/* Hero Section */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-primary/5 via-transparent to-sky-100/50" />
        <div className="absolute top-0 right-0 w-96 h-96 bg-primary/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2" />
        <div className="absolute bottom-0 left-0 w-96 h-96 bg-sky-200/30 rounded-full blur-3xl translate-y-1/2 -translate-x-1/2" />

        <div className="container mx-auto px-4 py-8 max-w-6xl relative">
          <Button asChild variant="ghost" className="mb-8 hover:bg-white/50">
            <Link href="/">
              <ArrowLeft className="mr-2 h-4 w-4" />
              Voltar ao Sistema
            </Link>
          </Button>

          <div className="text-center py-12 md:py-20">
            <Badge variant="secondary" className="mb-6 px-4 py-2 text-sm">
              <Sparkles className="h-4 w-4 mr-2" />
              Solucao Inovadora - LC 182/2021
            </Badge>

            <div className="flex items-center justify-center gap-3 mb-6">
              <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-primary shadow-xl shadow-primary/25">
                <Heart className="h-8 w-8 text-white" />
              </div>
              <h1 className="text-4xl md:text-6xl font-bold bg-gradient-to-r from-primary via-primary to-sky-600 bg-clip-text text-transparent">
                VitalConnect
              </h1>
            </div>

            <p className="text-xl md:text-2xl text-muted-foreground max-w-3xl mx-auto mb-4">
              Sistema Inteligente de Captacao de Corneas para
              <span className="text-primary font-semibold"> Centrais de Transplantes</span> e
              <span className="text-primary font-semibold"> Bancos de Olhos</span>
            </p>

            <p className="text-lg text-muted-foreground max-w-2xl mx-auto mb-8">
              Transformando o processo de notificacao e captacao de corneas atraves de
              inteligencia artificial e automacao em tempo real
            </p>

            <div className="flex flex-wrap justify-center gap-4">
              <Badge variant="outline" className="px-4 py-2 text-sm border-primary/30">
                <Building2 className="h-4 w-4 mr-2" />
                SES-GO
              </Badge>
              <Badge variant="outline" className="px-4 py-2 text-sm border-primary/30">
                <Activity className="h-4 w-4 mr-2" />
                Central de Transplantes
              </Badge>
              <Badge variant="outline" className="px-4 py-2 text-sm border-primary/30">
                <Eye className="h-4 w-4 mr-2" />
                Banco de Olhos
              </Badge>
            </div>
          </div>
        </div>
      </section>

      {/* Stats Section */}
      <section className="py-12 bg-gradient-to-r from-primary/5 via-sky-50 to-primary/5">
        <div className="container mx-auto px-4 max-w-6xl">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 md:gap-6">
            <StatCard value={6} suffix="h" label="Janela Critica" icon={Timer} />
            <StatCard value={70} suffix="%" label="Corneas Perdidas" icon={AlertTriangle} />
            <StatCard value={3} suffix="x" label="Aumento Captacao" icon={TrendingUp} />
            <StatCard value={24} suffix="/7" label="Monitoramento" icon={Activity} />
          </div>
        </div>
      </section>

      {/* Main Content */}
      <section className="py-16">
        <div className="container mx-auto px-4 max-w-6xl">

          {/* Problem Section */}
          <Card className="mb-12 overflow-hidden border-none shadow-xl">
            <div className="bg-gradient-to-r from-red-500/10 to-orange-500/10 p-1">
              <CardHeader className="bg-white rounded-t-lg">
                <div className="flex items-center gap-3">
                  <div className="p-3 bg-red-100 rounded-xl">
                    <AlertTriangle className="h-6 w-6 text-red-600" />
                  </div>
                  <div>
                    <CardTitle className="text-2xl">O Problema: Corneas que Salvam Vidas Estao Sendo Perdidas</CardTitle>
                    <CardDescription>Um desafio critico de saude publica que afeta milhares de pacientes</CardDescription>
                  </div>
                </div>
              </CardHeader>
            </div>
            <CardContent className="pt-6 space-y-6">
              <div className="grid md:grid-cols-2 gap-8">
                <div>
                  <p className="text-lg text-muted-foreground mb-4">
                    A captacao de corneas para transplante enfrenta um desafio critico: existe uma
                    <strong className="text-foreground"> janela de apenas 6 horas</strong> apos o obito
                    para que a equipe de captacao seja notificada e realize o procedimento.
                  </p>
                  <p className="text-muted-foreground mb-4">
                    Atualmente, <strong className="text-red-600">ate 70% das corneas potencialmente viaveis sao perdidas</strong> devido a falhas no processo de notificacao.
                  </p>
                </div>
                <div className="space-y-3">
                  <h4 className="font-semibold flex items-center gap-2">
                    <Target className="h-5 w-5 text-red-500" />
                    Causas Identificadas:
                  </h4>
                  <ul className="space-y-2">
                    {[
                      'Comunicacao manual e lenta entre hospitais e centrais',
                      'Falta de integracao entre sistemas hospitalares',
                      'Ausencia de alertas automaticos para equipes de plantao',
                      'Dificuldade em rastrear janela critica de 6 horas',
                      'Processos burocraticos que consomem tempo precioso',
                    ].map((item, i) => (
                      <li key={i} className="flex items-start gap-2 text-muted-foreground">
                        <ChevronRight className="h-4 w-4 mt-1 text-red-400 shrink-0" />
                        {item}
                      </li>
                    ))}
                  </ul>
                </div>
              </div>

              <div className="bg-red-50 border border-red-100 rounded-xl p-6 mt-6">
                <div className="flex items-start gap-4">
                  <Heart className="h-8 w-8 text-red-500 shrink-0" />
                  <div>
                    <h4 className="font-semibold text-red-800 mb-1">Impacto Humano</h4>
                    <p className="text-red-700">
                      Cada cornea perdida representa uma pessoa que permanece na fila de transplante,
                      aguardando por uma segunda chance de enxergar. No Brasil, mais de
                      <strong> 20.000 pessoas</strong> aguardam por um transplante de cornea.
                    </p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Solution Section */}
          <Card className="mb-12 overflow-hidden border-none shadow-xl">
            <div className="bg-gradient-to-r from-primary/10 to-sky-500/10 p-1">
              <CardHeader className="bg-white rounded-t-lg">
                <div className="flex items-center gap-3">
                  <div className="p-3 bg-primary/10 rounded-xl">
                    <Lightbulb className="h-6 w-6 text-primary" />
                  </div>
                  <div>
                    <CardTitle className="text-2xl">A Solucao: VitalConnect</CardTitle>
                    <CardDescription>Inovacao tecnologica a servico da vida</CardDescription>
                  </div>
                </div>
              </CardHeader>
            </div>
            <CardContent className="pt-6">
              <p className="text-lg text-muted-foreground mb-8">
                O VitalConnect e uma plataforma inteligente que <strong className="text-primary">automatiza todo o processo
                de deteccao, notificacao e gestao</strong> de potenciais doadores de corneas, garantindo que
                nenhuma oportunidade de salvar vidas seja perdida.
              </p>

              <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
                {[
                  {
                    icon: Eye,
                    title: 'Deteccao Automatica',
                    description: 'Monitoramento continuo 24/7 dos sistemas hospitalares via integracao HL7 FHIR',
                  },
                  {
                    icon: Brain,
                    title: 'IA para Triagem',
                    description: 'Algoritmos inteligentes aplicam criterios de elegibilidade automaticamente',
                  },
                  {
                    icon: Bell,
                    title: 'Alertas em Tempo Real',
                    description: 'Notificacoes instantaneas via dashboard, push e WhatsApp para equipes',
                  },
                  {
                    icon: Clock,
                    title: 'Gestao de Tempo',
                    description: 'Contagem regressiva visual da janela critica de 6 horas',
                  },
                ].map((feature, i) => (
                  <div key={i} className="group p-6 rounded-2xl bg-gradient-to-br from-white to-sky-50 border border-sky-100 hover:shadow-lg transition-all duration-300 hover:-translate-y-1">
                    <div className="h-12 w-12 rounded-xl bg-primary/10 flex items-center justify-center mb-4 group-hover:bg-primary group-hover:text-white transition-colors">
                      <feature.icon className="h-6 w-6 text-primary group-hover:text-white" />
                    </div>
                    <h3 className="font-semibold mb-2">{feature.title}</h3>
                    <p className="text-sm text-muted-foreground">{feature.description}</p>
                  </div>
                ))}
              </div>

              {/* Pipeline Visual */}
              <div className="bg-gradient-to-br from-sky-50 to-white rounded-2xl p-8 border border-sky-100">
                <h3 className="text-xl font-semibold mb-6 flex items-center gap-2">
                  <Workflow className="h-5 w-5 text-primary" />
                  Fluxo de Funcionamento
                </h3>
                <div className="space-y-2">
                  <PipelineStep
                    step={1}
                    title="Monitoramento Continuo"
                    description="O sistema monitora em tempo real os registros de obitos nos hospitais integrados via listener service"
                    icon={Activity}
                  />
                  <PipelineStep
                    step={2}
                    title="Deteccao Automatica"
                    description="Ao detectar um obito, o sistema automaticamente coleta dados demograficos e clinicos do paciente"
                    icon={Database}
                  />
                  <PipelineStep
                    step={3}
                    title="Triagem Inteligente"
                    description="Criterios de elegibilidade configuraveis sao aplicados automaticamente, filtrando potenciais doadores"
                    icon={Shield}
                  />
                  <PipelineStep
                    step={4}
                    title="Notificacao Instantanea"
                    description="A equipe de plantao recebe alerta imediato com todos os dados necessarios para acao rapida"
                    icon={Bell}
                  />
                  <PipelineStep
                    step={5}
                    title="Gestao e Acompanhamento"
                    description="Dashboard em tempo real permite gerenciar casos, atualizar status e gerar relatorios"
                    icon={LineChart}
                    isLast
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Innovation Section - LC 182/2021 */}
          <Card className="mb-12 overflow-hidden border-none shadow-xl bg-gradient-to-br from-purple-50 to-white">
            <CardHeader>
              <div className="flex items-center gap-3">
                <div className="p-3 bg-purple-100 rounded-xl">
                  <Sparkles className="h-6 w-6 text-purple-600" />
                </div>
                <div>
                  <CardTitle className="text-2xl">Inovacao Conforme LC 182/2021</CardTitle>
                  <CardDescription>Enquadramento como Solucao Inovadora para Administracao Publica</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid md:grid-cols-2 gap-6">
                <div className="space-y-4">
                  <h4 className="font-semibold text-purple-800">Caracteristicas Inovadoras:</h4>
                  <ul className="space-y-3">
                    {[
                      'Primeira solucao integrada de monitoramento automatico para captacao de corneas no Brasil',
                      'Uso de IA para triagem automatica de elegibilidade',
                      'Sistema de alertas em tempo real com contagem regressiva da janela critica',
                      'Arquitetura multi-tenant para gestao centralizada de multiplos hospitais',
                      'Assistente virtual com IA generativa para suporte a decisao clinica',
                    ].map((item, i) => (
                      <li key={i} className="flex items-start gap-2">
                        <CheckCircle2 className="h-5 w-5 text-purple-500 shrink-0 mt-0.5" />
                        <span className="text-muted-foreground">{item}</span>
                      </li>
                    ))}
                  </ul>
                </div>
                <div className="space-y-4">
                  <h4 className="font-semibold text-purple-800">Diferencial Competitivo:</h4>
                  <ul className="space-y-3">
                    {[
                      'Solucao 100% web, sem necessidade de instalacao',
                      'Integracao nativa com padrao HL7 FHIR',
                      'Desenvolvimento especifico para realidade do SUS',
                      'Codigo aberto e adaptavel as necessidades locais',
                      'Suporte a multiplos metodos de notificacao (push, email, WhatsApp)',
                    ].map((item, i) => (
                      <li key={i} className="flex items-start gap-2">
                        <CheckCircle2 className="h-5 w-5 text-purple-500 shrink-0 mt-0.5" />
                        <span className="text-muted-foreground">{item}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* TRL Section */}
          <Card className="mb-12 overflow-hidden border-none shadow-xl">
            <CardHeader>
              <div className="flex items-center gap-3">
                <div className="p-3 bg-green-100 rounded-xl">
                  <Layers className="h-6 w-6 text-green-600" />
                </div>
                <div>
                  <CardTitle className="text-2xl">Grau de Desenvolvimento (TRL)</CardTitle>
                  <CardDescription>Technology Readiness Level - Nivel de Maturidade Tecnologica</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-6">
              <TRLIndicator />

              <div className="grid md:grid-cols-3 gap-4 mt-8">
                {[
                  {
                    title: 'Desenvolvido',
                    items: ['Backend API completo em Go', 'Frontend responsivo em React/Next.js', 'Sistema de autenticacao JWT', 'Modulo de notificacoes'],
                    color: 'green',
                  },
                  {
                    title: 'Em Validacao',
                    items: ['Integracao HL7 FHIR', 'Assistente IA', 'Relatorios avancados', 'App mobile PWA'],
                    color: 'yellow',
                  },
                  {
                    title: 'Roadmap',
                    items: ['Machine Learning preditivo', 'Integracao SNT', 'Blockchain para rastreabilidade', 'Analytics avancado'],
                    color: 'blue',
                  },
                ].map((col, i) => (
                  <div key={i} className={cn(
                    "p-5 rounded-xl border",
                    col.color === 'green' && "bg-green-50 border-green-200",
                    col.color === 'yellow' && "bg-yellow-50 border-yellow-200",
                    col.color === 'blue' && "bg-blue-50 border-blue-200",
                  )}>
                    <h4 className={cn(
                      "font-semibold mb-3",
                      col.color === 'green' && "text-green-800",
                      col.color === 'yellow' && "text-yellow-800",
                      col.color === 'blue' && "text-blue-800",
                    )}>{col.title}</h4>
                    <ul className="space-y-2">
                      {col.items.map((item, j) => (
                        <li key={j} className="text-sm text-muted-foreground flex items-center gap-2">
                          <div className={cn(
                            "h-2 w-2 rounded-full",
                            col.color === 'green' && "bg-green-500",
                            col.color === 'yellow' && "bg-yellow-500",
                            col.color === 'blue' && "bg-blue-500",
                          )} />
                          {item}
                        </li>
                      ))}
                    </ul>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          {/* Integration Section */}
          <Card className="mb-12 overflow-hidden border-none shadow-xl">
            <CardHeader>
              <div className="flex items-center gap-3">
                <div className="p-3 bg-sky-100 rounded-xl">
                  <Globe className="h-6 w-6 text-sky-600" />
                </div>
                <div>
                  <CardTitle className="text-2xl">Integracoes e Interoperabilidade</CardTitle>
                  <CardDescription>Compatibilidade com sistemas existentes da rede de saude</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid md:grid-cols-2 gap-8">
                <div>
                  <h4 className="font-semibold mb-4 flex items-center gap-2">
                    <Server className="h-5 w-5 text-sky-500" />
                    Sistemas Integraveis
                  </h4>
                  <div className="space-y-3">
                    {[
                      { name: 'Prontuario Eletronico (PEP)', status: 'Compativel' },
                      { name: 'Sistema de Gestao Hospitalar', status: 'Compativel' },
                      { name: 'HL7 FHIR', status: 'Nativo' },
                      { name: 'e-SUS', status: 'Em desenvolvimento' },
                      { name: 'SNT - Sistema Nacional de Transplantes', status: 'Planejado' },
                    ].map((sys, i) => (
                      <div key={i} className="flex items-center justify-between p-3 bg-sky-50 rounded-lg">
                        <span className="font-medium">{sys.name}</span>
                        <Badge variant={sys.status === 'Nativo' ? 'default' : sys.status === 'Compativel' ? 'secondary' : 'outline'}>
                          {sys.status}
                        </Badge>
                      </div>
                    ))}
                  </div>
                </div>
                <div>
                  <h4 className="font-semibold mb-4 flex items-center gap-2">
                    <Zap className="h-5 w-5 text-sky-500" />
                    Capacidades de Integracao
                  </h4>
                  <ul className="space-y-3">
                    {[
                      'API RESTful documentada com OpenAPI/Swagger',
                      'Webhooks para eventos em tempo real',
                      'Suporte a mensageria HL7 v2.x e FHIR R4',
                      'Conectores para bancos de dados hospitalares',
                      'Exportacao de dados em formatos padrao (CSV, JSON, XML)',
                    ].map((cap, i) => (
                      <li key={i} className="flex items-start gap-2 text-muted-foreground">
                        <CheckCircle2 className="h-4 w-4 text-sky-500 shrink-0 mt-1" />
                        {cap}
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Economic Viability Section */}
          <Card className="mb-12 overflow-hidden border-none shadow-xl bg-gradient-to-br from-emerald-50 to-white">
            <CardHeader>
              <div className="flex items-center gap-3">
                <div className="p-3 bg-emerald-100 rounded-xl">
                  <TrendingUp className="h-6 w-6 text-emerald-600" />
                </div>
                <div>
                  <CardTitle className="text-2xl">Viabilidade Economica e Custo-Beneficio</CardTitle>
                  <CardDescription>Analise de retorno sobre investimento e sustentabilidade</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid md:grid-cols-3 gap-6">
                {[
                  {
                    title: 'Reducao de Custos',
                    value: 'R$ 15.000',
                    description: 'Economia por transplante viabilizado vs. tratamento continuo',
                    icon: LineChart,
                  },
                  {
                    title: 'ROI Estimado',
                    value: '300%',
                    description: 'Retorno sobre investimento no primeiro ano',
                    icon: TrendingUp,
                  },
                  {
                    title: 'Payback',
                    value: '4 meses',
                    description: 'Tempo estimado para retorno do investimento',
                    icon: Timer,
                  },
                ].map((metric, i) => (
                  <div key={i} className="p-6 bg-white rounded-xl border border-emerald-100 text-center">
                    <metric.icon className="h-8 w-8 mx-auto mb-3 text-emerald-500" />
                    <div className="text-3xl font-bold text-emerald-600 mb-1">{metric.value}</div>
                    <div className="font-medium mb-2">{metric.title}</div>
                    <p className="text-sm text-muted-foreground">{metric.description}</p>
                  </div>
                ))}
              </div>

              <div className="bg-emerald-50 border border-emerald-100 rounded-xl p-6">
                <h4 className="font-semibold text-emerald-800 mb-4">Modelo de Negocio Sustentavel</h4>
                <div className="grid md:grid-cols-2 gap-6">
                  <div>
                    <h5 className="font-medium mb-2">Custos Operacionais Baixos:</h5>
                    <ul className="space-y-2 text-sm text-muted-foreground">
                      <li className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-emerald-500" />
                        Infraestrutura em nuvem escalavel (pay-as-you-go)
                      </li>
                      <li className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-emerald-500" />
                        Sem necessidade de hardware dedicado
                      </li>
                      <li className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-emerald-500" />
                        Manutencao automatizada e atualizacoes continuas
                      </li>
                    </ul>
                  </div>
                  <div>
                    <h5 className="font-medium mb-2">Escalabilidade:</h5>
                    <ul className="space-y-2 text-sm text-muted-foreground">
                      <li className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-emerald-500" />
                        Arquitetura multi-tenant para multiplos hospitais
                      </li>
                      <li className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-emerald-500" />
                        Expansao sem custos proporcionais
                      </li>
                      <li className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-emerald-500" />
                        Replicavel para outros estados
                      </li>
                    </ul>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Security Section */}
          <Card className="mb-12 overflow-hidden border-none shadow-xl">
            <CardHeader>
              <div className="flex items-center gap-3">
                <div className="p-3 bg-slate-100 rounded-xl">
                  <Lock className="h-6 w-6 text-slate-600" />
                </div>
                <div>
                  <CardTitle className="text-2xl">Seguranca e Conformidade</CardTitle>
                  <CardDescription>Protecao de dados e adequacao a LGPD</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid md:grid-cols-2 gap-8">
                <div className="space-y-4">
                  <h4 className="font-semibold">Seguranca de Dados:</h4>
                  <ul className="space-y-3">
                    {[
                      'Criptografia AES-256 para dados em repouso',
                      'TLS 1.3 para dados em transito',
                      'Autenticacao JWT com refresh tokens',
                      'Controle de acesso baseado em papeis (RBAC)',
                      'Logs de auditoria completos',
                    ].map((item, i) => (
                      <li key={i} className="flex items-center gap-2 text-muted-foreground">
                        <Shield className="h-4 w-4 text-slate-500" />
                        {item}
                      </li>
                    ))}
                  </ul>
                </div>
                <div className="space-y-4">
                  <h4 className="font-semibold">Conformidade LGPD:</h4>
                  <ul className="space-y-3">
                    {[
                      'Minimizacao de coleta de dados pessoais',
                      'Anonimizacao de dados para relatorios',
                      'Retencao de dados configuravel',
                      'Direito ao esquecimento implementado',
                      'Consentimento explicito para processamento',
                    ].map((item, i) => (
                      <li key={i} className="flex items-center gap-2 text-muted-foreground">
                        <FileCheck className="h-4 w-4 text-slate-500" />
                        {item}
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Technical Stack */}
          <Card className="mb-12 overflow-hidden border-none shadow-xl">
            <CardHeader>
              <div className="flex items-center gap-3">
                <div className="p-3 bg-indigo-100 rounded-xl">
                  <Server className="h-6 w-6 text-indigo-600" />
                </div>
                <div>
                  <CardTitle className="text-2xl">Arquitetura Tecnologica</CardTitle>
                  <CardDescription>Stack moderno, escalavel e de alta performance</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid md:grid-cols-3 gap-6">
                <div className="p-5 bg-indigo-50 rounded-xl">
                  <h4 className="font-semibold mb-4 text-indigo-800">Backend</h4>
                  <ul className="space-y-2 text-sm">
                    {[
                      { name: 'Go (Golang)', desc: 'Alta performance' },
                      { name: 'Gin Framework', desc: 'API RESTful' },
                      { name: 'PostgreSQL 15+', desc: 'Banco relacional' },
                      { name: 'Redis 7+', desc: 'Cache e Pub/Sub' },
                      { name: 'JWT', desc: 'Autenticacao segura' },
                    ].map((tech, i) => (
                      <li key={i} className="flex justify-between">
                        <span className="font-medium">{tech.name}</span>
                        <span className="text-muted-foreground">{tech.desc}</span>
                      </li>
                    ))}
                  </ul>
                </div>
                <div className="p-5 bg-sky-50 rounded-xl">
                  <h4 className="font-semibold mb-4 text-sky-800">Frontend</h4>
                  <ul className="space-y-2 text-sm">
                    {[
                      { name: 'Next.js 14+', desc: 'App Router' },
                      { name: 'React 18+', desc: 'UI reativa' },
                      { name: 'TypeScript', desc: 'Type safety' },
                      { name: 'Tailwind CSS', desc: 'Estilizacao' },
                      { name: 'TanStack Query', desc: 'Data fetching' },
                    ].map((tech, i) => (
                      <li key={i} className="flex justify-between">
                        <span className="font-medium">{tech.name}</span>
                        <span className="text-muted-foreground">{tech.desc}</span>
                      </li>
                    ))}
                  </ul>
                </div>
                <div className="p-5 bg-purple-50 rounded-xl">
                  <h4 className="font-semibold mb-4 text-purple-800">IA & Servicos</h4>
                  <ul className="space-y-2 text-sm">
                    {[
                      { name: 'Python/FastAPI', desc: 'Servico de IA' },
                      { name: 'OpenAI GPT-4', desc: 'LLM' },
                      { name: 'Docker', desc: 'Containerizacao' },
                      { name: 'GitHub Actions', desc: 'CI/CD' },
                      { name: 'Render/AWS', desc: 'Cloud hosting' },
                    ].map((tech, i) => (
                      <li key={i} className="flex justify-between">
                        <span className="font-medium">{tech.name}</span>
                        <span className="text-muted-foreground">{tech.desc}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Testing Methodology */}
          <Card className="mb-12 overflow-hidden border-none shadow-xl">
            <CardHeader>
              <div className="flex items-center gap-3">
                <div className="p-3 bg-orange-100 rounded-xl">
                  <FileCheck className="h-6 w-6 text-orange-600" />
                </div>
                <div>
                  <CardTitle className="text-2xl">Metodologia de Testes</CardTitle>
                  <CardDescription>Validacao rigorosa para garantia de qualidade</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid md:grid-cols-2 gap-8">
                <div>
                  <h4 className="font-semibold mb-4">Estrategia de Testes:</h4>
                  <ul className="space-y-3">
                    {[
                      'Testes unitarios com cobertura > 80%',
                      'Testes de integracao para APIs',
                      'Testes end-to-end com Playwright',
                      'Testes de carga e performance',
                      'Testes de seguranca (OWASP)',
                    ].map((item, i) => (
                      <li key={i} className="flex items-center gap-2 text-muted-foreground">
                        <CheckCircle2 className="h-4 w-4 text-orange-500" />
                        {item}
                      </li>
                    ))}
                  </ul>
                </div>
                <div>
                  <h4 className="font-semibold mb-4">Plano de Piloto:</h4>
                  <ul className="space-y-3">
                    {[
                      'Fase 1: Hospital piloto (2 meses)',
                      'Fase 2: Expansao regional (3 meses)',
                      'Fase 3: Implantacao estadual (6 meses)',
                      'Monitoramento continuo de metricas',
                      'Feedback iterativo com usuarios',
                    ].map((item, i) => (
                      <li key={i} className="flex items-center gap-2 text-muted-foreground">
                        <ArrowRight className="h-4 w-4 text-orange-500" />
                        {item}
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Team Section */}
          <Card className="mb-12 overflow-hidden border-none shadow-xl bg-gradient-to-br from-primary/5 to-white">
            <CardHeader>
              <div className="flex items-center gap-3">
                <div className="p-3 bg-primary/10 rounded-xl">
                  <Users className="h-6 w-6 text-primary" />
                </div>
                <div>
                  <CardTitle className="text-2xl">Equipe e Parceiros</CardTitle>
                  <CardDescription>Profissionais comprometidos com a inovacao em saude</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid md:grid-cols-2 gap-8">
                <div>
                  <h4 className="font-semibold mb-4">Proponente:</h4>
                  <div className="p-4 bg-white rounded-xl border">
                    <p className="font-medium">Secretaria de Estado da Saude de Goias</p>
                    <p className="text-sm text-muted-foreground">Central Estadual de Transplantes</p>
                    <p className="text-sm text-muted-foreground">Banco de Olhos de Goias</p>
                  </div>
                </div>
                <div>
                  <h4 className="font-semibold mb-4">Competencias:</h4>
                  <ul className="space-y-2">
                    {[
                      'Desenvolvimento de software',
                      'Inteligencia Artificial',
                      'Gestao de transplantes',
                      'Integracao de sistemas de saude',
                    ].map((item, i) => (
                      <li key={i} className="flex items-center gap-2 text-muted-foreground">
                        <CheckCircle2 className="h-4 w-4 text-primary" />
                        {item}
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            </CardContent>
          </Card>

        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gradient-to-r from-primary to-sky-700 text-white py-12">
        <div className="container mx-auto px-4 max-w-6xl">
          <div className="grid md:grid-cols-3 gap-8 mb-8">
            <div>
              <div className="flex items-center gap-2 mb-4">
                <Heart className="h-6 w-6" />
                <span className="text-xl font-bold">VitalConnect</span>
              </div>
              <p className="text-white/80 text-sm">
                Transformando a captacao de corneas atraves da tecnologia e salvando vidas.
              </p>
            </div>
            <div>
              <h4 className="font-semibold mb-4">Contato</h4>
              <p className="text-white/80 text-sm">Secretaria de Estado da Saude de Goias</p>
              <p className="text-white/80 text-sm">Central Estadual de Transplantes</p>
              <p className="text-white/80 text-sm">Banco de Olhos de Goias</p>
            </div>
            <div>
              <h4 className="font-semibold mb-4">Edital</h4>
              <p className="text-white/80 text-sm">CPSI - Contrato Publico de Solucao Inovadora</p>
              <p className="text-white/80 text-sm">Edital No 01/2025 - SES/GO</p>
              <p className="text-white/80 text-sm">Lei Complementar No 182/2021</p>
            </div>
          </div>
          <div className="border-t border-white/20 pt-8 text-center">
            <p className="text-white/60 text-sm">
              VitalConnect - Sistema de Captacao de Corneas
            </p>
            <p className="text-white/40 text-sm mt-2">
              Versao 1.0.0 - Desenvolvido para o bem da saude publica
            </p>
          </div>
        </div>
      </footer>
    </div>
  );
}
