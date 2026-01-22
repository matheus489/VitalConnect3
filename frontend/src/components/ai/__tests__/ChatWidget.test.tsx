import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ChatWidget } from '../ChatWidget';
import * as useAIChatModule from '@/hooks/useAIChat';

// Mock the useAIChat hook
vi.mock('@/hooks/useAIChat', () => ({
  useAIChat: vi.fn(),
}));

const mockUseAIChat = vi.mocked(useAIChatModule.useAIChat);

describe('ChatWidget', () => {
  const defaultMockReturn = {
    messages: [],
    isLoading: false,
    isStreaming: false,
    thinkingStep: null,
    currentTool: null,
    sessionId: null,
    pendingConfirmation: null,
    unreadCount: 0,
    sendMessage: vi.fn(),
    confirmAction: vi.fn(),
    clearMessages: vi.fn(),
    loadHistory: vi.fn(),
    startNewSession: vi.fn(),
    markAsRead: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseAIChat.mockReturnValue(defaultMockReturn);
  });

  it('renders floating action button', () => {
    render(<ChatWidget />);

    const button = screen.getByRole('button', { name: /abrir assistente/i });
    expect(button).toBeInTheDocument();
  });

  it('toggles chat panel on FAB click', async () => {
    const user = userEvent.setup();
    render(<ChatWidget />);

    // Click to open
    const fabButton = screen.getByRole('button', { name: /abrir assistente/i });
    await user.click(fabButton);

    // Panel should be open - check for the header text
    expect(screen.getByText('Assistente SIDOT')).toBeInTheDocument();

    // Click to close (the FAB button now shows X icon)
    const closeButton = screen.getByRole('button', { name: /fechar assistente/i });
    await user.click(closeButton);

    // Panel should be closed (translated off screen)
    const panel = screen.getByRole('dialog', { hidden: true });
    expect(panel).toHaveClass('translate-x-full');
  });

  it('displays unread message badge', () => {
    mockUseAIChat.mockReturnValue({
      ...defaultMockReturn,
      unreadCount: 5,
    });

    render(<ChatWidget />);

    const badge = screen.getByText('5');
    expect(badge).toBeInTheDocument();
  });

  it('marks messages as read when opening panel', async () => {
    const markAsRead = vi.fn();
    mockUseAIChat.mockReturnValue({
      ...defaultMockReturn,
      unreadCount: 3,
      markAsRead,
    });

    const user = userEvent.setup();
    render(<ChatWidget />);

    const fabButton = screen.getByRole('button', { name: /abrir assistente/i });
    await user.click(fabButton);

    expect(markAsRead).toHaveBeenCalled();
  });

  it('sends message when form is submitted', async () => {
    const sendMessage = vi.fn();
    mockUseAIChat.mockReturnValue({
      ...defaultMockReturn,
      sendMessage,
    });

    const user = userEvent.setup();
    render(<ChatWidget />);

    // Open panel
    const fabButton = screen.getByRole('button', { name: /abrir assistente/i });
    await user.click(fabButton);

    // Type message
    const input = screen.getByPlaceholderText(/digite sua mensagem/i);
    await user.type(input, 'Hello AI');

    // Submit
    const sendButton = screen.getByRole('button', { name: /enviar/i });
    await user.click(sendButton);

    expect(sendMessage).toHaveBeenCalledWith('Hello AI');
  });

  it('displays messages in chat panel', async () => {
    mockUseAIChat.mockReturnValue({
      ...defaultMockReturn,
      messages: [
        {
          id: '1',
          role: 'user',
          content: 'Hello',
          created_at: new Date().toISOString(),
        },
        {
          id: '2',
          role: 'assistant',
          content: 'Hi there!',
          created_at: new Date().toISOString(),
        },
      ],
    });

    const user = userEvent.setup();
    render(<ChatWidget />);

    // Open panel
    const fabButton = screen.getByRole('button', { name: /abrir assistente/i });
    await user.click(fabButton);

    expect(screen.getByText('Hello')).toBeInTheDocument();
    expect(screen.getByText('Hi there!')).toBeInTheDocument();
  });

  it('shows thinking indicator when loading with messages', async () => {
    // ThinkingIndicator only shows when there are messages (it's inside the messages view)
    mockUseAIChat.mockReturnValue({
      ...defaultMockReturn,
      messages: [
        {
          id: '1',
          role: 'user',
          content: 'What are the pending occurrences?',
          created_at: new Date().toISOString(),
        },
      ],
      isLoading: true,
      thinkingStep: 'Consultando Base de Conhecimento...',
    });

    const user = userEvent.setup();
    render(<ChatWidget />);

    // Open panel
    const fabButton = screen.getByRole('button', { name: /abrir assistente/i });
    await user.click(fabButton);

    // Wait for the thinking indicator to appear
    await waitFor(() => {
      expect(screen.getByText('Consultando Base de Conhecimento...')).toBeInTheDocument();
    });
  });

  it('closes panel on escape key', async () => {
    const user = userEvent.setup();
    render(<ChatWidget />);

    // Open panel
    const fabButton = screen.getByRole('button', { name: /abrir assistente/i });
    await user.click(fabButton);

    expect(screen.getByText('Assistente SIDOT')).toBeInTheDocument();

    // Press escape
    await user.keyboard('{Escape}');

    // Panel should be closed
    await waitFor(() => {
      const panel = screen.getByRole('dialog', { hidden: true });
      expect(panel).toHaveClass('translate-x-full');
    });
  });
});
