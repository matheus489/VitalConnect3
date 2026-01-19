'use client';

import Link from 'next/link';
import { useState, useEffect, useRef } from 'react';
import { motion, useInView, useScroll, useTransform } from 'framer-motion';
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
  ChevronDown,
  Quote,
  CircleDot,
  Mail,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { cn } from '@/lib/utils';

// Animated Counter Hook
function useAnimatedCounter(end: number, duration: number = 2000) {
  const [count, setCount] = useState(0);
  const [hasStarted, setHasStarted] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const isInView = useInView(ref, { once: true, margin: '-50px' });

  useEffect(() => {
    if (!isInView || hasStarted) return;
    setHasStarted(true);

    let startTime: number;
    let animationFrame: number;

    const animate = (timestamp: number) => {
      if (!startTime) startTime = timestamp;
      const progress = Math.min((timestamp - startTime) / duration, 1);
      const easeOut = 1 - Math.pow(1 - progress, 4);
      setCount(Math.floor(easeOut * end));

      if (progress < 1) {
        animationFrame = requestAnimationFrame(animate);
      }
    };

    animationFrame = requestAnimationFrame(animate);
    return () => cancelAnimationFrame(animationFrame);
  }, [end, duration, isInView, hasStarted]);

  return { count, ref };
}

// Stat Card Component with Animation
function StatCard({ value, label, suffix = '', icon: Icon, delay = 0 }: { value: number; label: string; suffix?: string; icon: React.ElementType; delay?: number }) {
  const { count, ref } = useAnimatedCounter(value);

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, y: 30 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay }}
      viewport={{ once: true }}
      whileHover={{ y: -8, scale: 1.02 }}
      className="text-center p-6 rounded-2xl bg-white/90 backdrop-blur-sm shadow-lg border border-sky-100 cursor-pointer transition-shadow hover:shadow-xl"
    >
      <motion.div
        initial={{ scale: 0 }}
        whileInView={{ scale: 1 }}
        transition={{ duration: 0.5, delay: delay + 0.2, type: 'spring' }}
        viewport={{ once: true }}
      >
        <Icon className="h-8 w-8 mx-auto mb-3 text-primary" />
      </motion.div>
      <div className="text-4xl font-bold text-primary mb-1">
        {count}{suffix}
      </div>
      <div className="text-sm text-muted-foreground font-medium">{label}</div>
    </motion.div>
  );
}

// Pipeline Step Component with Animation
function PipelineStep({ step, title, description, icon: Icon, isLast = false, delay = 0 }: { step: number; title: string; description: string; icon: React.ElementType; isLast?: boolean; delay?: number }) {
  return (
    <motion.div
      initial={{ opacity: 0, x: -30 }}
      whileInView={{ opacity: 1, x: 0 }}
      transition={{ duration: 0.5, delay }}
      viewport={{ once: true }}
      className="flex items-start gap-4 group"
    >
      <div className="flex flex-col items-center">
        <motion.div
          whileHover={{ scale: 1.15, rotate: 5 }}
          className="flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-primary to-sky-600 text-white font-bold shadow-lg shadow-primary/30"
        >
          {step}
        </motion.div>
        {!isLast && (
          <motion.div
            initial={{ height: 0 }}
            whileInView={{ height: '100%' }}
            transition={{ duration: 0.8, delay: delay + 0.3 }}
            viewport={{ once: true }}
            className="w-0.5 min-h-[60px] bg-gradient-to-b from-primary to-primary/20 mt-2"
          />
        )}
      </div>
      <div className="pb-8">
        <div className="flex items-center gap-2 mb-1">
          <Icon className="h-5 w-5 text-primary" />
          <h4 className="font-semibold text-lg">{title}</h4>
        </div>
        <p className="text-muted-foreground">{description}</p>
      </div>
    </motion.div>
  );
}

// TRL Indicator Component with Animation
function TRLIndicator() {
  const levels = [
    { level: 1, completed: true },
    { level: 2, completed: true },
    { level: 3, completed: true },
    { level: 4, completed: true },
    { level: 5, completed: true },
    { level: 6, completed: true },
    { level: 7, completed: true },
    { level: 8, current: true },
    { level: 9, completed: false },
  ];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm font-medium">Nivel de Maturidade Tecnologica (TRL)</span>
        <Badge variant="default" className="bg-gradient-to-r from-primary to-sky-600">TRL 7-8</Badge>
      </div>
      <div className="relative h-3 bg-gray-200 rounded-full overflow-hidden">
        <motion.div
          initial={{ width: 0 }}
          whileInView={{ width: '78%' }}
          transition={{ duration: 1.5, ease: 'easeOut' }}
          viewport={{ once: true }}
          className="absolute inset-y-0 left-0 bg-gradient-to-r from-primary to-sky-600 rounded-full"
        />
      </div>
      <div className="grid grid-cols-9 gap-1 mt-4">
        {levels.map((l, i) => (
          <motion.div
            key={l.level}
            initial={{ scale: 0, opacity: 0 }}
            whileInView={{ scale: 1, opacity: 1 }}
            transition={{ duration: 0.3, delay: i * 0.1 }}
            viewport={{ once: true }}
            className="text-center"
          >
            <div className={cn(
              "w-full h-8 rounded-lg flex items-center justify-center text-xs font-bold transition-all",
              l.completed ? "bg-primary text-white" : l.current ? "bg-primary/60 text-white animate-pulse" : "bg-gray-200 text-gray-500"
            )}>
              {l.level}
            </div>
          </motion.div>
        ))}
      </div>
      <p className="text-sm text-muted-foreground text-center mt-2">
        Sistema em estagio de <strong>demonstracao em ambiente operacional</strong>, pronto para piloto em producao
      </p>
    </div>
  );
}

