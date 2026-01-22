'use client';

import { useState, useEffect, useCallback } from 'react';
import {
  Settings,
  Mail,
  MessageSquare,
  Bell,
  Save,
  Eye,
  EyeOff,
  CheckCircle,
  AlertCircle,
  Loader2,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Switch } from '@/components/ui/switch';
import {
  fetchAdminSettings,
  upsertAdminSetting,
  type AdminSystemSetting,
} from '@/lib/api/admin';
import { toast } from 'sonner';

interface SettingField {
  key: string;
  label: string;
  type: 'text' | 'password' | 'number' | 'email';
  placeholder?: string;
  required?: boolean;
  sensitive?: boolean;
}

interface SettingSection {
  key: string;
  title: string;
  description: string;
  icon: React.ElementType;
  iconColor: string;
  fields: SettingField[];
}

const settingSections: SettingSection[] = [
  {
    key: 'smtp',
    title: 'Configuracoes de Email (SMTP)',
    description: 'Configure o servidor SMTP para envio de emails',
    icon: Mail,
    iconColor: 'text-blue-400',
    fields: [
      { key: 'host', label: 'Host SMTP', type: 'text', placeholder: 'smtp.example.com', required: true },
      { key: 'port', label: 'Porta', type: 'number', placeholder: '587', required: true },
      { key: 'user', label: 'Usuario', type: 'email', placeholder: 'email@example.com', required: true },
      { key: 'password', label: 'Senha', type: 'password', placeholder: '********', required: true, sensitive: true },
      { key: 'from_address', label: 'Email de Origem', type: 'email', placeholder: 'noreply@example.com', required: true },
      { key: 'from_name', label: 'Nome de Origem', type: 'text', placeholder: 'SIDOT' },
    ],
  },
  {
    key: 'twilio',
    title: 'Configuracoes de SMS (Twilio)',
    description: 'Configure o Twilio para envio de SMS',
    icon: MessageSquare,
    iconColor: 'text-emerald-400',
    fields: [
      { key: 'account_sid', label: 'Account SID', type: 'text', placeholder: 'AC...', required: true, sensitive: true },
      { key: 'auth_token', label: 'Auth Token', type: 'password', placeholder: '********', required: true, sensitive: true },
      { key: 'from_number', label: 'Numero de Origem', type: 'text', placeholder: '+5511999999999', required: true },
    ],
  },
  {
    key: 'fcm',
    title: 'Configuracoes de Push (FCM)',
    description: 'Configure o Firebase Cloud Messaging para notificacoes push',
    icon: Bell,
    iconColor: 'text-amber-400',
    fields: [
      { key: 'server_key', label: 'Server Key', type: 'password', placeholder: '********', required: true, sensitive: true },
      { key: 'project_id', label: 'Project ID', type: 'text', placeholder: 'my-project-id' },
    ],
  },
];

interface SettingSectionFormProps {
  section: SettingSection;
  setting: AdminSystemSetting | null;
  onSave: (key: string, value: Record<string, unknown>) => Promise<void>;
  isSaving: boolean;
}

