'use client';

import { useState, useEffect, useCallback } from 'react';
import { api } from '@/lib/api';

// Firebase configuration - Replace with your Firebase project config
// Get from Firebase Console > Project Settings > General > Your apps > Firebase SDK snippet
const FIREBASE_CONFIG = {
  apiKey: process.env.NEXT_PUBLIC_FIREBASE_API_KEY || '',
  authDomain: process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN || '',
  projectId: process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID || '',
  storageBucket: process.env.NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET || '',
  messagingSenderId: process.env.NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID || '',
  appId: process.env.NEXT_PUBLIC_FIREBASE_APP_ID || '',
};

// VAPID key for web push - Get from Firebase Console > Project Settings > Cloud Messaging > Web Push certificates
const VAPID_KEY = process.env.NEXT_PUBLIC_FIREBASE_VAPID_KEY || '';

export type PushPermissionState = 'granted' | 'denied' | 'default' | 'unsupported';

export interface PushNotificationState {
  isSupported: boolean;
  permission: PushPermissionState;
  isSubscribed: boolean;
  isLoading: boolean;
  error: string | null;
  token: string | null;
}

export function usePushNotifications() {
  const [state, setState] = useState<PushNotificationState>({
    isSupported: false,
    permission: 'default',
    isSubscribed: false,
    isLoading: true,
    error: null,
    token: null,
  });

  // Check if push notifications are supported
  useEffect(() => {
    const checkSupport = async () => {
      const isSupported =
        typeof window !== 'undefined' &&
        'serviceWorker' in navigator &&
        'PushManager' in window &&
        'Notification' in window;

      if (!isSupported) {
        setState((prev) => ({
          ...prev,
          isSupported: false,
          isLoading: false,
          error: 'Push notifications are not supported in this browser',
        }));
        return;
      }

      // Get current permission state
      const permission = Notification.permission as PushPermissionState;

      // Check if already subscribed
      let isSubscribed = false;
      let token: string | null = null;

      if (permission === 'granted') {
        try {
          const registration = await navigator.serviceWorker.ready;
          const subscription = await registration.pushManager.getSubscription();
          isSubscribed = !!subscription;
          // Note: For FCM, we need to use Firebase messaging to get the token
        } catch (e) {
          console.error('Error checking subscription:', e);
        }
      }

      setState((prev) => ({
        ...prev,
        isSupported: true,
        permission,
        isSubscribed,
        isLoading: false,
        token,
      }));
    };

    checkSupport();
  }, []);

  // Register service worker
  const registerServiceWorker = useCallback(async () => {
    if (!('serviceWorker' in navigator)) {
      throw new Error('Service workers not supported');
    }

    const registration = await navigator.serviceWorker.register('/sw.js', {
      scope: '/',
    });

    // Wait for the service worker to be ready
    await navigator.serviceWorker.ready;

    return registration;
  }, []);

  // Request permission and subscribe
  const subscribe = useCallback(async () => {
    if (!state.isSupported) {
      setState((prev) => ({ ...prev, error: 'Push notifications not supported' }));
      return false;
    }

    setState((prev) => ({ ...prev, isLoading: true, error: null }));

    try {
      // Register service worker first
      await registerServiceWorker();

      // Request permission
      const permission = await Notification.requestPermission();

      if (permission !== 'granted') {
        setState((prev) => ({
          ...prev,
          permission: permission as PushPermissionState,
          isLoading: false,
          error: 'Permission denied for push notifications',
        }));
        return false;
      }

      // For FCM, we would initialize Firebase here and get the token
      // Since we're using a simple approach, we'll use the native PushManager
      const registration = await navigator.serviceWorker.ready;

      // Subscribe to push (using VAPID key if available)
      let subscription: PushSubscription;
      if (VAPID_KEY) {
        subscription = await registration.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: urlBase64ToUint8Array(VAPID_KEY),
        });
      } else {
        // Fallback for testing without VAPID
        subscription = await registration.pushManager.subscribe({
          userVisibleOnly: true,
        });
      }

      // Get the token (endpoint for native push, or FCM token)
      const token = subscription.endpoint;

      // Send subscription to backend
      try {
        await api.post('/push/subscribe', {
          token: token,
          platform: 'web',
          user_agent: navigator.userAgent,
        });
      } catch (e) {
        console.error('Failed to register subscription on server:', e);
        // Continue anyway - local subscription is still valid
      }

      setState((prev) => ({
        ...prev,
        permission: 'granted',
        isSubscribed: true,
        isLoading: false,
        token,
      }));

      return true;
    } catch (error) {
      console.error('Push subscription error:', error);
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error: error instanceof Error ? error.message : 'Failed to subscribe',
      }));
      return false;
    }
  }, [state.isSupported, registerServiceWorker]);

  // Unsubscribe
  const unsubscribe = useCallback(async () => {
    setState((prev) => ({ ...prev, isLoading: true, error: null }));

    try {
      const registration = await navigator.serviceWorker.ready;
      const subscription = await registration.pushManager.getSubscription();

      if (subscription) {
        // Unsubscribe locally
        await subscription.unsubscribe();

        // Remove from backend
        try {
          await api.delete('/push/unsubscribe', {
            data: { token: subscription.endpoint },
          });
        } catch (e) {
          console.error('Failed to unregister subscription on server:', e);
        }
      }

      setState((prev) => ({
        ...prev,
        isSubscribed: false,
        isLoading: false,
        token: null,
      }));

      return true;
    } catch (error) {
      console.error('Push unsubscription error:', error);
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error: error instanceof Error ? error.message : 'Failed to unsubscribe',
      }));
      return false;
    }
  }, []);

  return {
    ...state,
    subscribe,
    unsubscribe,
    isConfigured: !!VAPID_KEY || !!FIREBASE_CONFIG.apiKey,
  };
}

// Helper function to convert VAPID key
function urlBase64ToUint8Array(base64String: string): ArrayBuffer {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');

  const rawData = window.atob(base64);
  const outputArray = new Uint8Array(rawData.length);

  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i);
  }
  return outputArray.buffer;
}

export default usePushNotifications;