// Timeline Component with Enhanced Animation
function ProjectTimeline() {
  const timelineItems = [
    {
      phase: 'Fase 1',
      title: 'Concepcao e Pesquisa',
      period: 'Meses 1-2',
      status: 'completed',
      icon: Lightbulb,
      color: 'emerald',
      items: ['Levantamento de requisitos', 'Analise de processos atuais', 'Design da arquitetura'],
      deliverables: ['Documento de Requisitos', 'Arquitetura do Sistema'],
    },
    {
      phase: 'Fase 2',
      title: 'Desenvolvimento MVP',
      period: 'Meses 3-6',
      status: 'completed',
      icon: Server,
      color: 'emerald',
      items: ['Backend API em Go', 'Frontend React/Next.js', 'Sistema de notificacoes'],
      deliverables: ['API Backend', 'Dashboard Web', 'Sistema de Alertas'],
    },
    {
      phase: 'Fase 3',
      title: 'Validacao e Testes',
      period: 'Meses 7-8',
      status: 'current',
      icon: FileCheck,
      color: 'blue',
      items: ['Testes de integracao', 'Validacao com usuarios', 'Ajustes de UX'],
      deliverables: ['Relatorio de Testes', 'Feedback de Usuarios'],
    },
    {
      phase: 'Fase 4',
      title: 'Piloto Hospitalar',
      period: 'Meses 9-12',
      status: 'upcoming',
      icon: Building2,
      color: 'gray',
      items: ['Implantacao hospital piloto', 'Monitoramento de metricas', 'Iteracoes de melhoria'],
      deliverables: ['Sistema em Producao', 'Metricas de Impacto'],
    },
    {
      phase: 'Fase 5',
      title: 'Expansao Estadual',
      period: 'Ano 2',
      status: 'upcoming',
      icon: Globe,
      color: 'gray',
      items: ['Rollout para rede estadual', 'Integracao com SNT', 'Escala nacional'],
      deliverables: ['Rede Estadual Conectada', 'Integracao Nacional'],
    },
  ];

  const completedCount = timelineItems.filter(item => item.status === 'completed').length;
  const currentIndex = timelineItems.findIndex(item => item.status === 'current');
  const progressPercentage = ((completedCount + 0.5) / timelineItems.length) * 100;

  return (
    <div className="space-y-8">
      {/* Progress Overview */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        className="bg-gradient-to-r from-blue-50 to-sky-50 rounded-2xl p-6 border border-blue-100"
      >
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-4">
          <div>
            <h4 className="font-semibold text-lg text-gray-800">Progresso do Projeto</h4>
            <p className="text-sm text-muted-foreground">
              Fase atual: <span className="font-medium text-blue-600">{timelineItems[currentIndex]?.title || 'Concluido'}</span>
            </p>
          </div>
          <div className="flex items-center gap-6">
            <div className="text-center">
              <div className="text-2xl font-bold text-emerald-600">{completedCount}</div>
              <div className="text-xs text-muted-foreground">Concluidas</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600">1</div>
              <div className="text-xs text-muted-foreground">Em Andamento</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-gray-400">{timelineItems.length - completedCount - 1}</div>
              <div className="text-xs text-muted-foreground">Pendentes</div>
            </div>
          </div>
        </div>
        <div className="relative h-3 bg-gray-200 rounded-full overflow-hidden">
          <motion.div
            initial={{ width: 0 }}
            whileInView={{ width: `${progressPercentage}%` }}
            transition={{ duration: 1.5, ease: [0.25, 0.46, 0.45, 0.94] }}
            viewport={{ once: true }}
            className="absolute inset-y-0 left-0 bg-gradient-to-r from-emerald-500 via-blue-500 to-blue-400 rounded-full"
          />
          {/* Progress markers */}
          <div className="absolute inset-0 flex justify-between px-1">
            {timelineItems.map((_, i) => (
              <div key={i} className="w-0.5 h-full bg-white/50" />
            ))}
          </div>
        </div>
        <div className="flex justify-between mt-2 text-xs text-muted-foreground">
          <span>Inicio</span>
          <span>{Math.round(progressPercentage)}% Completo</span>
          <span>Conclusao</span>
        </div>
      </motion.div>

      {/* Timeline */}
      <div className="relative">
        {/* Central Line */}
        <div className="absolute left-6 md:left-1/2 top-0 bottom-0 w-1 bg-gradient-to-b from-emerald-400 via-blue-400 to-gray-200 md:-translate-x-1/2 rounded-full" />

        {/* Animated Progress Line */}
        <motion.div
          initial={{ height: 0 }}
          whileInView={{ height: `${progressPercentage}%` }}
          transition={{ duration: 2, ease: [0.25, 0.46, 0.45, 0.94] }}
          viewport={{ once: true }}
          className="absolute left-6 md:left-1/2 top-0 w-1 bg-gradient-to-b from-emerald-500 to-blue-500 md:-translate-x-1/2 rounded-full z-[1]"
          style={{ boxShadow: '0 0 10px rgba(59, 130, 246, 0.5)' }}
        />

        <div className="space-y-8 md:space-y-12">
          {timelineItems.map((item, i) => {
            const Icon = item.icon;
            const isLeft = i % 2 === 0;

            return (
              <motion.div
                key={i}
                initial={{ opacity: 0, x: isLeft ? -50 : 50 }}
                whileInView={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.6, delay: i * 0.15, ease: [0.25, 0.46, 0.45, 0.94] }}
                viewport={{ once: true, margin: '-50px' }}
                className={cn(
                  "relative flex flex-col md:flex-row items-start gap-4 md:gap-8",
                  isLeft ? "md:flex-row-reverse" : ""
                )}
              >
                {/* Timeline Node */}
                <motion.div
                  initial={{ scale: 0, rotate: -180 }}
                  whileInView={{ scale: 1, rotate: 0 }}
                  transition={{ duration: 0.5, delay: i * 0.15 + 0.2, type: 'spring', stiffness: 200 }}
                  viewport={{ once: true }}
                  className={cn(
                    "absolute left-6 md:left-1/2 -translate-x-1/2 z-10 flex items-center justify-center w-12 h-12 rounded-full border-4 border-white shadow-lg",
                    item.status === 'completed' && "bg-gradient-to-br from-emerald-400 to-emerald-600",
                    item.status === 'current' && "bg-gradient-to-br from-blue-400 to-blue-600",
                    item.status === 'upcoming' && "bg-gradient-to-br from-gray-300 to-gray-400"
                  )}
                >
                  {item.status === 'completed' && (
                    <CheckCircle2 className="h-6 w-6 text-white" />
                  )}
                  {item.status === 'current' && (
                    <motion.div
                      animate={{ scale: [1, 1.2, 1] }}
                      transition={{ duration: 2, repeat: Infinity }}
                    >
                      <Activity className="h-6 w-6 text-white" />
                    </motion.div>
                  )}
                  {item.status === 'upcoming' && (
                    <Clock className="h-6 w-6 text-white" />
                  )}
                </motion.div>

                {/* Content Card */}
                <div className={cn(
                  "ml-20 md:ml-0 md:w-5/12",
                  isLeft ? "md:pr-16" : "md:pl-16"
                )}>
                  <motion.div
                    whileHover={{ y: -5, scale: 1.02 }}
                    transition={{ duration: 0.2 }}
                    className={cn(
                      "relative p-6 rounded-2xl border-2 shadow-lg cursor-pointer transition-shadow hover:shadow-xl overflow-hidden",
                      item.status === 'completed' && "bg-gradient-to-br from-emerald-50 to-white border-emerald-200",
                      item.status === 'current' && "bg-gradient-to-br from-blue-50 to-white border-blue-300 ring-2 ring-blue-200 ring-offset-2",
                      item.status === 'upcoming' && "bg-gradient-to-br from-gray-50 to-white border-gray-200"
                    )}
                  >
                    {/* Status Ribbon */}
                    {item.status === 'current' && (
                      <div className="absolute top-0 right-0">
                        <div className="bg-blue-500 text-white text-xs font-bold px-3 py-1 rounded-bl-lg">
                          EM ANDAMENTO
                        </div>
                      </div>
                    )}

                    {/* Background Icon */}
                    <div className="absolute -right-4 -bottom-4 opacity-5">
                      <Icon className="h-32 w-32" />
                    </div>

                    <div className="relative z-10">
                      {/* Header */}
                      <div className={cn(
                        "flex items-start justify-between mb-4",
                        isLeft ? "md:flex-row-reverse md:text-right" : ""
                      )}>
                        <div className={cn(isLeft && "md:text-right")}>
                          <Badge
                            className={cn(
                              "mb-2 font-bold",
                              item.status === 'completed' && "bg-emerald-100 text-emerald-700 border-emerald-300",
                              item.status === 'current' && "bg-blue-100 text-blue-700 border-blue-300",
                              item.status === 'upcoming' && "bg-gray-100 text-gray-600 border-gray-300"
                            )}
                            variant="outline"
                          >
                            {item.phase}
                          </Badge>
                          <h4 className="text-xl font-bold text-gray-800 mb-1">{item.title}</h4>
                          <div className={cn(
                            "flex items-center gap-2 text-sm text-muted-foreground",
                            isLeft && "md:justify-end"
                          )}>
                            <Timer className="h-4 w-4" />
                            <span>{item.period}</span>
                          </div>
                        </div>
                        <motion.div
                          whileHover={{ rotate: 15, scale: 1.1 }}
                          className={cn(
                            "p-3 rounded-xl",
                            item.status === 'completed' && "bg-emerald-100",
                            item.status === 'current' && "bg-blue-100",
                            item.status === 'upcoming' && "bg-gray-100"
                          )}
                        >
                          <Icon className={cn(
                            "h-6 w-6",
                            item.status === 'completed' && "text-emerald-600",
                            item.status === 'current' && "text-blue-600",
                            item.status === 'upcoming' && "text-gray-500"
                          )} />
                        </motion.div>
                      </div>

                      {/* Tasks */}
                      <div className="mb-4">
                        <h5 className={cn(
                          "text-sm font-semibold mb-2 text-gray-700",
                          isLeft && "md:text-right"
                        )}>Atividades:</h5>
                        <ul className="space-y-2">
                          {item.items.map((task, j) => (
                            <motion.li
                              key={j}
                              initial={{ opacity: 0, x: isLeft ? 20 : -20 }}
                              whileInView={{ opacity: 1, x: 0 }}
                              transition={{ delay: i * 0.15 + j * 0.1 + 0.3 }}
                              viewport={{ once: true }}
                              className={cn(
                                "flex items-center gap-2 text-sm",
                                isLeft && "md:flex-row-reverse"
                              )}
                            >
                              <div className={cn(
                                "h-2 w-2 rounded-full shrink-0",
                                item.status === 'completed' && "bg-emerald-500",
                                item.status === 'current' && "bg-blue-500",
                                item.status === 'upcoming' && "bg-gray-400"
                              )} />
                              <span className="text-muted-foreground">{task}</span>
                            </motion.li>
                          ))}
                        </ul>
                      </div>

                      {/* Deliverables */}
                      <div className={cn(
                        "flex flex-wrap gap-2",
                        isLeft && "md:justify-end"
                      )}>
                        {item.deliverables.map((deliverable, k) => (
                          <motion.span
                            key={k}
                            initial={{ opacity: 0, scale: 0.8 }}
                            whileInView={{ opacity: 1, scale: 1 }}
                            transition={{ delay: i * 0.15 + k * 0.1 + 0.5 }}
                            viewport={{ once: true }}
                            className={cn(
                              "inline-flex items-center gap-1 text-xs px-2 py-1 rounded-full",
                              item.status === 'completed' && "bg-emerald-100 text-emerald-700",
                              item.status === 'current' && "bg-blue-100 text-blue-700",
                              item.status === 'upcoming' && "bg-gray-100 text-gray-600"
                            )}
                          >
                            <CheckCircle2 className="h-3 w-3" />
                            {deliverable}
                          </motion.span>
                        ))}
                      </div>
                    </div>
                  </motion.div>
                </div>

                {/* Empty space for alternating layout */}
                <div className="hidden md:block md:w-5/12" />
              </motion.div>
            );
          })}
        </div>

        {/* End Node */}
        <motion.div
          initial={{ scale: 0 }}
          whileInView={{ scale: 1 }}
          transition={{ duration: 0.5, delay: 0.8, type: 'spring' }}
          viewport={{ once: true }}
          className="absolute left-6 md:left-1/2 -translate-x-1/2 bottom-0 translate-y-1/2 z-10"
        >
          <div className="relative">
            <motion.div
              animate={{ scale: [1, 1.2, 1] }}
              transition={{ duration: 2, repeat: Infinity }}
              className="absolute inset-0 bg-primary/20 rounded-full"
            />
            <div className="relative flex items-center justify-center w-16 h-16 rounded-full bg-gradient-to-br from-primary to-sky-600 border-4 border-white shadow-xl">
              <Target className="h-8 w-8 text-white" />
            </div>
          </div>
        </motion.div>
      </div>

      {/* Legend */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.5 }}
        viewport={{ once: true }}
        className="flex flex-wrap justify-center gap-6 pt-12 mt-8 border-t border-gray-200"
      >
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 rounded-full bg-gradient-to-br from-emerald-400 to-emerald-600 shadow" />
          <span className="text-sm text-muted-foreground">Concluido</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 rounded-full bg-gradient-to-br from-blue-400 to-blue-600 shadow animate-pulse" />
          <span className="text-sm text-muted-foreground">Em Andamento</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 rounded-full bg-gradient-to-br from-gray-300 to-gray-400 shadow" />
          <span className="text-sm text-muted-foreground">Pendente</span>
        </div>
      </motion.div>
    </div>
  );
}