function SettingSectionForm({
  section,
  setting,
  onSave,
  isSaving,
}: SettingSectionFormProps) {
  const [formData, setFormData] = useState<Record<string, string>>({});
  const [showPasswords, setShowPasswords] = useState<Record<string, boolean>>({});
  const [isDirty, setIsDirty] = useState(false);

  // Initialize form data from setting
  useEffect(() => {
    if (setting?.value) {
      const initialData: Record<string, string> = {};
      section.fields.forEach((field) => {
        const value = (setting.value as Record<string, unknown>)[field.key];
        initialData[field.key] = value !== undefined ? String(value) : '';
      });
      setFormData(initialData);
      setIsDirty(false);
    } else {
      // Reset to empty
      const emptyData: Record<string, string> = {};
      section.fields.forEach((field) => {
        emptyData[field.key] = '';
      });
      setFormData(emptyData);
      setIsDirty(false);
    }
  }, [setting, section.fields]);

  const handleChange = (fieldKey: string, value: string) => {
    setFormData((prev) => ({ ...prev, [fieldKey]: value }));
    setIsDirty(true);
  };

  const togglePassword = (fieldKey: string) => {
    setShowPasswords((prev) => ({ ...prev, [fieldKey]: !prev[fieldKey] }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validate required fields
    const missingFields = section.fields.filter(
      (field) => field.required && !formData[field.key]?.trim()
    );
    if (missingFields.length > 0) {
      toast.error(`Preencha os campos obrigatorios: ${missingFields.map((f) => f.label).join(', ')}`);
      return;
    }

    // Build value object
    const value: Record<string, unknown> = {};
    section.fields.forEach((field) => {
      const fieldValue = formData[field.key];
      if (fieldValue !== undefined && fieldValue !== '') {
        // Skip masked values (***) to preserve existing encrypted values
        if (field.sensitive && fieldValue === '***') {
          // Keep the original value
          if (setting?.value) {
            value[field.key] = (setting.value as Record<string, unknown>)[field.key];
          }
        } else {
          value[field.key] = field.type === 'number' ? parseInt(fieldValue, 10) : fieldValue;
        }
      }
    });

    await onSave(section.key, value);
    setIsDirty(false);
  };

  const Icon = section.icon;

  // Mask sensitive values that come from the server
  const getMaskedValue = (field: SettingField, value: string) => {
    if (field.sensitive && setting?.is_encrypted && value && value !== '***') {
      // If it's a masked value from server, show ***
      return '***';
    }
    return value;
  };

  return (
    <Card className="bg-slate-800 border-slate-700">
      <CardHeader>
        <div className="flex items-center gap-3">
          <div className={`h-10 w-10 rounded-lg bg-slate-900 flex items-center justify-center`}>
            <Icon className={`h-5 w-5 ${section.iconColor}`} />
          </div>
          <div>
            <CardTitle className="text-white">{section.title}</CardTitle>
            <CardDescription className="text-slate-400">
              {section.description}
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            {section.fields.map((field) => (
              <div key={field.key} className="space-y-2">
                <label className="text-sm font-medium text-slate-300">
                  {field.label}
                  {field.required && <span className="text-red-400 ml-1">*</span>}
                </label>
                <div className="relative">
                  <Input
                    type={
                      field.type === 'password'
                        ? showPasswords[field.key]
                          ? 'text'
                          : 'password'
                        : field.type
                    }
                    value={
                      field.sensitive && setting?.is_encrypted && formData[field.key] === '***'
                        ? '***'
                        : formData[field.key] || ''
                    }
                    onChange={(e) => handleChange(field.key, e.target.value)}
                    placeholder={field.placeholder}
                    className="bg-slate-900 border-slate-700 text-white placeholder:text-slate-500 pr-10"
                  />
                  {field.type === 'password' && (
                    <button
                      type="button"
                      onClick={() => togglePassword(field.key)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-300"
                    >
                      {showPasswords[field.key] ? (
                        <EyeOff className="h-4 w-4" />
                      ) : (
                        <Eye className="h-4 w-4" />
                      )}
                    </button>
                  )}
                </div>
                {field.sensitive && setting?.is_encrypted && (
                  <p className="text-xs text-slate-500">
                    Valor criptografado. Digite um novo valor para alterar.
                  </p>
                )}
              </div>
            ))}
          </div>
          <div className="flex items-center justify-between pt-4 border-t border-slate-700">
            <div className="flex items-center gap-2">
              {setting ? (
                <div className="flex items-center gap-2 text-emerald-400">
                  <CheckCircle className="h-4 w-4" />
                  <span className="text-xs">Configurado</span>
                </div>
              ) : (
                <div className="flex items-center gap-2 text-slate-500">
                  <AlertCircle className="h-4 w-4" />
                  <span className="text-xs">Nao configurado</span>
                </div>
              )}
            </div>
            <Button
              type="submit"
              disabled={isSaving || !isDirty}
              className="gap-2 bg-violet-600 hover:bg-violet-700 disabled:opacity-50"
            >
              {isSaving ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Salvando...
                </>
              ) : (
                <>
                  <Save className="h-4 w-4" />
                  Salvar
                </>
              )}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

export default function SettingsPage() {
  const [settings, setSettings] = useState<AdminSystemSetting[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [savingSection, setSavingSection] = useState<string | null>(null);

  const loadSettings = useCallback(async () => {
    try {
      setIsLoading(true);
      const data = await fetchAdminSettings();
      setSettings(data);
    } catch (err) {
      console.error('Failed to fetch settings:', err);
      toast.error('Falha ao carregar configuracoes');
      setSettings([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadSettings();
  }, [loadSettings]);

  const handleSave = async (key: string, value: Record<string, unknown>) => {
    try {
      setSavingSection(key);
      await upsertAdminSetting(key, {
        value,
        description: settingSections.find((s) => s.key === key)?.description,
        is_encrypted: true, // All settings with sensitive data should be encrypted
      });
      toast.success('Configuracoes salvas com sucesso');
      loadSettings(); // Reload to get updated values
    } catch (err) {
      console.error('Failed to save settings:', err);
      toast.error('Falha ao salvar configuracoes');
    } finally {
      setSavingSection(null);
    }
  };

  const getSettingByKey = (key: string): AdminSystemSetting | null => {
    return settings.find((s) => s.key === key) || null;
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white">
            Configuracoes Globais
          </h1>
          <p className="text-slate-400">
            Gerencie as configuracoes do sistema
          </p>
        </div>
        <div className="flex items-center justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-violet-600 border-t-transparent" />
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center gap-4">
        <div className="h-12 w-12 rounded-lg bg-slate-800 flex items-center justify-center">
          <Settings className="h-6 w-6 text-violet-400" />
        </div>
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white">
            Configuracoes Globais
          </h1>
          <p className="text-slate-400">
            Gerencie as configuracoes de integracao do sistema
          </p>
        </div>
      </div>

      {/* Info Card */}
      <Card className="bg-blue-400/10 border-blue-400/20">
        <CardContent className="py-4">
          <div className="flex items-start gap-3">
            <AlertCircle className="h-5 w-5 text-blue-400 shrink-0 mt-0.5" />
            <div>
              <p className="text-sm font-medium text-blue-400">
                Valores sensiveis sao criptografados
              </p>
              <p className="text-xs text-blue-400/80 mt-1">
                Campos como senhas e tokens sao armazenados de forma segura usando criptografia AES-256.
                Para alterar um valor existente, basta digitar o novo valor.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Settings Sections */}
      <div className="space-y-6">
        {settingSections.map((section) => (
          <SettingSectionForm
            key={section.key}
            section={section}
            setting={getSettingByKey(section.key)}
            onSave={handleSave}
            isSaving={savingSection === section.key}
          />
        ))}
      </div>
    </div>
  );
}
