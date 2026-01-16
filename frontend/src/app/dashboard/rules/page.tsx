'use client';

import { useState } from 'react';
import { Plus, FileText, Trash2, Edit } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useTriagemRules, useCreateTriagemRule, useUpdateTriagemRule, useToggleTriagemRule, useDeleteTriagemRule } from '@/hooks';
import type { TriagemRule, CreateTriagemRuleInput, UpdateTriagemRuleInput, TriagemRuleConfig } from '@/types';

type RuleType = 'idade_maxima' | 'janela_horas' | 'causas_excludentes' | 'identificacao';

function getRuleType(config: TriagemRuleConfig): RuleType | null {
  if (config.idade_maxima !== undefined) return 'idade_maxima';
  if (config.janela_horas !== undefined) return 'janela_horas';
  if (config.causas_excludentes !== undefined && config.causas_excludentes.length > 0) return 'causas_excludentes';
  if (config.identificacao_desconhecida_inelegivel !== undefined) return 'identificacao';
  return null;
}

function getRuleDisplayValue(config: TriagemRuleConfig): string {
  if (config.idade_maxima !== undefined) return `Idade maxima: ${config.idade_maxima} anos`;
  if (config.janela_horas !== undefined) return `Janela: ${config.janela_horas} horas`;
  if (config.causas_excludentes !== undefined && config.causas_excludentes.length > 0) {
    return `CIDs: ${config.causas_excludentes.slice(0, 3).join(', ')}${config.causas_excludentes.length > 3 ? '...' : ''}`;
  }
  if (config.identificacao_desconhecida_inelegivel) return 'Identificacao desconhecida inelegivel';
  return 'Configuracao vazia';
}

function getRuleTypeBadge(config: TriagemRuleConfig) {
  const type = getRuleType(config);
  const variants: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
    idade_maxima: { label: 'Idade', variant: 'default' },
    janela_horas: { label: 'Tempo', variant: 'secondary' },
    causas_excludentes: { label: 'CID', variant: 'destructive' },
    identificacao: { label: 'ID', variant: 'outline' },
  };

  if (!type) return <Badge variant="outline">Outro</Badge>;
  const { label, variant } = variants[type];
  return <Badge variant={variant}>{label}</Badge>;
}