// FAQ Accordion with Animation
function FAQAccordion() {
  const [openIndex, setOpenIndex] = useState<number | null>(0);

  const faqs = [
    {
      question: 'Como o VitalConnect se integra aos sistemas hospitalares existentes?',
      answer: 'O VitalConnect utiliza o padrao HL7 FHIR para integracao, permitindo conexao com a maioria dos sistemas de prontuario eletronico e gestao hospitalar. A integracao e feita via API RESTful ou mensageria, sem necessidade de alteracoes nos sistemas existentes.',
    },
    {
      question: 'Qual e o tempo de implementacao do sistema?',
      answer: 'A implementacao em um hospital piloto leva cerca de 2 meses, incluindo integracao, treinamento e periodo de acompanhamento. A expansao para outros hospitais e mais rapida, cerca de 2-3 semanas por unidade.',
    },
    {
      question: 'O sistema esta em conformidade com a LGPD?',
      answer: 'Sim, o VitalConnect foi desenvolvido com privacidade por design. Utilizamos criptografia AES-256, minimizacao de dados, anonimizacao para relatorios e controles de acesso rigorosos. Todos os processos sao auditaveis.',
    },
    {
      question: 'Como funciona o sistema de alertas em tempo real?',
      answer: 'O sistema monitora continuamente os registros de obito e, ao detectar um potencial doador, envia notificacoes instantaneas via dashboard web, push notifications, email e WhatsApp para a equipe de plantao, com countdown da janela critica de 6 horas.',
    },
    {
      question: 'Qual e o custo de manutencao do sistema?',
      answer: 'O VitalConnect opera em infraestrutura cloud com modelo pay-as-you-go, resultando em custos operacionais baixos. Estimamos um custo mensal de R$ 2.000 a R$ 5.000 dependendo do volume, com suporte e atualizacoes inclusos.',
    },
  ];

  return (
    <div className="space-y-4">
      {faqs.map((faq, i) => (
        <motion.div
          key={i}
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: i * 0.1 }}
          viewport={{ once: true }}
          className="border border-gray-200 rounded-2xl overflow-hidden bg-white shadow-sm hover:shadow-md transition-shadow"
        >
          <button
            onClick={() => setOpenIndex(openIndex === i ? null : i)}
            className="w-full px-6 py-5 text-left flex items-center justify-between hover:bg-gray-50 transition-colors"
          >
            <span className="font-semibold text-gray-900 pr-8">{faq.question}</span>
            <motion.div
              animate={{ rotate: openIndex === i ? 180 : 0 }}
              transition={{ duration: 0.3 }}
            >
              <ChevronDown className="h-5 w-5 text-gray-500" />
            </motion.div>
          </button>
          <motion.div
            initial={false}
            animate={{
              height: openIndex === i ? 'auto' : 0,
              opacity: openIndex === i ? 1 : 0,
            }}
            transition={{ duration: 0.3 }}
            className="overflow-hidden"
          >
            <div className="px-6 pb-5 text-gray-600">
              {faq.answer}
            </div>
          </motion.div>
        </motion.div>
      ))}
    </div>
  );
}

// Testimonial Card with Animation
function TestimonialCard({ quote, author, role, institution, delay = 0 }: { quote: string; author: string; role: string; institution: string; delay?: number }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 30 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay }}
      viewport={{ once: true }}
      whileHover={{ y: -5 }}
      className="p-8 bg-white rounded-3xl shadow-xl border border-gray-100 relative hover:shadow-2xl transition-shadow"
    >
      <Quote className="absolute top-6 left-6 h-8 w-8 text-primary/20" />
      <div className="pt-8">
        <p className="text-gray-700 italic mb-6 text-lg leading-relaxed">"{quote}"</p>
        <div className="flex items-center gap-4">
          <motion.div
            whileHover={{ scale: 1.1 }}
            className="w-12 h-12 bg-gradient-to-br from-primary to-sky-600 rounded-full flex items-center justify-center text-white font-bold shadow-lg"
          >
            {author.charAt(0)}
          </motion.div>
          <div>
            <div className="font-semibold text-gray-900">{author}</div>
            <div className="text-sm text-gray-500">{role}</div>
            <div className="text-sm text-primary">{institution}</div>
          </div>
        </div>
      </div>
    </motion.div>
  );
}

// Feature Card with Hover Animation
function FeatureCard({ icon: Icon, title, description, delay = 0 }: { icon: React.ElementType; title: string; description: string; delay?: number }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 30 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay }}
      viewport={{ once: true }}
      whileHover={{ y: -8, scale: 1.02 }}
      className="group p-6 rounded-2xl bg-gradient-to-br from-white to-sky-50 border border-sky-100 hover:shadow-xl transition-all duration-300 cursor-pointer"
    >
      <motion.div
        whileHover={{ rotate: 10, scale: 1.1 }}
        className="h-12 w-12 rounded-xl bg-primary/10 flex items-center justify-center mb-4 group-hover:bg-primary transition-colors duration-300"
      >
        <Icon className="h-6 w-6 text-primary group-hover:text-white transition-colors duration-300" />
      </motion.div>
      <h3 className="font-semibold mb-2">{title}</h3>
      <p className="text-sm text-muted-foreground">{description}</p>
    </motion.div>
  );
}

// Animated Section Wrapper
function AnimatedSection({ children, className = '', delay = 0 }: { children: React.ReactNode; className?: string; delay?: number }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 40 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6, delay }}
      viewport={{ once: true, margin: '-50px' }}
      className={className}
    >
      {children}
    </motion.div>
  );
}

// Animated Card Wrapper
function AnimatedCard({ children, className = '', delay = 0 }: { children: React.ReactNode; className?: string; delay?: number }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 30 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay }}
      viewport={{ once: true }}
    >
      <Card className={className}>
        {children}
      </Card>
    </motion.div>
  );
}

