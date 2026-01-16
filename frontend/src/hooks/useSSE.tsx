'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { getAccessToken } from '@/lib/api';
import type { SSENotificationEvent } from '@/types';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

interface UseSSEOptions {
  onNotification?: (event: SSENotificationEvent) => void;
  enabled?: boolean;
}

interface UseSSEReturn {
  isConnected: boolean;
  pendingCount: number;
  lastEvent: SSENotificationEvent | null;
  soundEnabled: boolean;
  toggleSound: () => void;
}

export function useSSE(options: UseSSEOptions = {}): UseSSEReturn {
  const { onNotification, enabled = true } = options;
  const [isConnected, setIsConnected] = useState(false);
  const [pendingCount, setPendingCount] = useState(0);
  const [lastEvent, setLastEvent] = useState<SSENotificationEvent | null>(null);
  const [soundEnabled, setSoundEnabled] = useState(() => {
    if (typeof window === 'undefined') return true;
    const stored = localStorage.getItem('vitalconnect_sound_enabled');
    return stored === null ? true : stored === 'true';
  });

  const eventSourceRef = useRef<EventSource | null>(null);
  const audioContextRef = useRef<AudioContext | null>(null);

  const playAlertSound = useCallback(() => {
    if (!soundEnabled) return;

    try {
      if (!audioContextRef.current) {
        audioContextRef.current = new AudioContext();
      }

      const ctx = audioContextRef.current;

      // Resume context if suspended (browser autoplay policy)
      if (ctx.state === 'suspended') {
        ctx.resume();
      }

      // Play a more noticeable alert: two-tone "ding-ding"
      const playTone = (frequency: number, startTime: number, duration: number) => {
        const oscillator = ctx.createOscillator();
        const gainNode = ctx.createGain();

        oscillator.connect(gainNode);
        gainNode.connect(ctx.destination);

        oscillator.type = 'sine';
        oscillator.frequency.setValueAtTime(frequency, startTime);

        gainNode.gain.setValueAtTime(0.5, startTime);
        gainNode.gain.exponentialRampToValueAtTime(0.01, startTime + duration);

        oscillator.start(startTime);
        oscillator.stop(startTime + duration);
      };

      // First tone (higher pitch)
      playTone(1200, ctx.currentTime, 0.15);
      // Second tone (even higher, after short pause)
      playTone(1500, ctx.currentTime + 0.2, 0.15);
      // Third tone (highest)
      playTone(1800, ctx.currentTime + 0.4, 0.2);

    } catch (err) {
      console.warn('Could not play alert sound:', err);
    }
  }, [soundEnabled]);

  const toggleSound = useCallback(() => {
    setSoundEnabled((prev) => {
      const newValue = !prev;
      localStorage.setItem('vitalconnect_sound_enabled', String(newValue));

      // Initialize AudioContext on user interaction (required by browser policy)
      if (newValue && !audioContextRef.current) {
        audioContextRef.current = new AudioContext();
      }

      // Play a test beep when enabling sound
      if (newValue && audioContextRef.current) {
        const ctx = audioContextRef.current;
        if (ctx.state === 'suspended') {
          ctx.resume();
        }

        // Short confirmation beep
        const oscillator = ctx.createOscillator();
        const gainNode = ctx.createGain();
        oscillator.connect(gainNode);
        gainNode.connect(ctx.destination);
        oscillator.type = 'sine';
        oscillator.frequency.setValueAtTime(800, ctx.currentTime);
        gainNode.gain.setValueAtTime(0.3, ctx.currentTime);
        gainNode.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.1);
        oscillator.start(ctx.currentTime);
        oscillator.stop(ctx.currentTime + 0.1);
      }

      return newValue;
    });
  }, []);

  useEffect(() => {
    if (!enabled) return;

    const token = getAccessToken();
    if (!token) return;

    const url = `${API_URL}/notifications/stream?token=${encodeURIComponent(token)}`;
    const eventSource = new EventSource(url);
    eventSourceRef.current = eventSource;

    eventSource.onopen = () => {
      setIsConnected(true);
    };

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as SSENotificationEvent;
        setLastEvent(data);

        if (data.type === 'new_occurrence') {
          setPendingCount((prev) => prev + 1);
          playAlertSound();
        }

        onNotification?.(data);
      } catch (error) {
        console.error('Failed to parse SSE event:', error);
      }
    };

    eventSource.onerror = () => {
      setIsConnected(false);
      eventSource.close();

      // Reconnect after 5 seconds
      setTimeout(() => {
        if (eventSourceRef.current === eventSource) {
          eventSourceRef.current = null;
        }
      }, 5000);
    };

    return () => {
      eventSource.close();
      eventSourceRef.current = null;
      setIsConnected(false);
    };
  }, [enabled, onNotification, playAlertSound]);

  return {
    isConnected,
    pendingCount,
    lastEvent,
    soundEnabled,
    toggleSound,
  };
}

export function resetPendingCount(
  setPendingCount: React.Dispatch<React.SetStateAction<number>>
) {
  setPendingCount(0);
}