export default function RulesPage() {
  const { data, isLoading, error } = useTriagemRules();
  const toggleMutation = useToggleTriagemRule();
  const deleteMutation = useDeleteTriagemRule();
  const createMutation = useCreateTriagemRule();
  const updateMutation = useUpdateTriagemRule();

  const [deleteRuleId, setDeleteRuleId] = useState<string | null>(null);
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [editingRule, setEditingRule] = useState<TriagemRule | null>(null);

  // Form state
  const [formName, setFormName] = useState('');
  const [formDescription, setFormDescription] = useState('');
  const [formPriority, setFormPriority] = useState('100');
  const [formRuleType, setFormRuleType] = useState<RuleType>('idade_maxima');
  const [formIdadeMaxima, setFormIdadeMaxima] = useState('70');
  const [formJanelaHoras, setFormJanelaHoras] = useState('6');
  const [formCausasExcludentes, setFormCausasExcludentes] = useState('');

  const rules = data?.data ?? [];

  const handleToggle = async (id: string, currentAtivo: boolean) => {
    try {
      await toggleMutation.mutateAsync({ id, ativo: !currentAtivo });
    } catch (err) {
      console.error('Error toggling rule:', err);
    }
  };

  const handleDelete = async () => {
    if (!deleteRuleId) return;
    try {
      await deleteMutation.mutateAsync(deleteRuleId);
      setDeleteRuleId(null);
    } catch (err) {
      console.error('Error deleting rule:', err);
    }
  };

  const openCreateForm = () => {
    setEditingRule(null);
    setFormName('');
    setFormDescription('');
    setFormPriority('100');
    setFormRuleType('idade_maxima');
    setFormIdadeMaxima('70');
    setFormJanelaHoras('6');
    setFormCausasExcludentes('');
    setIsFormOpen(true);
  };

  const openEditForm = (rule: TriagemRule) => {
    setEditingRule(rule);
    setFormName(rule.nome);
    setFormDescription(rule.descricao || '');
    setFormPriority(String(rule.prioridade));

    const type = getRuleType(rule.regras);
    setFormRuleType(type || 'idade_maxima');
    setFormIdadeMaxima(String(rule.regras.idade_maxima ?? 70));
    setFormJanelaHoras(String(rule.regras.janela_horas ?? 6));
    setFormCausasExcludentes(rule.regras.causas_excludentes?.join(', ') ?? '');

    setIsFormOpen(true);
  };

  const handleSubmit = async () => {
    const regras: TriagemRuleConfig = {};

    switch (formRuleType) {
      case 'idade_maxima':
        regras.idade_maxima = parseInt(formIdadeMaxima, 10);
        break;
      case 'janela_horas':
        regras.janela_horas = parseInt(formJanelaHoras, 10);
        break;
      case 'causas_excludentes':
        regras.causas_excludentes = formCausasExcludentes.split(',').map(s => s.trim()).filter(Boolean);
        break;
      case 'identificacao':
        regras.identificacao_desconhecida_inelegivel = true;
        break;
    }

    try {
      if (editingRule) {
        const input: UpdateTriagemRuleInput = {
          nome: formName,
          descricao: formDescription || undefined,
          regras,
          prioridade: parseInt(formPriority, 10),
        };
        await updateMutation.mutateAsync({ id: editingRule.id, input });
      } else {
        const input: CreateTriagemRuleInput = {
          nome: formName,
          descricao: formDescription || undefined,
          regras,
          prioridade: parseInt(formPriority, 10),
        };
        await createMutation.mutateAsync(input);
      }
      setIsFormOpen(false);
    } catch (err) {
      console.error('Error saving rule:', err);
    }
  };

  if (error) {
    return (
      <div className="p-6">
        <Card>
          <CardContent className="pt-6">
            <p className="text-destructive">Erro ao carregar regras de triagem</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Regras de Triagem</h1>
          <p className="text-muted-foreground">
            Configure as regras que determinam elegibilidade para doacao de corneas
          </p>
        </div>
        <Button onClick={openCreateForm}>
          <Plus className="h-4 w-4 mr-2" />
          Nova Regra
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Regras Ativas
          </CardTitle>
          <CardDescription>
            {rules.length} regra{rules.length !== 1 ? 's' : ''} configurada{rules.length !== 1 ? 's' : ''}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <div key={i} className="flex items-center space-x-4">
                  <Skeleton className="h-12 w-12 rounded" />
                  <div className="space-y-2">
                    <Skeleton className="h-4 w-[200px]" />
                    <Skeleton className="h-4 w-[150px]" />
                  </div>
                </div>
              ))}
            </div>
          ) : rules.length === 0 ? (
            <div className="text-center py-10">
              <FileText className="mx-auto h-12 w-12 text-muted-foreground" />
              <h3 className="mt-4 text-lg font-semibold">Nenhuma regra configurada</h3>
              <p className="mt-2 text-sm text-muted-foreground">
                Crie sua primeira regra para definir os criterios de triagem.
              </p>
              <Button className="mt-4" onClick={openCreateForm}>
                <Plus className="h-4 w-4 mr-2" />
                Criar Regra
              </Button>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Nome</TableHead>
                  <TableHead>Tipo</TableHead>
                  <TableHead>Configuracao</TableHead>
                  <TableHead>Prioridade</TableHead>
                  <TableHead>Ativo</TableHead>
                  <TableHead className="text-right">Acoes</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {rules.map((rule) => (
                  <TableRow key={rule.id}>
                    <TableCell className="font-medium">{rule.nome}</TableCell>
                    <TableCell>{getRuleTypeBadge(rule.regras)}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {getRuleDisplayValue(rule.regras)}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{rule.prioridade}</Badge>
                    </TableCell>
                    <TableCell>
                      <Switch
                        checked={rule.ativo}
                        onCheckedChange={() => handleToggle(rule.id, rule.ativo)}
                        disabled={toggleMutation.isPending}
                      />
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => openEditForm(rule)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => setDeleteRuleId(rule.id)}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={!!deleteRuleId} onOpenChange={() => setDeleteRuleId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Excluir Regra</AlertDialogTitle>
            <AlertDialogDescription>
              Tem certeza que deseja excluir esta regra? Esta acao nao pode ser desfeita.
              A regra sera desativada e nao sera mais aplicada na triagem.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancelar</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Excluir
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Create/Edit Form Dialog */}
      <Dialog open={isFormOpen} onOpenChange={setIsFormOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>{editingRule ? 'Editar Regra' : 'Nova Regra'}</DialogTitle>
            <DialogDescription>
              {editingRule
                ? 'Modifique os parametros da regra de triagem.'
                : 'Configure uma nova regra para o motor de triagem.'}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Nome</Label>
              <Input
                id="name"
                value={formName}
                onChange={(e) => setFormName(e.target.value)}
                placeholder="Ex: Idade Limite 70 anos"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="description">Descricao (opcional)</Label>
              <Textarea
                id="description"
                value={formDescription}
                onChange={(e) => setFormDescription(e.target.value)}
                placeholder="Descreva o proposito desta regra..."
                rows={2}
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="type">Tipo de Regra</Label>
                <Select value={formRuleType} onValueChange={(v) => setFormRuleType(v as RuleType)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="idade_maxima">Idade Maxima</SelectItem>
                    <SelectItem value="janela_horas">Janela de Tempo</SelectItem>
                    <SelectItem value="causas_excludentes">CIDs Excludentes</SelectItem>
                    <SelectItem value="identificacao">Identificacao Desconhecida</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="priority">Prioridade</Label>
                <Input
                  id="priority"
                  type="number"
                  min="1"
                  max="1000"
                  value={formPriority}
                  onChange={(e) => setFormPriority(e.target.value)}
                />
              </div>
            </div>

            {/* Dynamic fields based on rule type */}
            {formRuleType === 'idade_maxima' && (
              <div className="grid gap-2">
                <Label htmlFor="idade">Idade Maxima (anos)</Label>
                <Input
                  id="idade"
                  type="number"
                  min="0"
                  max="120"
                  value={formIdadeMaxima}
                  onChange={(e) => setFormIdadeMaxima(e.target.value)}
                  placeholder="70"
                />
                <p className="text-xs text-muted-foreground">
                  Pacientes com idade acima deste valor serao inelegiveis
                </p>
              </div>
            )}

            {formRuleType === 'janela_horas' && (
              <div className="grid gap-2">
                <Label htmlFor="janela">Janela de Tempo (horas)</Label>
                <Input
                  id="janela"
                  type="number"
                  min="1"
                  max="48"
                  value={formJanelaHoras}
                  onChange={(e) => setFormJanelaHoras(e.target.value)}
                  placeholder="6"
                />
                <p className="text-xs text-muted-foreground">
                  Tempo maximo apos obito para considerar elegivel
                </p>
              </div>
            )}

            {formRuleType === 'causas_excludentes' && (
              <div className="grid gap-2">
                <Label htmlFor="cids">CIDs Excludentes</Label>
                <Textarea
                  id="cids"
                  value={formCausasExcludentes}
                  onChange={(e) => setFormCausasExcludentes(e.target.value)}
                  placeholder="A00, B20, C00 (separados por virgula)"
                  rows={3}
                />
                <p className="text-xs text-muted-foreground">
                  Codigos CID que tornam o paciente inelegivel (separados por virgula)
                </p>
              </div>
            )}

            {formRuleType === 'identificacao' && (
              <div className="p-4 bg-muted rounded-lg">
                <p className="text-sm">
                  Esta regra marca automaticamente pacientes com identificacao desconhecida
                  como <strong>inelegiveis</strong> para doacao.
                </p>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsFormOpen(false)}>
              Cancelar
            </Button>
            <Button
              onClick={handleSubmit}
              disabled={!formName || createMutation.isPending || updateMutation.isPending}
            >
              {createMutation.isPending || updateMutation.isPending
                ? 'Salvando...'
                : editingRule ? 'Salvar' : 'Criar'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
