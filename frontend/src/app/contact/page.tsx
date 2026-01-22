'use client';

import Link from 'next/link';
import { useState } from 'react';
import { motion } from 'framer-motion';
import {
  ArrowLeft,
  Mail,
  Phone,
  MapPin,
  Clock,
  Send,
  Building2,
  Heart,
  MessageSquare,
  Users,
  CheckCircle2,
  Loader2,
  AlertCircle,
  Globe,
  Linkedin,
  Github,
  Twitter,
  ExternalLink,
  ChevronRight,
  FileText,
  HelpCircle,
  Sparkles,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { cn } from '@/lib/utils';

// Form state type
interface FormData {
  name: string;
  email: string;
  phone: string;
  institution: string;
  subject: string;
  message: string;
}

interface FormErrors {
  name?: string;
  email?: string;
  message?: string;
}

// Contact Info Card Component
function ContactCard({
  icon: Icon,
  title,
  content,
  description,
  href,
  delay = 0
}: {
  icon: React.ElementType;
  title: string;
  content: string;
  description?: string;
  href?: string;
  delay?: number;
}) {
  const CardWrapper = href ? 'a' : 'div';

  return (
    <motion.div
      initial={{ opacity: 0, y: 30 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay }}
      viewport={{ once: true }}
      whileHover={{ y: -5, scale: 1.02 }}
    >
      <CardWrapper
        href={href}
        target={href?.startsWith('http') ? '_blank' : undefined}
        rel={href?.startsWith('http') ? 'noopener noreferrer' : undefined}
        className={cn(
          "block p-6 bg-white rounded-2xl border border-gray-100 shadow-sm hover:shadow-xl transition-all duration-300",
          href && "cursor-pointer"
        )}
      >
        <div className="flex items-start gap-4">
          <div className="p-3 bg-primary/10 rounded-xl shrink-0">
            <Icon className="h-6 w-6 text-primary" />
          </div>
          <div className="flex-1 min-w-0">
            <h3 className="font-semibold text-gray-800 mb-1">{title}</h3>
            <p className="text-primary font-medium break-all">{content}</p>
            {description && (
              <p className="text-sm text-muted-foreground mt-1">{description}</p>
            )}
          </div>
          {href && (
            <ExternalLink className="h-4 w-4 text-muted-foreground shrink-0" />
          )}
        </div>
      </CardWrapper>
    </motion.div>
  );
}

// Quick Link Card
function QuickLinkCard({
  icon: Icon,
  title,
  description,
  href,
  color = 'blue',
  delay = 0,
}: {
  icon: React.ElementType;
  title: string;
  description: string;
  href: string;
  color?: 'blue' | 'purple' | 'green' | 'orange';
  delay?: number;
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, delay }}
      viewport={{ once: true }}
    >
      <Link href={href}>
        <motion.div
          whileHover={{ y: -5, scale: 1.02 }}
          className={cn(
            "p-5 rounded-2xl border-2 transition-all duration-300 cursor-pointer group",
            color === 'blue' && "bg-blue-50 border-blue-200 hover:border-blue-400",
            color === 'purple' && "bg-purple-50 border-purple-200 hover:border-purple-400",
            color === 'green' && "bg-green-50 border-green-200 hover:border-green-400",
            color === 'orange' && "bg-orange-50 border-orange-200 hover:border-orange-400"
          )}
        >
          <div className="flex items-center gap-4">
            <div className={cn(
              "p-2.5 rounded-xl transition-colors",
              color === 'blue' && "bg-blue-100 group-hover:bg-blue-200",
              color === 'purple' && "bg-purple-100 group-hover:bg-purple-200",
              color === 'green' && "bg-green-100 group-hover:bg-green-200",
              color === 'orange' && "bg-orange-100 group-hover:bg-orange-200"
            )}>
              <Icon className={cn(
                "h-5 w-5",
                color === 'blue' && "text-blue-600",
                color === 'purple' && "text-purple-600",
                color === 'green' && "text-green-600",
                color === 'orange' && "text-orange-600"
              )} />
            </div>
            <div className="flex-1">
              <h4 className="font-semibold text-gray-800">{title}</h4>
              <p className="text-sm text-muted-foreground">{description}</p>
            </div>
            <ChevronRight className={cn(
              "h-5 w-5 transition-transform group-hover:translate-x-1",
              color === 'blue' && "text-blue-400",
              color === 'purple' && "text-purple-400",
              color === 'green' && "text-green-400",
              color === 'orange' && "text-orange-400"
            )} />
          </div>
        </motion.div>
      </Link>
    </motion.div>
  );
}

