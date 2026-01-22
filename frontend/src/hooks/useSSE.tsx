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
    const stored = localStorage.getItem('sidot_sound_enabled');
    return stored === null ? true : stored === 'true';
  });

  const eventSourceRef = useRef<EventSource | null>(null);
  const audioContextRef = useRef<AudioContext | null>(null);
  const audioUnlockedRef = useRef(false);

  // Play sound using AudioContext (works after user interaction)
  const playAudioContextSound = useCallback(() => {
    console.log('[SIDOT] playAudioContextSound called');
    try {
      // Create or reuse AudioContext
      if (!audioContextRef.current) {
        audioContextRef.current = new AudioContext();
      }
      const ctx = audioContextRef.current;

      // Resume if suspended
      if (ctx.state === 'suspended') {
        ctx.resume();
      }

      const playTone = (frequency: number, startTime: number, duration: number) => {
        const oscillator = ctx.createOscillator();
        const gainNode = ctx.createGain();
        oscillator.connect(gainNode);
        gainNode.connect(ctx.destination);
        oscillator.type = 'sine';
        oscillator.frequency.setValueAtTime(frequency, startTime);
        gainNode.gain.setValueAtTime(0.8, startTime);
        gainNode.gain.exponentialRampToValueAtTime(0.01, startTime + duration);
        oscillator.start(startTime);
        oscillator.stop(startTime + duration);
      };

      // Urgent beep pattern - 4 beeps
      playTone(880, ctx.currentTime, 0.15);
      playTone(880, ctx.currentTime + 0.2, 0.15);
      playTone(1100, ctx.currentTime + 0.5, 0.15);
      playTone(1100, ctx.currentTime + 0.7, 0.2);
      console.log('[SIDOT] AudioContext sound played');
    } catch (err) {
      console.warn('[SIDOT] AudioContext sound failed:', err);
    }
  }, []);

  // Play alert sound
  const playAlertSound = useCallback(() => {
    console.log('[SIDOT] playAlertSound called, soundEnabled:', soundEnabled, 'audioUnlocked:', audioUnlockedRef.current);
    if (!soundEnabled) {
      console.log('[SIDOT] Sound is disabled, skipping');
      return;
    }

    if (!audioUnlockedRef.current) {
      console.log('[SIDOT] Audio not unlocked yet - user needs to interact with page first');
      // Show visual notification that sound needs to be enabled
      return;
    }

    playAudioContextSound();
  }, [soundEnabled, playAudioContextSound]);

  // Unlock audio on user interaction
  const unlockAudio = useCallback(() => {
    if (audioUnlockedRef.current) return;

    console.log('[SIDOT] Unlocking audio...');
    try {
      // Create AudioContext on first interaction
      if (!audioContextRef.current) {
        audioContextRef.current = new AudioContext();
      }

      // Resume if suspended
      if (audioContextRef.current.state === 'suspended') {
        audioContextRef.current.resume();
      }

      audioUnlockedRef.current = true;
      console.log('[SIDOT] Audio unlocked successfully');
    } catch (err) {
      console.warn('[SIDOT] Failed to unlock audio:', err);
    }
  }, []);

  // Toggle sound and unlock audio
  const toggleSound = useCallback(() => {
    // Unlock audio on user interaction
    unlockAudio();

    setSoundEnabled((prev) => {
      const newValue = !prev;
      localStorage.setItem('sidot_sound_enabled', String(newValue));
      console.log('[SIDOT] Sound toggled:', newValue);

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
        gainNode.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.15);
        oscillator.start(ctx.currentTime);
        oscillator.stop(ctx.currentTime + 0.15);
      }

      return newValue;
    });
  }, [unlockAudio]);

  // Add click listener to unlock audio on first interaction
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const handleInteraction = () => {
      unlockAudio();
      // Remove listeners after first interaction
      document.removeEventListener('click', handleInteraction);
      document.removeEventListener('keydown', handleInteraction);
      document.removeEventListener('touchstart', handleInteraction);
    };

    document.addEventListener('click', handleInteraction);
    document.addEventListener('keydown', handleInteraction);
    document.addEventListener('touchstart', handleInteraction);

    console.log('[SIDOT] Audio notification system initialized, waiting for user interaction to unlock');

    return () => {
      document.removeEventListener('click', handleInteraction);
      document.removeEventListener('keydown', handleInteraction);
      document.removeEventListener('touchstart', handleInteraction);
    };
  }, [unlockAudio]);

  // SSE connection
  useEffect(() => {
    if (!enabled) return;

    const token = getAccessToken();
    if (!token) return;

    const url = `${API_URL}/notifications/stream?token=${encodeURIComponent(token)}`;
    const eventSource = new EventSource(url);
    eventSourceRef.current = eventSource;

    eventSource.onopen = () => {
      setIsConnected(true);
      console.log('[SIDOT] SSE connected');
    };

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as SSENotificationEvent;
        console.log('[SIDOT] SSE event received:', data.type, data);
        setLastEvent(data);

        if (data.type === 'new_occurrence') {
          console.log('[SIDOT] New occurrence detected, playing alert sound');
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
      console.log('[SIDOT] SSE disconnected');

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
