'use client';

import { Bell, BellOff, CheckCircle, XCircle, AlertTriangle } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { usePushNotifications } from '@/hooks/usePushNotifications';

export function PushNotificationSettings() {
  const {
    isSupported,
    permission,
    isSubscribed,
    isLoading,
    error,
    isConfigured,
    subscribe,
    unsubscribe,
  } = usePushNotifications();

  const getStatusBadge = () => {
    if (!isSupported) {
      return (
        <Badge variant="secondary" className="gap-1">
          <XCircle className="h-3 w-3" />
          Nao Suportado
        </Badge>
      );
    }
    if (permission === 'denied') {
      return (
        <Badge variant="destructive" className="gap-1">
          <XCircle className="h-3 w-3" />
          Bloqueado
        </Badge>
      );
    }
    if (isSubscribed) {
      return (
        <Badge variant="default" className="gap-1">
          <CheckCircle className="h-3 w-3" />
          Ativo
        </Badge>
      );
    }
    return (
      <Badge variant="secondary" className="gap-1">
        <AlertTriangle className="h-3 w-3" />
        Inativo
      </Badge>
    );
  };

  const handleToggle = async () => {
    if (isSubscribed) {
      await unsubscribe();
    } else {
      await subscribe();
    }
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            {isSubscribed ? (
              <Bell className="h-5 w-5 text-primary" />
            ) : (
              <BellOff className="h-5 w-5 text-muted-foreground" />
            )}
            <CardTitle className="text-base">Notificacoes Push</CardTitle>
          </div>
          {getStatusBadge()}
        </div>
        <CardDescription>
          Receba alertas em tempo real no navegador, mesmo quando o app estiver fechado
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {!isSupported && (
          <p className="text-sm text-muted-foreground">
            Seu navegador nao suporta notificacoes push. Tente usar Chrome, Firefox ou Edge.
          </p>
        )}

        {isSupported && permission === 'denied' && (
          <div className="text-sm text-destructive bg-destructive/10 p-3 rounded-lg">
            <p className="font-medium">Notificacoes bloqueadas</p>
            <p className="mt-1 text-muted-foreground">
              Voce bloqueou as notificacoes para este site. Para reativar, acesse as configuracoes
              do seu navegador e permita notificacoes para este site.
            </p>
          </div>
        )}

        {isSupported && permission !== 'denied' && (
          <>
            <Button
              onClick={handleToggle}
              disabled={isLoading || !isConfigured}
              variant={isSubscribed ? 'outline' : 'default'}
              className="w-full"
            >
              {isLoading ? (
                'Processando...'
              ) : isSubscribed ? (
                <>
                  <BellOff className="h-4 w-4 mr-2" />
                  Desativar Notificacoes
                </>
              ) : (
                <>
                  <Bell className="h-4 w-4 mr-2" />
                  Ativar Notificacoes
                </>
              )}
            </Button>

            {!isConfigured && (
              <p className="text-xs text-muted-foreground text-center">
                Configuracao Firebase/VAPID pendente. Contate o administrador.
              </p>
            )}
          </>
        )}

        {error && (
          <p className="text-sm text-destructive">{error}</p>
        )}

        <div className="text-xs text-muted-foreground space-y-1">
          <p>Quando ativo, voce recebera notificacoes para:</p>
          <ul className="list-disc list-inside ml-2">
            <li>Novas ocorrencias de obito elegivel</li>
            <li>Atualizacoes em ocorrencias atribuidas a voce</li>
            <li>Alertas do sistema</li>
          </ul>
        </div>
      </CardContent>
    </Card>
  );
}

export default PushNotificationSettings;