// Team Member Card
function TeamMemberCard({
  name,
  role,
  email,
  delay = 0,
}: {
  name: string;
  role: string;
  email: string;
  delay?: number;
}) {
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.9 }}
      whileInView={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.4, delay }}
      viewport={{ once: true }}
      whileHover={{ y: -3 }}
      className="p-4 bg-white rounded-xl border border-gray-100 shadow-sm hover:shadow-md transition-all"
    >
      <div className="flex items-center gap-3">
        <div className="w-12 h-12 rounded-full bg-gradient-to-br from-primary to-sky-500 flex items-center justify-center text-white font-bold text-lg">
          {name.charAt(0)}
        </div>
        <div className="flex-1 min-w-0">
          <h4 className="font-semibold text-gray-800 truncate">{name}</h4>
          <p className="text-sm text-muted-foreground truncate">{role}</p>
        </div>
      </div>
      <a
        href={`mailto:${email}`}
        className="mt-3 flex items-center gap-2 text-sm text-primary hover:underline"
      >
        <Mail className="h-3.5 w-3.5" />
        {email}
      </a>
    </motion.div>
  );
}

export default function ContactPage() {
  const [formData, setFormData] = useState<FormData>({
    name: '',
    email: '',
    phone: '',
    institution: '',
    subject: '',
    message: '',
  });
  const [errors, setErrors] = useState<FormErrors>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitStatus, setSubmitStatus] = useState<'idle' | 'success' | 'error'>('idle');

  const validateForm = (): boolean => {
    const newErrors: FormErrors = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Nome e obrigatorio';
    }

    if (!formData.email.trim()) {
      newErrors.email = 'Email e obrigatorio';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      newErrors.email = 'Email invalido';
    }

    if (!formData.message.trim()) {
      newErrors.message = 'Mensagem e obrigatoria';
    } else if (formData.message.length < 10) {
      newErrors.message = 'Mensagem deve ter pelo menos 10 caracteres';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) return;

    setIsSubmitting(true);
    setSubmitStatus('idle');

    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 2000));

      // Here you would send the form data to your backend
      console.log('Form submitted:', formData);

      setSubmitStatus('success');
      setFormData({
        name: '',
        email: '',
        phone: '',
        institution: '',
        subject: '',
        message: '',
      });
    } catch {
      setSubmitStatus('error');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));

    // Clear error when user starts typing
    if (errors[name as keyof FormErrors]) {
      setErrors(prev => ({ ...prev, [name]: undefined }));
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-sky-50 via-white to-sky-50">
      {/* Hero Section */}
      <section className="relative overflow-hidden">
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
              <Link href="/about">
                <ArrowLeft className="mr-2 h-4 w-4" />
                Voltar ao Sobre
              </Link>
            </Button>
          </motion.div>

          <div className="text-center py-12 md:py-16">
            <motion.div
              initial={{ opacity: 0, y: -20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6 }}
            >
              <Badge variant="secondary" className="mb-6 px-4 py-2 text-sm">
                <MessageSquare className="h-4 w-4 mr-2" />
                Fale Conosco
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
                <Mail className="h-8 w-8 text-white" />
              </motion.div>
              <h1 className="text-4xl md:text-5xl font-bold bg-gradient-to-r from-primary via-primary to-sky-600 bg-clip-text text-transparent">
                Entre em Contato
              </h1>
            </motion.div>

            <motion.p
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: 0.4 }}
              className="text-xl text-muted-foreground max-w-2xl mx-auto mb-4"
            >
              Estamos prontos para responder suas duvidas e apresentar como o
              <span className="text-primary font-semibold"> SIDOT</span> pode transformar
              a captacao de corneas na sua instituicao
            </motion.p>

            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: 0.5 }}
              className="flex flex-wrap justify-center gap-4 mt-8"
            >
              {[
                { icon: Clock, label: 'Resposta em 24h' },
                { icon: Users, label: 'Suporte Especializado' },
                { icon: Heart, label: 'Compromisso com a Vida' },
              ].map((item, i) => (
                <motion.div key={i} whileHover={{ scale: 1.05 }}>
                  <Badge variant="outline" className="px-4 py-2 text-sm border-primary/30">
                    <item.icon className="h-4 w-4 mr-2" />
                    {item.label}
                  </Badge>
                </motion.div>
              ))}
            </motion.div>
          </div>
        </div>
      </section>

      {/* Main Content */}
      <section className="py-12">
        <div className="container mx-auto px-4 max-w-6xl">
          <div className="grid lg:grid-cols-5 gap-8">

            {/* Contact Form - 3 columns */}
            <motion.div
              initial={{ opacity: 0, x: -30 }}
              whileInView={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.6 }}
              viewport={{ once: true }}
              className="lg:col-span-3"
            >
              <Card className="border-none shadow-xl overflow-hidden">
                <div className="bg-gradient-to-r from-primary to-sky-600 p-6 text-white">
                  <div className="flex items-center gap-3">
                    <Send className="h-6 w-6" />
                    <div>
                      <h2 className="text-xl font-bold">Envie sua Mensagem</h2>
                      <p className="text-white/80 text-sm">Preencha o formulario abaixo</p>
                    </div>
                  </div>
                </div>

                <CardContent className="p-6">
                  {submitStatus === 'success' ? (
                    <motion.div
                      initial={{ opacity: 0, scale: 0.9 }}
                      animate={{ opacity: 1, scale: 1 }}
                      className="py-12 text-center"
                    >
                      <motion.div
                        initial={{ scale: 0 }}
                        animate={{ scale: 1 }}
                        transition={{ type: 'spring', delay: 0.2 }}
                        className="w-20 h-20 rounded-full bg-green-100 flex items-center justify-center mx-auto mb-6"
                      >
                        <CheckCircle2 className="h-10 w-10 text-green-600" />
                      </motion.div>
                      <h3 className="text-2xl font-bold text-gray-800 mb-2">Mensagem Enviada!</h3>
                      <p className="text-muted-foreground mb-6">
                        Obrigado pelo contato. Responderemos em ate 24 horas uteis.
                      </p>
                      <Button onClick={() => setSubmitStatus('idle')}>
                        Enviar outra mensagem
                      </Button>
                    </motion.div>
                  ) : (
                    <form onSubmit={handleSubmit} className="space-y-5">
                      <div className="grid md:grid-cols-2 gap-5">
                        <div className="space-y-2">
                          <Label htmlFor="name" className="text-gray-700">
                            Nome Completo <span className="text-red-500">*</span>
                          </Label>
                          <Input
                            id="name"
                            name="name"
                            value={formData.name}
                            onChange={handleChange}
                            placeholder="Seu nome"
                            className={cn(
                              "h-11",
                              errors.name && "border-red-500 focus-visible:ring-red-500"
                            )}
                          />
                          {errors.name && (
                            <p className="text-sm text-red-500 flex items-center gap-1">
                              <AlertCircle className="h-3.5 w-3.5" />
                              {errors.name}
                            </p>
                          )}
                        </div>

                        <div className="space-y-2">
                          <Label htmlFor="email" className="text-gray-700">
                            Email <span className="text-red-500">*</span>
                          </Label>
                          <Input
                            id="email"
                            name="email"
                            type="email"
                            value={formData.email}
                            onChange={handleChange}
                            placeholder="seu@email.com"
                            className={cn(
                              "h-11",
                              errors.email && "border-red-500 focus-visible:ring-red-500"
                            )}
                          />
                          {errors.email && (
                            <p className="text-sm text-red-500 flex items-center gap-1">
                              <AlertCircle className="h-3.5 w-3.5" />
                              {errors.email}
                            </p>
                          )}
                        </div>
                      </div>

                      <div className="grid md:grid-cols-2 gap-5">
                        <div className="space-y-2">
                          <Label htmlFor="phone" className="text-gray-700">
                            Telefone
                          </Label>
                          <Input
                            id="phone"
                            name="phone"
                            value={formData.phone}
                            onChange={handleChange}
                            placeholder="(00) 00000-0000"
                            className="h-11"
                          />
                        </div>

                        <div className="space-y-2">
                          <Label htmlFor="institution" className="text-gray-700">
                            Instituicao
                          </Label>
                          <Input
                            id="institution"
                            name="institution"
                            value={formData.institution}
                            onChange={handleChange}
                            placeholder="Nome do hospital ou organizacao"
                            className="h-11"
                          />
                        </div>
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="subject" className="text-gray-700">
                          Assunto
                        </Label>
                        <select
                          id="subject"
                          name="subject"
                          value={formData.subject}
                          onChange={handleChange}
                          className="w-full h-11 px-3 rounded-md border border-input bg-background text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                        >
                          <option value="">Selecione um assunto</option>
                          <option value="demonstracao">Solicitar Demonstracao</option>
                          <option value="parceria">Parceria Institucional</option>
                          <option value="duvidas">Duvidas Tecnicas</option>
                          <option value="suporte">Suporte ao Usuario</option>
                          <option value="imprensa">Imprensa e Comunicacao</option>
                          <option value="outros">Outros Assuntos</option>
                        </select>
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="message" className="text-gray-700">
                          Mensagem <span className="text-red-500">*</span>
                        </Label>
                        <textarea
                          id="message"
                          name="message"
                          value={formData.message}
                          onChange={handleChange}
                          placeholder="Descreva sua mensagem, duvida ou solicitacao..."
                          rows={5}
                          className={cn(
                            "w-full px-3 py-2 rounded-md border border-input bg-background text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 resize-none",
                            errors.message && "border-red-500 focus-visible:ring-red-500"
                          )}
                        />
                        {errors.message && (
                          <p className="text-sm text-red-500 flex items-center gap-1">
                            <AlertCircle className="h-3.5 w-3.5" />
                            {errors.message}
                          </p>
                        )}
                      </div>

                      {submitStatus === 'error' && (
                        <motion.div
                          initial={{ opacity: 0, y: -10 }}
                          animate={{ opacity: 1, y: 0 }}
                          className="p-4 bg-red-50 border border-red-200 rounded-lg flex items-start gap-3"
                        >
                          <AlertCircle className="h-5 w-5 text-red-600 shrink-0 mt-0.5" />
                          <div>
                            <p className="font-medium text-red-800">Erro ao enviar mensagem</p>
                            <p className="text-sm text-red-700">Por favor, tente novamente ou entre em contato por email.</p>
                          </div>
                        </motion.div>
                      )}

                      <Button
                        type="submit"
                        size="lg"
                        className="w-full h-12 text-base"
                        disabled={isSubmitting}
                      >
                        {isSubmitting ? (
                          <>
                            <Loader2 className="mr-2 h-5 w-5 animate-spin" />
                            Enviando...
                          </>
                        ) : (
                          <>
                            <Send className="mr-2 h-5 w-5" />
                            Enviar Mensagem
                          </>
                        )}
                      </Button>

                      <p className="text-xs text-muted-foreground text-center">
                        Ao enviar, voce concorda com nossa politica de privacidade.
                        Seus dados serao utilizados apenas para responder sua solicitacao.
                      </p>
                    </form>
                  )}
                </CardContent>
              </Card>
            </motion.div>

            {/* Contact Info - 2 columns */}
            <div className="lg:col-span-2 space-y-6">
              {/* Contact Cards */}
              <ContactCard
                icon={Mail}
                title="Email"
                content="contato@sidot.com.br"
                description="Resposta em ate 24 horas uteis"
                href="mailto:contato@sidot.com.br"
                delay={0}
              />

              <ContactCard
                icon={Phone}
                title="Telefone"
                content="(62) 3201-4500"
                description="Seg a Sex, 8h as 18h"
                href="tel:+556232014500"
                delay={0.1}
              />

              <ContactCard
                icon={MapPin}
                title="Endereco"
                content="Av. Anhanguera, 5195 - Setor Coimbra"
                description="Goiania - GO, 74043-011"
                href="https://maps.google.com/?q=Av.+Anhanguera,+5195+-+Setor+Coimbra,+Goiania+-+GO"
                delay={0.2}
              />

              <ContactCard
                icon={Building2}
                title="Instituicao"
                content="SES-GO / Central de Transplantes"
                description="Secretaria de Estado da Saude de Goias"
                delay={0.3}
              />

              {/* Office Hours */}
              <motion.div
                initial={{ opacity: 0, y: 30 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.4 }}
                viewport={{ once: true }}
                className="p-6 bg-gradient-to-br from-primary/10 to-sky-100 rounded-2xl border border-primary/20"
              >
                <div className="flex items-center gap-3 mb-4">
                  <Clock className="h-5 w-5 text-primary" />
                  <h3 className="font-semibold text-gray-800">Horario de Atendimento</h3>
                </div>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Segunda a Sexta</span>
                    <span className="font-medium text-gray-800">08:00 - 18:00</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Sabado</span>
                    <span className="font-medium text-gray-800">08:00 - 12:00</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Domingo e Feriados</span>
                    <span className="font-medium text-gray-500">Fechado</span>
                  </div>
                </div>
                <div className="mt-4 pt-4 border-t border-primary/20">
                  <p className="text-xs text-muted-foreground">
                    <strong className="text-primary">Plantao 24h:</strong> Para casos urgentes relacionados a captacao,
                    ligue para (62) 99999-0000
                  </p>
                </div>
              </motion.div>
            </div>
          </div>

          {/* Quick Links */}
          <motion.div
            initial={{ opacity: 0, y: 40 }}
            whileInView={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            viewport={{ once: true }}
            className="mt-16"
          >
            <div className="text-center mb-8">
              <h2 className="text-2xl font-bold text-gray-800 mb-2">Links Rapidos</h2>
              <p className="text-muted-foreground">Acesse informacoes importantes sobre o SIDOT</p>
            </div>

            <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-4">
              <QuickLinkCard
                icon={FileText}
                title="Sobre o Projeto"
                description="Conheca nossa solucao"
                href="/about"
                color="blue"
                delay={0}
              />
              <QuickLinkCard
                icon={HelpCircle}
                title="FAQ"
                description="Perguntas frequentes"
                href="/about#faq"
                color="purple"
                delay={0.1}
              />
              <QuickLinkCard
                icon={Sparkles}
                title="Demonstracao"
                description="Agende uma apresentacao"
                href="/login"
                color="green"
                delay={0.2}
              />
              <QuickLinkCard
                icon={Building2}
                title="Para Hospitais"
                description="Informacoes para parceiros"
                href="/about#integrations"
                color="orange"
                delay={0.3}
              />
            </div>
          </motion.div>

          {/* Team Section */}
          <motion.div
            initial={{ opacity: 0, y: 40 }}
            whileInView={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            viewport={{ once: true }}
            className="mt-16"
          >
            <Card className="border-none shadow-xl overflow-hidden">
              <CardHeader className="bg-gradient-to-r from-sky-50 to-primary/10 border-b">
                <div className="flex items-center gap-3">
                  <Users className="h-6 w-6 text-primary" />
                  <div>
                    <CardTitle>Equipe de Contato</CardTitle>
                    <CardDescription>Entre em contato diretamente com nossa equipe</CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="p-6">
                <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
                  <TeamMemberCard
                    name="Dr. Ricardo Oliveira"
                    role="Coordenador do Projeto"
                    email="ricardo.oliveira@ses.go.gov.br"
                    delay={0}
                  />
                  <TeamMemberCard
                    name="Dra. Ana Beatriz"
                    role="Diretora de Transplantes"
                    email="ana.beatriz@ses.go.gov.br"
                    delay={0.1}
                  />
                  <TeamMemberCard
                    name="Carlos Eduardo"
                    role="Suporte Tecnico"
                    email="suporte@sidot.com.br"
                    delay={0.2}
                  />
                </div>
              </CardContent>
            </Card>
          </motion.div>

          {/* Map Section */}
          <motion.div
            initial={{ opacity: 0, y: 40 }}
            whileInView={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            viewport={{ once: true }}
            className="mt-16"
          >
            <Card className="border-none shadow-xl overflow-hidden">
              <CardHeader>
                <div className="flex items-center gap-3">
                  <MapPin className="h-6 w-6 text-primary" />
                  <div>
                    <CardTitle>Nossa Localizacao</CardTitle>
                    <CardDescription>Secretaria de Estado da Saude de Goias</CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="p-0">
                <div className="relative h-80 bg-gray-100">
                  <iframe
                    src="https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3821.8726813376985!2d-49.27102472393783!3d-16.693444784127356!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x935ef11357cd11c5%3A0x3c3d3e3e3e3e3e3e!2sAv.%20Anhanguera%2C%205195%20-%20St.%20Coimbra%2C%20Goi%C3%A2nia%20-%20GO!5e0!3m2!1spt-BR!2sbr!4v1234567890"
                    width="100%"
                    height="100%"
                    style={{ border: 0 }}
                    allowFullScreen
                    loading="lazy"
                    referrerPolicy="no-referrer-when-downgrade"
                    className="absolute inset-0"
                  />
                </div>
                <div className="p-4 bg-gray-50 border-t flex items-center justify-between">
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <MapPin className="h-4 w-4" />
                    Av. Anhanguera, 5195 - Setor Coimbra, Goiania - GO
                  </div>
                  <Button variant="outline" size="sm" asChild>
                    <a
                      href="https://maps.google.com/?q=Av.+Anhanguera,+5195+-+Setor+Coimbra,+Goiania+-+GO"
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      <ExternalLink className="h-4 w-4 mr-2" />
                      Abrir no Maps
                    </a>
                  </Button>
                </div>
              </CardContent>
            </Card>
          </motion.div>

          {/* Social Links */}
          <motion.div
            initial={{ opacity: 0, y: 40 }}
            whileInView={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            viewport={{ once: true }}
            className="mt-16 text-center"
          >
            <h3 className="text-lg font-semibold text-gray-800 mb-4">Siga-nos nas Redes Sociais</h3>
            <div className="flex justify-center gap-4">
              {[
                { icon: Globe, label: 'Website', href: 'https://saude.go.gov.br' },
                { icon: Linkedin, label: 'LinkedIn', href: '#' },
                { icon: Twitter, label: 'Twitter', href: '#' },
                { icon: Github, label: 'GitHub', href: '#' },
              ].map((social, i) => (
                <motion.a
                  key={i}
                  href={social.href}
                  target="_blank"
                  rel="noopener noreferrer"
                  whileHover={{ y: -3, scale: 1.1 }}
                  className="p-3 bg-white rounded-xl border border-gray-200 shadow-sm hover:shadow-md hover:border-primary/30 transition-all"
                  title={social.label}
                >
                  <social.icon className="h-5 w-5 text-gray-600 hover:text-primary transition-colors" />
                </motion.a>
              ))}
            </div>
          </motion.div>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gradient-to-r from-primary to-sky-700 text-white py-12 mt-16">
        <div className="container mx-auto px-4 max-w-6xl">
          <div className="grid md:grid-cols-3 gap-8 mb-8">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
            >
              <div className="flex items-center gap-2 mb-4">
                <Heart className="h-6 w-6" />
                <span className="text-xl font-bold">SIDOT</span>
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
              <h4 className="font-semibold mb-4">Contato Rapido</h4>
              <div className="space-y-2 text-white/80 text-sm">
                <p className="flex items-center gap-2">
                  <Mail className="h-4 w-4" />
                  contato@sidot.com.br
                </p>
                <p className="flex items-center gap-2">
                  <Phone className="h-4 w-4" />
                  (62) 3201-4500
                </p>
                <p className="flex items-center gap-2">
                  <MapPin className="h-4 w-4" />
                  Goiania - GO
                </p>
              </div>
            </motion.div>
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
              viewport={{ once: true }}
            >
              <h4 className="font-semibold mb-4">Links Uteis</h4>
              <div className="space-y-2 text-white/80 text-sm">
                <Link href="/about" className="flex items-center gap-2 hover:text-white transition-colors">
                  <ChevronRight className="h-4 w-4" />
                  Sobre o Projeto
                </Link>
                <Link href="/login" className="flex items-center gap-2 hover:text-white transition-colors">
                  <ChevronRight className="h-4 w-4" />
                  Acessar Sistema
                </Link>
                <a href="https://saude.go.gov.br" target="_blank" rel="noopener noreferrer" className="flex items-center gap-2 hover:text-white transition-colors">
                  <ChevronRight className="h-4 w-4" />
                  SES-GO
                </a>
              </div>
            </motion.div>
          </div>
          <div className="border-t border-white/20 pt-8 text-center">
            <p className="text-white/60 text-sm">
              SIDOT - Sistema de Captacao de Corneas
            </p>
            <p className="text-white/40 text-sm mt-2">
              Desenvolvido para o bem da saude publica
            </p>
          </div>
        </div>
      </footer>
    </div>
  );
}
