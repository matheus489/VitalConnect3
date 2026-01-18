'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import {
  ArrowLeft,
  Save,
  Building,
  Users,
  Building2,
  Activity,
  Palette,
  FileText,
  BarChart3,
  Power,
  Upload,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  fetchAdminTenant,
  updateAdminTenant,
  updateAdminTenantTheme,
  toggleAdminTenantStatus,
  uploadTenantAssets,
  type AdminTenantWithMetrics,
  type ThemeConfig,
} from '@/lib/api/admin';
import { TenantThemeEditor } from '@/components/admin/TenantThemeEditor';
import { toast } from 'sonner';

type TabId = 'details' | 'theme' | 'metrics';

interface TabProps {
  id: TabId;
  label: string;
  icon: React.ElementType;
  active: boolean;
  onClick: () => void;
}

function Tab({ id, label, icon: Icon, active, onClick }: TabProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg transition-colors ${
        active
          ? 'bg-violet-600/20 text-violet-400'
          : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200'
      }`}
    >
      <Icon className="h-4 w-4" />
      {label}
    </button>
  );
}

interface MetricCardProps {
  title: string;
  value: number;
  icon: React.ElementType;
  iconColor?: string;
}

function MetricCard({ title, value, icon: Icon, iconColor = 'text-violet-400' }: MetricCardProps) {
  return (
    <Card className="bg-slate-800 border-slate-700">
      <CardContent className="pt-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-slate-400">{title}</p>
            <p className="text-2xl font-bold text-white mt-1">{value}</p>
          </div>
          <div className="h-12 w-12 rounded-lg bg-slate-900 flex items-center justify-center">
            <Icon className={`h-6 w-6 ${iconColor}`} />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export default function TenantEditorPage() {
  const params = useParams();
  const router = useRouter();
  const tenantId = params.id as string;

  const [tenant, setTenant] = useState<AdminTenantWithMetrics | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [isThemeSaving, setIsThemeSaving] = useState(false);
  const [activeTab, setActiveTab] = useState<TabId>('details');

  // Form state for details tab
  const [formName, setFormName] = useState('');
  const [formSlug, setFormSlug] = useState('');

  // File input refs for asset uploads
  const logoInputRef = useRef<HTMLInputElement>(null);
  const faviconInputRef = useRef<HTMLInputElement>(null);

  const loadTenant = useCallback(async () => {
    try {
      setIsLoading(true);
      const data = await fetchAdminTenant(tenantId);
      setTenant(data);
      setFormName(data.name);
      setFormSlug(data.slug);
    } catch (err) {
      console.error('Failed to fetch tenant:', err);
      toast.error('Falha ao carregar tenant');
      // Mock data for development when API is not available
      const mockTenant: AdminTenantWithMetrics = {
        id: tenantId,
        name: 'Tenant de Exemplo',
        slug: 'tenant-exemplo',
        is_active: true,
        theme_config: {},
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        user_count: 0,
        hospital_count: 0,
        occurrence_count: 0,
      };
      setTenant(mockTenant);
      setFormName(mockTenant.name);
      setFormSlug(mockTenant.slug);
    } finally {
      setIsLoading(false);
    }
  }, [tenantId]);

  useEffect(() => {
    loadTenant();
  }, [loadTenant]);

  const handleSaveDetails = async () => {
    if (!formName.trim() || !formSlug.trim()) {
      toast.error('Preencha todos os campos');
      return;
    }

    try {
      setIsSaving(true);
      await updateAdminTenant(tenantId, {
        name: formName.trim(),
        slug: formSlug.trim(),
      });
      toast.success('Tenant atualizado com sucesso');
      loadTenant();
    } catch (err) {
      console.error('Failed to update tenant:', err);
      toast.error('Falha ao atualizar tenant');
    } finally {
      setIsSaving(false);
    }
  };

  const handleSaveTheme = async (themeConfig: ThemeConfig) => {
    try {
      setIsThemeSaving(true);
      await updateAdminTenantTheme(tenantId, themeConfig);
      toast.success('Tema atualizado com sucesso');
      loadTenant();
    } catch (err) {
      console.error('Failed to update theme:', err);
      toast.error('Falha ao atualizar tema');
      throw err;
    } finally {
      setIsThemeSaving(false);
    }
  };

  const handleToggleStatus = async () => {
    if (!tenant) return;

    try {
      await toggleAdminTenantStatus(tenantId);
      toast.success(
        tenant.is_active
          ? 'Tenant desativado com sucesso'
          : 'Tenant ativado com sucesso'
      );
      loadTenant();
    } catch (err) {
      console.error('Failed to toggle tenant status:', err);
      toast.error('Falha ao alterar status do tenant');
    }
  };

  const handleUploadLogo = () => {
    logoInputRef.current?.click();
  };

  const handleUploadFavicon = () => {
    faviconInputRef.current?.click();
  };

  const handleFileUpload = async (
    event: React.ChangeEvent<HTMLInputElement>,
    type: 'logo' | 'favicon'
  ) => {
    const file = event.target.files?.[0];
    if (!file) return;

    // Validate file type
    if (!file.type.startsWith('image/')) {
      toast.error('Por favor, selecione um arquivo de imagem');
      return;
    }

    // Validate file size (max 2MB)
    if (file.size > 2 * 1024 * 1024) {
      toast.error('O arquivo deve ter no maximo 2MB');
      return;
    }

    try {
      const formData = new FormData();
      formData.append(type, file);

      await uploadTenantAssets(tenantId, formData);
      toast.success(`${type === 'logo' ? 'Logo' : 'Favicon'} atualizado com sucesso`);
      loadTenant();
    } catch (err) {
      console.error(`Failed to upload ${type}:`, err);
      toast.error(`Falha ao fazer upload do ${type === 'logo' ? 'logo' : 'favicon'}`);
    }

    // Clear the input
    event.target.value = '';
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-violet-600 border-t-transparent" />
      </div>
    );
  }

  if (!tenant) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <Building className="h-12 w-12 text-slate-600 mb-4" />
        <h3 className="text-lg font-medium text-slate-300">Tenant nao encontrado</h3>
        <Link href="/admin/tenants">
          <Button variant="ghost" className="mt-4 text-violet-400 hover:text-violet-300">
            Voltar para lista
          </Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6 h-full flex flex-col">
      {/* Hidden file inputs for asset uploads */}
      <input
        ref={logoInputRef}
        type="file"
        accept="image/*"
        className="hidden"
        onChange={(e) => handleFileUpload(e, 'logo')}
      />
      <input
        ref={faviconInputRef}
        type="file"
        accept="image/*"
        className="hidden"
        onChange={(e) => handleFileUpload(e, 'favicon')}
      />

      {/* Page Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between flex-shrink-0">
        <div className="flex items-center gap-4">
          <Link href="/admin/tenants">
            <Button
              variant="ghost"
              size="icon"
              className="text-slate-400 hover:text-white hover:bg-slate-800"
            >
              <ArrowLeft className="h-5 w-5" />
            </Button>
          </Link>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold tracking-tight text-white">
                {tenant.name}
              </h1>
              <Badge
                variant={tenant.is_active ? 'default' : 'secondary'}
                className={
                  tenant.is_active
                    ? 'bg-emerald-400/10 text-emerald-400 border-emerald-400/20'
                    : 'bg-slate-600/20 text-slate-400 border-slate-600/20'
                }
              >
                {tenant.is_active ? 'Ativo' : 'Inativo'}
              </Badge>
            </div>
            <p className="text-slate-400 font-mono text-sm">/{tenant.slug}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={handleToggleStatus}
            className={`gap-2 bg-slate-800 border-slate-700 ${
              tenant.is_active
                ? 'text-amber-400 hover:bg-amber-400/10'
                : 'text-emerald-400 hover:bg-emerald-400/10'
            }`}
          >
            <Power className="h-4 w-4" />
            {tenant.is_active ? 'Desativar' : 'Ativar'}
          </Button>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 border-b border-slate-700 pb-4 flex-shrink-0">
        <Tab
          id="details"
          label="Detalhes"
          icon={FileText}
          active={activeTab === 'details'}
          onClick={() => setActiveTab('details')}
        />
        <Tab
          id="theme"
          label="Editor de Tema"
          icon={Palette}
          active={activeTab === 'theme'}
          onClick={() => setActiveTab('theme')}
        />
        <Tab
          id="metrics"
          label="Metricas"
          icon={BarChart3}
          active={activeTab === 'metrics'}
          onClick={() => setActiveTab('metrics')}
        />
      </div>

      {/* Tab Content */}
      <div className="flex-1 min-h-0">
        {activeTab === 'details' && (
          <Card className="bg-slate-800 border-slate-700">
            <CardHeader>
              <CardTitle className="text-white">Informacoes do Tenant</CardTitle>
              <CardDescription className="text-slate-400">
                Edite as informacoes basicas do tenant
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-slate-300">
                  Nome do Tenant
                </label>
                <Input
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  placeholder="Nome do tenant"
                  className="bg-slate-900 border-slate-700 text-white placeholder:text-slate-500"
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-slate-300">
                  Slug (identificador unico)
                </label>
                <Input
                  value={formSlug}
                  onChange={(e) =>
                    setFormSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))
                  }
                  placeholder="slug-do-tenant"
                  className="bg-slate-900 border-slate-700 text-white placeholder:text-slate-500"
                />
                <p className="text-xs text-slate-500">
                  Usado na URL e identificacao. Use apenas letras minusculas, numeros e
                  hifens.
                </p>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-slate-300">ID</label>
                <Input
                  value={tenant.id}
                  disabled
                  className="bg-slate-900 border-slate-700 text-slate-500 font-mono text-sm"
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-slate-300">Criado em</label>
                <Input
                  value={new Date(tenant.created_at).toLocaleString('pt-BR')}
                  disabled
                  className="bg-slate-900 border-slate-700 text-slate-500"
                />
              </div>

              {/* Asset Upload Section */}
              <div className="space-y-4 pt-4 border-t border-slate-700">
                <h3 className="text-sm font-medium text-slate-300">Assets do Tenant</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <label className="text-sm text-slate-400">Logo</label>
                    <div className="flex items-center gap-4">
                      {tenant.logo_url ? (
                        <img
                          src={tenant.logo_url}
                          alt="Logo"
                          className="h-12 w-12 rounded-lg object-cover border border-slate-600"
                        />
                      ) : (
                        <div className="h-12 w-12 rounded-lg bg-slate-900 border border-slate-700 flex items-center justify-center">
                          <Building className="h-6 w-6 text-slate-600" />
                        </div>
                      )}
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={handleUploadLogo}
                        className="gap-2 bg-slate-900 border-slate-700 text-slate-300 hover:bg-slate-800"
                      >
                        <Upload className="h-4 w-4" />
                        Upload
                      </Button>
                    </div>
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm text-slate-400">Favicon</label>
                    <div className="flex items-center gap-4">
                      {tenant.favicon_url ? (
                        <img
                          src={tenant.favicon_url}
                          alt="Favicon"
                          className="h-12 w-12 rounded-lg object-cover border border-slate-600"
                        />
                      ) : (
                        <div className="h-12 w-12 rounded-lg bg-slate-900 border border-slate-700 flex items-center justify-center">
                          <Building className="h-6 w-6 text-slate-600" />
                        </div>
                      )}
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={handleUploadFavicon}
                        className="gap-2 bg-slate-900 border-slate-700 text-slate-300 hover:bg-slate-800"
                      >
                        <Upload className="h-4 w-4" />
                        Upload
                      </Button>
                    </div>
                  </div>
                </div>
              </div>

              <div className="pt-4">
                <Button
                  onClick={handleSaveDetails}
                  disabled={isSaving}
                  className="gap-2 bg-violet-600 hover:bg-violet-700"
                >
                  <Save className="h-4 w-4" />
                  {isSaving ? 'Salvando...' : 'Salvar Alteracoes'}
                </Button>
              </div>
            </CardContent>
          </Card>
        )}

        {activeTab === 'theme' && (
          <TenantThemeEditor
            tenantId={tenantId}
            tenantName={tenant.name}
            initialThemeConfig={tenant.theme_config || {}}
            logoUrl={tenant.logo_url}
            faviconUrl={tenant.favicon_url}
            onSave={handleSaveTheme}
            onUploadLogo={handleUploadLogo}
            onUploadFavicon={handleUploadFavicon}
            isSaving={isThemeSaving}
          />
        )}

        {activeTab === 'metrics' && (
          <div className="space-y-6">
            <div className="grid gap-4 md:grid-cols-3">
              <MetricCard
                title="Total de Usuarios"
                value={tenant.user_count}
                icon={Users}
                iconColor="text-blue-400"
              />
              <MetricCard
                title="Total de Hospitais"
                value={tenant.hospital_count}
                icon={Building2}
                iconColor="text-emerald-400"
              />
              <MetricCard
                title="Total de Ocorrencias"
                value={tenant.occurrence_count}
                icon={Activity}
                iconColor="text-amber-400"
              />
            </div>

            <Card className="bg-slate-800 border-slate-700">
              <CardHeader>
                <CardTitle className="text-white">Historico de Uso</CardTitle>
                <CardDescription className="text-slate-400">
                  Estatisticas detalhadas do tenant ao longo do tempo
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex flex-col items-center justify-center py-8 text-center">
                  <BarChart3 className="h-12 w-12 text-slate-600 mb-4" />
                  <p className="text-sm text-slate-500">
                    Graficos de uso e tendencias serao implementados em uma proxima versao
                  </p>
                </div>
              </CardContent>
            </Card>
          </div>
        )}
      </div>
    </div>
  );
}