export default function AboutPage() {
  const { scrollYProgress } = useScroll();
  const heroY = useTransform(scrollYProgress, [0, 0.3], [0, -50]);
  const heroOpacity = useTransform(scrollYProgress, [0, 0.2], [1, 0.8]);

  return (
    <div className="min-h-screen bg-gradient-to-b from-sky-50 via-white to-sky-50">
      {/* Progress Bar */}
      <motion.div
        className="fixed top-0 left-0 right-0 h-1 bg-gradient-to-r from-primary to-sky-600 z-50 origin-left"
        style={{ scaleX: scrollYProgress }}
      />

      {/* Hero Section */}
      <motion.section
        style={{ y: heroY, opacity: heroOpacity }}
        className="relative overflow-hidden"
      >
        <div className="absolute inset-0 bg-gradient-to-br from-primary/5 via-transparent to-sky-100/50" />
        <div className="absolute top-0 right-0 w-96 h-96 bg-primary/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2" />
        <div className="absolute bottom-0 left-0 w-96 h-96 bg-sky-200/30 rounded-full blur-3xl translate-y-1/2 -translate-x-1/2" />

        <div className="container mx-auto px-4 py-8 max-w-6xl relative">
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5 }}
          >
            <Button asChild variant="ghost" className="mb-8 hover:bg-white/50">
              <Link href="/">
                <ArrowLeft className="mr-2 h-4 w-4" />
                Voltar ao Sistema
              </Link>
            </Button>
          </motion.div>

          <div className="text-center py-12 md:py-20">
            <motion.div
              initial={{ opacity: 0, y: -20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6 }}
            >
              <Badge variant="secondary" className="mb-6 px-4 py-2 text-sm">
                <Sparkles className="h-4 w-4 mr-2" />
                Solucao Inovadora - LC 182/2021
              </Badge>
            </motion.div>

            <motion.div
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.6, delay: 0.2 }}
              className="flex items-center justify-center gap-3 mb-6"
            >
              <motion.div
                animate={{ scale: [1, 1.05, 1] }}
                transition={{ duration: 2, repeat: Infinity }}
                className="flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-primary to-sky-600 shadow-xl shadow-primary/30"
              >
                <Heart className="h-8 w-8 text-white" />
              </motion.div>
              <h1 className="text-4xl md:text-6xl font-bold bg-gradient-to-r from-primary via-primary to-sky-600 bg-clip-text text-transparent">
                VitalConnect
              </h1>
            </motion.div>

            <motion.p
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: 0.4 }}
              className="text-xl md:text-2xl text-muted-foreground max-w-3xl mx-auto mb-4"
            >
              Sistema Inteligente de Captacao de Corneas para
              <span className="text-primary font-semibold"> Centrais de Transplantes</span> e
              <span className="text-primary font-semibold"> Bancos de Olhos</span>
            </motion.p>

            <motion.p
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: 0.5 }}
              className="text-lg text-muted-foreground max-w-2xl mx-auto mb-8"
            >
              Transformando o processo de notificacao e captacao de corneas atraves de
              inteligencia artificial e automacao em tempo real
            </motion.p>

            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: 0.6 }}
              className="flex flex-wrap justify-center gap-4"
            >
              {[
                { icon: Building2, label: 'SES-GO' },
                { icon: Activity, label: 'Central de Transplantes' },
                { icon: Eye, label: 'Banco de Olhos' },
              ].map((item, i) => (
                <motion.div key={i} whileHover={{ scale: 1.05 }}>
                  <Badge variant="outline" className="px-4 py-2 text-sm border-primary/30">
                    <item.icon className="h-4 w-4 mr-2" />
                    {item.label}
                  </Badge>
                </motion.div>
              ))}
            </motion.div>

            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 1 }}
              className="mt-12"
            >
              <motion.div
                animate={{ y: [0, 10, 0] }}
                transition={{ duration: 2, repeat: Infinity }}
              >
                <ChevronDown className="h-8 w-8 mx-auto text-primary/50" />
              </motion.div>
            </motion.div>
          </div>
        </div>
      </motion.section>

      {/* Stats Section */}
      <section className="py-12 bg-gradient-to-r from-primary/5 via-sky-50 to-primary/5">
        <div className="container mx-auto px-4 max-w-6xl">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 md:gap-6">
            <StatCard value={6} suffix="h" label="Janela Critica" icon={Timer} delay={0} />
            <StatCard value={70} suffix="%" label="Corneas Perdidas" icon={AlertTriangle} delay={0.1} />
            <StatCard value={3} suffix="x" label="Aumento Captacao" icon={TrendingUp} delay={0.2} />
            <StatCard value={24} suffix="/7" label="Monitoramento" icon={Activity} delay={0.3} />
          </div>
        </div>
      </section>

      {/* Main Content */}
      <section className="py-16">
        <div className="container mx-auto px-4 max-w-6xl">

          {/* Problem Section - Redesigned */}
          <motion.div
            initial={{ opacity: 0, y: 40 }}
            whileInView={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            viewport={{ once: true }}
            className="mb-16"
          >
            {/* Section Header */}
            <div className="text-center mb-10">
              <motion.div
                initial={{ scale: 0 }}
                whileInView={{ scale: 1 }}
                transition={{ duration: 0.5, type: 'spring' }}
                viewport={{ once: true }}
                className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-red-500 to-orange-500 shadow-lg shadow-red-500/30 mb-4"
              >
                <AlertTriangle className="h-8 w-8 text-white" />
              </motion.div>
              <motion.h2
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.1 }}
                viewport={{ once: true }}
                className="text-3xl md:text-4xl font-bold text-gray-800 mb-3"
              >
                O Problema: Corneas que Salvam Vidas
                <span className="text-red-500"> Estao Sendo Perdidas</span>
              </motion.h2>
              <motion.p
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.2 }}
                viewport={{ once: true }}
                className="text-lg text-muted-foreground max-w-2xl mx-auto"
              >
                Um desafio critico de saude publica que afeta milhares de pacientes em todo o Brasil
              </motion.p>
            </div>

            {/* Critical Window Visual */}
            <motion.div
              initial={{ opacity: 0, scale: 0.95 }}
              whileInView={{ opacity: 1, scale: 1 }}
              transition={{ delay: 0.2 }}
              viewport={{ once: true }}
              className="bg-gradient-to-r from-red-50 via-orange-50 to-red-50 rounded-3xl p-8 mb-8 border border-red-100 relative overflow-hidden"
            >
              <div className="absolute top-0 right-0 w-64 h-64 bg-red-500/5 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2" />

              <div className="relative z-10">
                <div className="flex items-center justify-center gap-2 mb-6">
                  <Timer className="h-6 w-6 text-red-500" />
                  <h3 className="text-xl font-bold text-red-800">A Janela Critica de 6 Horas</h3>
                </div>

                {/* Timeline Visual */}
                <div className="relative max-w-3xl mx-auto">
                  <div className="h-4 bg-gray-200 rounded-full overflow-hidden">
                    <motion.div
                      initial={{ width: 0 }}
                      whileInView={{ width: '100%' }}
                      transition={{ duration: 2, ease: 'linear' }}
                      viewport={{ once: true }}
                      className="h-full bg-gradient-to-r from-green-500 via-yellow-500 to-red-500 rounded-full"
                    />
                  </div>

                  {/* Time Markers */}
                  <div className="flex justify-between mt-3 text-sm">
                    <div className="text-center">
                      <div className="font-bold text-green-600">0h</div>
                      <div className="text-xs text-muted-foreground">Obito</div>
                    </div>
                    <div className="text-center">
                      <div className="font-bold text-yellow-600">3h</div>
                      <div className="text-xs text-muted-foreground">Urgente</div>
                    </div>
                    <div className="text-center">
                      <div className="font-bold text-red-600">6h</div>
                      <div className="text-xs text-muted-foreground">Limite</div>
                    </div>
                  </div>

                  <p className="text-center mt-4 text-muted-foreground">
                    Apos <strong className="text-red-600">6 horas</strong>, a cornea se torna inviavel para transplante
                  </p>
                </div>
              </div>
            </motion.div>

            {/* Stats Grid */}
            <div className="grid md:grid-cols-3 gap-6 mb-8">
              {[
                { value: '70%', label: 'Corneas Perdidas', description: 'Por falhas na notificacao', color: 'red', icon: AlertTriangle },
                { value: '20.000+', label: 'Na Fila de Espera', description: 'Aguardando transplante', color: 'orange', icon: Heart },
                { value: '6h', label: 'Janela Critica', description: 'Tempo maximo viavel', color: 'yellow', icon: Clock },
              ].map((stat, i) => (
                <motion.div
                  key={i}
                  initial={{ opacity: 0, y: 30 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  transition={{ delay: i * 0.1 }}
                  viewport={{ once: true }}
                  whileHover={{ y: -5, scale: 1.02 }}
                  className={cn(
                    "relative p-6 rounded-2xl border-2 cursor-pointer transition-shadow hover:shadow-xl overflow-hidden",
                    stat.color === 'red' && "bg-red-50 border-red-200",
                    stat.color === 'orange' && "bg-orange-50 border-orange-200",
                    stat.color === 'yellow' && "bg-yellow-50 border-yellow-200"
                  )}
                >
                  <div className="absolute -right-4 -bottom-4 opacity-10">
                    <stat.icon className="h-24 w-24" />
                  </div>
                  <div className="relative z-10">
                    <div className={cn(
                      "text-4xl font-bold mb-1",
                      stat.color === 'red' && "text-red-600",
                      stat.color === 'orange' && "text-orange-600",
                      stat.color === 'yellow' && "text-yellow-600"
                    )}>
                      {stat.value}
                    </div>
                    <div className="font-semibold text-gray-800">{stat.label}</div>
                    <div className="text-sm text-muted-foreground">{stat.description}</div>
                  </div>
                </motion.div>
              ))}
            </div>

            {/* Causes Grid */}
            <div className="grid md:grid-cols-2 gap-6">
              <motion.div
                initial={{ opacity: 0, x: -30 }}
                whileInView={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.2 }}
                viewport={{ once: true }}
                className="bg-white rounded-2xl p-6 shadow-lg border border-gray-100"
              >
                <h4 className="font-bold text-lg mb-4 flex items-center gap-2 text-gray-800">
                  <Target className="h-5 w-5 text-red-500" />
                  Causas Identificadas
                </h4>
                <ul className="space-y-3">
                  {[
                    { text: 'Comunicacao manual e lenta entre hospitais e centrais', icon: Clock },
                    { text: 'Falta de integracao entre sistemas hospitalares', icon: Database },
                    { text: 'Ausencia de alertas automaticos para equipes', icon: Bell },
                    { text: 'Dificuldade em rastrear janela critica', icon: Timer },
                    { text: 'Processos burocraticos que consomem tempo', icon: FileCheck },
                  ].map((item, i) => (
                    <motion.li
                      key={i}
                      initial={{ opacity: 0, x: -20 }}
                      whileInView={{ opacity: 1, x: 0 }}
                      transition={{ delay: 0.3 + i * 0.1 }}
                      viewport={{ once: true }}
                      className="flex items-start gap-3 text-muted-foreground group"
                    >
                      <div className="p-1.5 rounded-lg bg-red-100 group-hover:bg-red-200 transition-colors shrink-0">
                        <item.icon className="h-4 w-4 text-red-500" />
                      </div>
                      <span>{item.text}</span>
                    </motion.li>
                  ))}
                </ul>
              </motion.div>

              <motion.div
                initial={{ opacity: 0, x: 30 }}
                whileInView={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.3 }}
                viewport={{ once: true }}
                className="bg-gradient-to-br from-red-500 to-orange-500 rounded-2xl p-6 text-white relative overflow-hidden"
              >
                <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjxnIGZpbGw9IiNmZmZmZmYiIGZpbGwtb3BhY2l0eT0iMC4xIj48cGF0aCBkPSJNMzYgMzRjMC0yLjIxLTEuNzktNC00LTRzLTQgMS43OS00IDQgMS43OSA0IDQgNCA0LTEuNzkgNC00eiIvPjwvZz48L2c+PC9zdmc+')] opacity-30" />
                <div className="relative z-10">
                  <div className="flex items-center gap-3 mb-4">
                    <motion.div
                      animate={{ scale: [1, 1.1, 1] }}
                      transition={{ duration: 2, repeat: Infinity }}
                    >
                      <Heart className="h-10 w-10" />
                    </motion.div>
                    <h4 className="font-bold text-xl">Impacto Humano</h4>
                  </div>
                  <p className="text-white/90 text-lg leading-relaxed">
                    Cada cornea perdida representa uma pessoa que permanece na fila de transplante,
                    aguardando por uma <strong>segunda chance de enxergar</strong>.
                  </p>
                  <div className="mt-6 p-4 bg-white/20 rounded-xl backdrop-blur-sm">
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="text-3xl font-bold">20.000+</div>
                        <div className="text-white/80 text-sm">pessoas na fila</div>
                      </div>
                      <Eye className="h-12 w-12 text-white/50" />
                    </div>
                  </div>
                </div>
              </motion.div>
            </div>
          </motion.div>

          {/* Solution Section - Redesigned */}
          <motion.div
            initial={{ opacity: 0, y: 40 }}
            whileInView={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            viewport={{ once: true }}
            className="mb-16"
          >
            {/* Section Header */}
            <div className="text-center mb-10">
              <motion.div
                initial={{ scale: 0 }}
                whileInView={{ scale: 1 }}
                transition={{ duration: 0.5, type: 'spring' }}
                viewport={{ once: true }}
                className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-primary to-sky-500 shadow-lg shadow-primary/30 mb-4"
              >
                <Lightbulb className="h-8 w-8 text-white" />
              </motion.div>
              <motion.h2
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.1 }}
                viewport={{ once: true }}
                className="text-3xl md:text-4xl font-bold text-gray-800 mb-3"
              >
                A Solucao: <span className="text-primary">VitalConnect</span>
              </motion.h2>
              <motion.p
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.2 }}
                viewport={{ once: true }}
                className="text-lg text-muted-foreground max-w-2xl mx-auto"
              >
                Inovacao tecnologica a servico da vida - transformando o processo de captacao de corneas
              </motion.p>
            </div>

            {/* Before/After Comparison */}
            <motion.div
              initial={{ opacity: 0, scale: 0.95 }}
              whileInView={{ opacity: 1, scale: 1 }}
              transition={{ delay: 0.2 }}
              viewport={{ once: true }}
              className="grid md:grid-cols-2 gap-6 mb-10"
            >
              {/* Before */}
              <div className="bg-gray-100 rounded-2xl p-6 border-2 border-gray-200 relative">
                <div className="absolute -top-3 left-6">
                  <Badge className="bg-gray-500 text-white font-bold px-4 py-1">ANTES</Badge>
                </div>
                <div className="pt-4">
                  <div className="flex items-center gap-3 mb-4 text-gray-600">
                    <AlertTriangle className="h-6 w-6" />
                    <span className="font-semibold">Processo Manual</span>
                  </div>
                  <ul className="space-y-3">
                    {[
                      'Notificacao por telefone/fax',
                      'Atrasos de horas na comunicacao',
                      'Sem rastreamento do tempo',
                      'Dados fragmentados',
                      'Alta taxa de perda',
                    ].map((item, i) => (
                      <li key={i} className="flex items-center gap-2 text-gray-500">
                        <div className="h-2 w-2 rounded-full bg-gray-400" />
                        {item}
                      </li>
                    ))}
                  </ul>
                  <div className="mt-6 p-4 bg-red-100 rounded-xl">
                    <div className="text-2xl font-bold text-red-600">70%</div>
                    <div className="text-sm text-red-700">Corneas perdidas</div>
                  </div>
                </div>
              </div>

              {/* After */}
              <div className="bg-gradient-to-br from-primary/10 to-sky-100 rounded-2xl p-6 border-2 border-primary/30 relative">
                <div className="absolute -top-3 left-6">
                  <Badge className="bg-primary text-white font-bold px-4 py-1">COM VITALCONNECT</Badge>
                </div>
                <div className="pt-4">
                  <div className="flex items-center gap-3 mb-4 text-primary">
                    <Zap className="h-6 w-6" />
                    <span className="font-semibold">Processo Automatizado</span>
                  </div>
                  <ul className="space-y-3">
                    {[
                      'Deteccao automatica de obitos',
                      'Notificacao em segundos',
                      'Countdown visual da janela',
                      'Dados centralizados',
                      'Aumento de captacao',
                    ].map((item, i) => (
                      <motion.li
                        key={i}
                        initial={{ opacity: 0, x: 20 }}
                        whileInView={{ opacity: 1, x: 0 }}
                        transition={{ delay: 0.3 + i * 0.1 }}
                        viewport={{ once: true }}
                        className="flex items-center gap-2 text-gray-700"
                      >
                        <CheckCircle2 className="h-5 w-5 text-primary shrink-0" />
                        {item}
                      </motion.li>
                    ))}
                  </ul>
                  <div className="mt-6 p-4 bg-green-100 rounded-xl">
                    <div className="text-2xl font-bold text-green-600">3x</div>
                    <div className="text-sm text-green-700">Aumento na captacao</div>
                  </div>
                </div>
              </div>
            </motion.div>

            {/* Features Grid */}
            <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-5 mb-10">
              {[
                { icon: Eye, title: 'Deteccao Automatica', description: 'Monitoramento continuo 24/7 dos sistemas hospitalares via integracao HL7 FHIR', color: 'blue' },
                { icon: Brain, title: 'IA para Triagem', description: 'Algoritmos inteligentes aplicam criterios de elegibilidade automaticamente', color: 'purple' },
                { icon: Bell, title: 'Alertas em Tempo Real', description: 'Notificacoes instantaneas via dashboard, push, email e WhatsApp', color: 'orange' },
                { icon: Clock, title: 'Gestao de Tempo', description: 'Contagem regressiva visual da janela critica de 6 horas', color: 'red' },
              ].map((feature, i) => (
                <motion.div
                  key={i}
                  initial={{ opacity: 0, y: 30 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  transition={{ delay: i * 0.1 }}
                  viewport={{ once: true }}
                  whileHover={{ y: -8, scale: 1.02 }}
                  className="group p-6 bg-white rounded-2xl border border-gray-200 shadow-sm hover:shadow-xl transition-all duration-300 cursor-pointer"
                >
                  <motion.div
                    whileHover={{ rotate: 10, scale: 1.1 }}
                    className={cn(
                      "h-14 w-14 rounded-2xl flex items-center justify-center mb-4 transition-colors duration-300",
                      feature.color === 'blue' && "bg-blue-100 group-hover:bg-blue-500",
                      feature.color === 'purple' && "bg-purple-100 group-hover:bg-purple-500",
                      feature.color === 'orange' && "bg-orange-100 group-hover:bg-orange-500",
                      feature.color === 'red' && "bg-red-100 group-hover:bg-red-500"
                    )}
                  >
                    <feature.icon className={cn(
                      "h-7 w-7 transition-colors duration-300",
                      feature.color === 'blue' && "text-blue-600 group-hover:text-white",
                      feature.color === 'purple' && "text-purple-600 group-hover:text-white",
                      feature.color === 'orange' && "text-orange-600 group-hover:text-white",
                      feature.color === 'red' && "text-red-600 group-hover:text-white"
                    )} />
                  </motion.div>
                  <h3 className="font-bold text-lg mb-2 text-gray-800">{feature.title}</h3>
                  <p className="text-sm text-muted-foreground">{feature.description}</p>
                </motion.div>
              ))}
            </div>

            {/* Pipeline Visual - Enhanced */}
            <motion.div
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              className="bg-gradient-to-br from-sky-50 via-white to-primary/5 rounded-3xl p-8 border border-sky-100 shadow-lg"
            >
              <div className="flex items-center justify-center gap-3 mb-8">
                <motion.div
                  whileHover={{ rotate: 360 }}
                  transition={{ duration: 0.5 }}
                  className="p-3 bg-primary/10 rounded-xl"
                >
                  <Workflow className="h-6 w-6 text-primary" />
                </motion.div>
                <h3 className="text-2xl font-bold text-gray-800">Fluxo de Funcionamento</h3>
              </div>

              {/* Horizontal Pipeline for Desktop */}
              <div className="hidden lg:block">
                <div className="relative flex items-start justify-between">
                  {/* Connection Line */}
                  <div className="absolute top-8 left-12 right-12 h-1 bg-gray-200 rounded-full">
                    <motion.div
                      initial={{ width: 0 }}
                      whileInView={{ width: '100%' }}
                      transition={{ duration: 2, ease: [0.25, 0.46, 0.45, 0.94] }}
                      viewport={{ once: true }}
                      className="h-full bg-gradient-to-r from-primary via-sky-500 to-primary rounded-full"
                    />
                  </div>

                  {[
                    { step: 1, title: 'Monitoramento', desc: 'Monitora registros de obito 24/7', icon: Activity },
                    { step: 2, title: 'Deteccao', desc: 'Coleta dados do paciente', icon: Database },
                    { step: 3, title: 'Triagem', desc: 'Aplica criterios de elegibilidade', icon: Shield },
                    { step: 4, title: 'Notificacao', desc: 'Alerta equipe instantaneamente', icon: Bell },
                    { step: 5, title: 'Gestao', desc: 'Dashboard em tempo real', icon: LineChart },
                  ].map((item, i) => (
                    <motion.div
                      key={i}
                      initial={{ opacity: 0, y: 20 }}
                      whileInView={{ opacity: 1, y: 0 }}
                      transition={{ delay: i * 0.15 }}
                      viewport={{ once: true }}
                      className="relative z-10 flex flex-col items-center w-1/5"
                    >
                      <motion.div
                        whileHover={{ scale: 1.1, y: -5 }}
                        className="flex items-center justify-center w-16 h-16 rounded-2xl bg-white border-2 border-primary shadow-lg shadow-primary/20 mb-4"
                      >
                        <item.icon className="h-8 w-8 text-primary" />
                      </motion.div>
                      <div className="text-center">
                        <div className="text-xs font-bold text-primary mb-1">PASSO {item.step}</div>
                        <div className="font-semibold text-gray-800 text-sm">{item.title}</div>
                        <div className="text-xs text-muted-foreground mt-1">{item.desc}</div>
                      </div>
                    </motion.div>
                  ))}
                </div>
              </div>

              {/* Vertical Pipeline for Mobile */}
              <div className="lg:hidden space-y-2">
                {[
                  { step: 1, title: 'Monitoramento Continuo', description: 'O sistema monitora em tempo real os registros de obitos nos hospitais integrados', icon: Activity },
                  { step: 2, title: 'Deteccao Automatica', description: 'Ao detectar um obito, coleta automaticamente dados demograficos e clinicos', icon: Database },
                  { step: 3, title: 'Triagem Inteligente', description: 'Criterios de elegibilidade sao aplicados automaticamente', icon: Shield },
                  { step: 4, title: 'Notificacao Instantanea', description: 'A equipe recebe alerta imediato com todos os dados necessarios', icon: Bell },
                  { step: 5, title: 'Gestao e Acompanhamento', description: 'Dashboard permite gerenciar casos e gerar relatorios', icon: LineChart },
                ].map((item, i) => (
                  <PipelineStep key={i} step={item.step} title={item.title} description={item.description} icon={item.icon} isLast={i === 4} delay={i * 0.1} />
                ))}
              </div>
            </motion.div>
          </motion.div>

          {/* Innovation Section */}
          <AnimatedCard className="mb-12 overflow-hidden border-none shadow-xl bg-gradient-to-br from-purple-50 to-white" delay={0.1}>
            <CardHeader>
              <div className="flex items-center gap-3">
                <motion.div whileHover={{ rotate: 10 }} className="p-3 bg-purple-100 rounded-xl">
                  <Sparkles className="h-6 w-6 text-purple-600" />
                </motion.div>
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
                      <motion.li
                        key={i}
                        initial={{ opacity: 0, x: -20 }}
                        whileInView={{ opacity: 1, x: 0 }}
                        transition={{ delay: i * 0.1 }}
                        viewport={{ once: true }}
                        className="flex items-start gap-2"
                      >
                        <CheckCircle2 className="h-5 w-5 text-purple-500 shrink-0 mt-0.5" />
                        <span className="text-muted-foreground">{item}</span>
                      </motion.li>
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
                      <motion.li
                        key={i}
                        initial={{ opacity: 0, x: 20 }}
                        whileInView={{ opacity: 1, x: 0 }}
                        transition={{ delay: i * 0.1 }}
                        viewport={{ once: true }}
                        className="flex items-start gap-2"
                      >
                        <CheckCircle2 className="h-5 w-5 text-purple-500 shrink-0 mt-0.5" />
                        <span className="text-muted-foreground">{item}</span>
                      </motion.li>
                    ))}
                  </ul>
                </div>
              </div>
            </CardContent>
          </AnimatedCard>

          {/* TRL Section */}
          <AnimatedCard className="mb-12 overflow-hidden border-none shadow-xl" delay={0.1}>
            <CardHeader>
              <div className="flex items-center gap-3">
                <motion.div whileHover={{ rotate: 10 }} className="p-3 bg-green-100 rounded-xl">
                  <Layers className="h-6 w-6 text-green-600" />
                </motion.div>
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
                  { title: 'Desenvolvido', items: ['Backend API completo em Go', 'Frontend responsivo em React/Next.js', 'Sistema de autenticacao JWT', 'Modulo de notificacoes'], color: 'green' },
                  { title: 'Em Validacao', items: ['Integracao HL7 FHIR', 'Assistente IA', 'Relatorios avancados', 'App mobile PWA'], color: 'yellow' },
                  { title: 'Roadmap', items: ['Machine Learning preditivo', 'Integracao SNT', 'Blockchain para rastreabilidade', 'Analytics avancado'], color: 'blue' },
                ].map((col, i) => (
                  <motion.div
                    key={i}
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    transition={{ delay: i * 0.1 }}
                    viewport={{ once: true }}
                    whileHover={{ scale: 1.02 }}
                    className={cn(
                      "p-5 rounded-xl border cursor-pointer transition-shadow hover:shadow-lg",
                      col.color === 'green' && "bg-green-50 border-green-200",
                      col.color === 'yellow' && "bg-yellow-50 border-yellow-200",
                      col.color === 'blue' && "bg-blue-50 border-blue-200",
                    )}
                  >
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
                  </motion.div>
                ))}
              </div>
            </CardContent>
          </AnimatedCard>

          {/* Integration Section */}
          <AnimatedCard className="mb-12 overflow-hidden border-none shadow-xl" delay={0.1}>
            <CardHeader>
              <div className="flex items-center gap-3">
                <motion.div whileHover={{ rotate: 10 }} className="p-3 bg-sky-100 rounded-xl">
                  <Globe className="h-6 w-6 text-sky-600" />
                </motion.div>
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
                      <motion.div
                        key={i}
                        initial={{ opacity: 0, x: -20 }}
                        whileInView={{ opacity: 1, x: 0 }}
                        transition={{ delay: i * 0.1 }}
                        viewport={{ once: true }}
                        whileHover={{ x: 5 }}
                        className="flex items-center justify-between p-3 bg-sky-50 rounded-lg cursor-pointer hover:bg-sky-100 transition-colors"
                      >
                        <span className="font-medium">{sys.name}</span>
                        <Badge variant={sys.status === 'Nativo' ? 'default' : sys.status === 'Compativel' ? 'secondary' : 'outline'}>
                          {sys.status}
                        </Badge>
                      </motion.div>
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
                      <motion.li
                        key={i}
                        initial={{ opacity: 0, x: 20 }}
                        whileInView={{ opacity: 1, x: 0 }}
                        transition={{ delay: i * 0.1 }}
                        viewport={{ once: true }}
                        className="flex items-start gap-2 text-muted-foreground"
                      >
                        <CheckCircle2 className="h-4 w-4 text-sky-500 shrink-0 mt-1" />
                        {cap}
                      </motion.li>
                    ))}
                  </ul>
                </div>
              </div>
            </CardContent>
          </AnimatedCard>

          {/* Economic Viability Section */}
          <AnimatedCard className="mb-12 overflow-hidden border-none shadow-xl bg-gradient-to-br from-emerald-50 to-white" delay={0.1}>
            <CardHeader>
              <div className="flex items-center gap-3">
                <motion.div whileHover={{ rotate: 10 }} className="p-3 bg-emerald-100 rounded-xl">
                  <TrendingUp className="h-6 w-6 text-emerald-600" />
                </motion.div>
                <div>
                  <CardTitle className="text-2xl">Viabilidade Economica e Custo-Beneficio</CardTitle>
                  <CardDescription>Analise de retorno sobre investimento e sustentabilidade</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid md:grid-cols-3 gap-6">
                {[
                  { title: 'Reducao de Custos', value: 'R$ 15.000', description: 'Economia por transplante viabilizado vs. tratamento continuo', icon: LineChart },
                  { title: 'ROI Estimado', value: '300%', description: 'Retorno sobre investimento no primeiro ano', icon: TrendingUp },
                  { title: 'Payback', value: '4 meses', description: 'Tempo estimado para retorno do investimento', icon: Timer },
                ].map((metric, i) => (
                  <motion.div
                    key={i}
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    transition={{ delay: i * 0.1 }}
                    viewport={{ once: true }}
                    whileHover={{ y: -5, scale: 1.02 }}
                    className="p-6 bg-white rounded-xl border border-emerald-100 text-center cursor-pointer hover:shadow-lg transition-shadow"
                  >
                    <metric.icon className="h-8 w-8 mx-auto mb-3 text-emerald-500" />
                    <div className="text-3xl font-bold text-emerald-600 mb-1">{metric.value}</div>
                    <div className="font-medium mb-2">{metric.title}</div>
                    <p className="text-sm text-muted-foreground">{metric.description}</p>
                  </motion.div>
                ))}
              </div>

              <motion.div
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                className="bg-emerald-50 border border-emerald-100 rounded-xl p-6"
              >
                <h4 className="font-semibold text-emerald-800 mb-4">Modelo de Negocio Sustentavel</h4>
                <div className="grid md:grid-cols-2 gap-6">
                  <div>
                    <h5 className="font-medium mb-2">Custos Operacionais Baixos:</h5>
                    <ul className="space-y-2 text-sm text-muted-foreground">
                      {['Infraestrutura em nuvem escalavel (pay-as-you-go)', 'Sem necessidade de hardware dedicado', 'Manutencao automatizada e atualizacoes continuas'].map((item, i) => (
                        <li key={i} className="flex items-center gap-2">
                          <CheckCircle2 className="h-4 w-4 text-emerald-500" />
                          {item}
                        </li>
                      ))}
                    </ul>
                  </div>
                  <div>
                    <h5 className="font-medium mb-2">Escalabilidade:</h5>
                    <ul className="space-y-2 text-sm text-muted-foreground">
                      {['Arquitetura multi-tenant para multiplos hospitais', 'Expansao sem custos proporcionais', 'Replicavel para outros estados'].map((item, i) => (
                        <li key={i} className="flex items-center gap-2">
                          <CheckCircle2 className="h-4 w-4 text-emerald-500" />
                          {item}
                        </li>
                      ))}
                    </ul>
                  </div>
                </div>
              </motion.div>
            </CardContent>
          </AnimatedCard>

          {/* Security Section */}
          <AnimatedCard className="mb-12 overflow-hidden border-none shadow-xl" delay={0.1}>
            <CardHeader>
              <div className="flex items-center gap-3">
                <motion.div whileHover={{ rotate: 10 }} className="p-3 bg-slate-100 rounded-xl">
                  <Lock className="h-6 w-6 text-slate-600" />
                </motion.div>
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
                    {['Criptografia AES-256 para dados em repouso', 'TLS 1.3 para dados em transito', 'Autenticacao JWT com refresh tokens', 'Controle de acesso baseado em papeis (RBAC)', 'Logs de auditoria completos'].map((item, i) => (
                      <motion.li
                        key={i}
                        initial={{ opacity: 0, x: -20 }}
                        whileInView={{ opacity: 1, x: 0 }}
                        transition={{ delay: i * 0.1 }}
                        viewport={{ once: true }}
                        className="flex items-center gap-2 text-muted-foreground"
                      >
                        <Shield className="h-4 w-4 text-slate-500" />
                        {item}
                      </motion.li>
                    ))}
                  </ul>
                </div>
                <div className="space-y-4">
                  <h4 className="font-semibold">Conformidade LGPD:</h4>
                  <ul className="space-y-3">
                    {['Minimizacao de coleta de dados pessoais', 'Anonimizacao de dados para relatorios', 'Retencao de dados configuravel', 'Direito ao esquecimento implementado', 'Consentimento explicito para processamento'].map((item, i) => (
                      <motion.li
                        key={i}
                        initial={{ opacity: 0, x: 20 }}
                        whileInView={{ opacity: 1, x: 0 }}
                        transition={{ delay: i * 0.1 }}
                        viewport={{ once: true }}
                        className="flex items-center gap-2 text-muted-foreground"
                      >
                        <FileCheck className="h-4 w-4 text-slate-500" />
                        {item}
                      </motion.li>
                    ))}
                  </ul>
                </div>
              </div>
            </CardContent>
          </AnimatedCard>

          {/* Technical Stack */}
          <AnimatedCard className="mb-12 overflow-hidden border-none shadow-xl" delay={0.1}>
            <CardHeader>
              <div className="flex items-center gap-3">
                <motion.div whileHover={{ rotate: 10 }} className="p-3 bg-indigo-100 rounded-xl">
                  <Server className="h-6 w-6 text-indigo-600" />
                </motion.div>
                <div>
                  <CardTitle className="text-2xl">Arquitetura Tecnologica</CardTitle>
                  <CardDescription>Stack moderno, escalavel e de alta performance</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid md:grid-cols-3 gap-6">
                {[
                  { title: 'Backend', color: 'indigo', techs: [{ name: 'Go (Golang)', desc: 'Alta performance' }, { name: 'Gin Framework', desc: 'API RESTful' }, { name: 'PostgreSQL 15+', desc: 'Banco relacional' }, { name: 'Redis 7+', desc: 'Cache e Pub/Sub' }, { name: 'JWT', desc: 'Autenticacao segura' }] },
                  { title: 'Frontend', color: 'sky', techs: [{ name: 'Next.js 14+', desc: 'App Router' }, { name: 'React 18+', desc: 'UI reativa' }, { name: 'TypeScript', desc: 'Type safety' }, { name: 'Tailwind CSS', desc: 'Estilizacao' }, { name: 'TanStack Query', desc: 'Data fetching' }] },
                  { title: 'IA & Servicos', color: 'purple', techs: [{ name: 'Python/FastAPI', desc: 'Servico de IA' }, { name: 'OpenAI GPT-4', desc: 'LLM' }, { name: 'Docker', desc: 'Containerizacao' }, { name: 'GitHub Actions', desc: 'CI/CD' }, { name: 'Render/AWS', desc: 'Cloud hosting' }] },
                ].map((stack, i) => (
                  <motion.div
                    key={i}
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    transition={{ delay: i * 0.1 }}
                    viewport={{ once: true }}
                    whileHover={{ scale: 1.02 }}
                    className={cn(
                      "p-5 rounded-xl cursor-pointer transition-shadow hover:shadow-lg",
                      stack.color === 'indigo' && "bg-indigo-50",
                      stack.color === 'sky' && "bg-sky-50",
                      stack.color === 'purple' && "bg-purple-50",
                    )}
                  >
                    <h4 className={cn(
                      "font-semibold mb-4",
                      stack.color === 'indigo' && "text-indigo-800",
                      stack.color === 'sky' && "text-sky-800",
                      stack.color === 'purple' && "text-purple-800",
                    )}>{stack.title}</h4>
                    <ul className="space-y-2 text-sm">
                      {stack.techs.map((tech, j) => (
                        <li key={j} className="flex justify-between">
                          <span className="font-medium">{tech.name}</span>
                          <span className="text-muted-foreground">{tech.desc}</span>
                        </li>
                      ))}
                    </ul>
                  </motion.div>
                ))}
              </div>
            </CardContent>
          </AnimatedCard>

          {/* Timeline Section */}
          <AnimatedCard className="mb-12 overflow-hidden border-none shadow-xl" delay={0.1}>
            <CardHeader>
              <div className="flex items-center gap-3">
                <motion.div whileHover={{ rotate: 10 }} className="p-3 bg-blue-100 rounded-xl">
                  <Clock className="h-6 w-6 text-blue-600" />
                </motion.div>
                <div>
                  <CardTitle className="text-2xl">Timeline do Projeto</CardTitle>
                  <CardDescription>Acompanhe as fases de desenvolvimento e implementacao</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <ProjectTimeline />
            </CardContent>
          </AnimatedCard>

          {/* Testimonials */}
          <AnimatedCard className="mb-12 overflow-hidden border-none shadow-xl bg-gradient-to-br from-sky-50 to-white" delay={0.1}>
            <CardHeader>
              <div className="flex items-center gap-3">
                <motion.div whileHover={{ rotate: 10 }} className="p-3 bg-sky-100 rounded-xl">
                  <Quote className="h-6 w-6 text-sky-600" />
                </motion.div>
                <div>
                  <CardTitle className="text-2xl">O que Esperam os Especialistas</CardTitle>
                  <CardDescription>Perspectivas de profissionais da area de transplantes</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid md:grid-cols-2 gap-8">
                <TestimonialCard
                  quote="A automacao do processo de notificacao pode revolucionar a captacao de corneas. Perdemos muitas oportunidades por falhas na comunicacao que um sistema como este poderia resolver."
                  author="Dra. Maria Silva"
                  role="Coordenadora de Transplantes"
                  institution="Hospital de Referencia"
                  delay={0}
                />
                <TestimonialCard
                  quote="A janela de 6 horas e nosso maior desafio. Ter um sistema que monitora automaticamente e nos alerta em tempo real seria transformador para nosso trabalho."
                  author="Dr. Carlos Santos"
                  role="Oftalmologista"
                  institution="Banco de Olhos"
                  delay={0.1}
                />
              </div>
            </CardContent>
          </AnimatedCard>

          {/* FAQ Section */}
          <AnimatedCard className="mb-12 overflow-hidden border-none shadow-xl" delay={0.1}>
            <CardHeader>
              <div className="flex items-center gap-3">
                <motion.div whileHover={{ rotate: 10 }} className="p-3 bg-orange-100 rounded-xl">
                  <HelpCircle className="h-6 w-6 text-orange-600" />
                </motion.div>
                <div>
                  <CardTitle className="text-2xl">Perguntas Frequentes</CardTitle>
                  <CardDescription>Tire suas duvidas sobre o VitalConnect</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <FAQAccordion />
            </CardContent>
          </AnimatedCard>

          {/* CTA Section */}
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            className="bg-gradient-to-r from-primary via-sky-600 to-primary rounded-3xl p-8 md:p-12 text-center text-white relative overflow-hidden"
          >
            <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjxnIGZpbGw9IiNmZmZmZmYiIGZpbGwtb3BhY2l0eT0iMC4xIj48cGF0aCBkPSJNMzYgMzRjMC0yLjIxLTEuNzktNC00LTRzLTQgMS43OS00IDQgMS43OSA0IDQgNCA0LTEuNzkgNC00eiIvPjwvZz48L2c+PC9zdmc+')] opacity-30" />
            <div className="relative z-10">
              <motion.h2
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                className="text-3xl md:text-4xl font-bold mb-4"
              >
                Quer saber mais sobre o VitalConnect?
              </motion.h2>
              <motion.p
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.1 }}
                viewport={{ once: true }}
                className="text-xl text-white/80 mb-8 max-w-2xl mx-auto"
              >
                Entre em contato conosco para conhecer como o sistema pode transformar
                a captacao de corneas na sua instituicao
              </motion.p>
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.2 }}
                viewport={{ once: true }}
                whileHover={{ scale: 1.05 }}
              >
                <Button asChild size="lg" variant="secondary" className="text-lg px-8 py-6 shadow-xl">
                  <Link href="/contact">
                    <Mail className="mr-2 h-5 w-5" />
                    Entrar em Contato
                  </Link>
                </Button>
              </motion.div>
            </div>
          </motion.div>

        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gradient-to-r from-primary to-sky-700 text-white py-12">
        <div className="container mx-auto px-4 max-w-6xl">
          <div className="grid md:grid-cols-3 gap-8 mb-8">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
            >
              <div className="flex items-center gap-2 mb-4">
                <Heart className="h-6 w-6" />
                <span className="text-xl font-bold">VitalConnect</span>
              </div>
              <p className="text-white/80 text-sm">
                Transformando a captacao de corneas atraves da tecnologia e salvando vidas.
              </p>
            </motion.div>
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.1 }}
              viewport={{ once: true }}
            >
              <h4 className="font-semibold mb-4">Contato</h4>
              <p className="text-white/80 text-sm">Secretaria de Estado da Saude de Goias</p>
              <p className="text-white/80 text-sm">Central Estadual de Transplantes</p>
              <p className="text-white/80 text-sm">Banco de Olhos de Goias</p>
            </motion.div>
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
              viewport={{ once: true }}
            >
              <h4 className="font-semibold mb-4">Edital</h4>
              <p className="text-white/80 text-sm">CPSI - Contrato Publico de Solucao Inovadora</p>
              <p className="text-white/80 text-sm">Edital No 01/2025 - SES/GO</p>
              <p className="text-white/80 text-sm">Lei Complementar No 182/2021</p>
            </motion.div>
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

// HelpCircle icon
function HelpCircle(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <circle cx="12" cy="12" r="10" />
      <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
      <path d="M12 17h.01" />
    </svg>
  );
}
