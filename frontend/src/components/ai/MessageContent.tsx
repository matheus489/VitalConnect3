'use client';

import { useMemo } from 'react';
import { OccurrenceCard } from './OccurrenceCard';
import { OccurrenceTable } from './OccurrenceTable';
import { ActionButtonGroup } from './ActionButton';
import type {
  ParsedMessageContent,
  MessageContentSegment,
  OccurrenceCardData,
  OccurrenceTableData,
  ActionButtonData,
  AIConfirmationRequest,
} from '@/types/ai';

// Component markers that can appear in AI responses
const COMPONENT_MARKERS = {
  OCCURRENCE_CARD_START: '{{occurrence_card:',
  OCCURRENCE_TABLE_START: '{{occurrence_table:',
  ACTION_BUTTONS_START: '{{action_buttons:',
  CONFIRMATION_START: '{{confirmation:',
  COMPONENT_END: '}}',
};

interface MessageContentProps {
  content: string;
  onOccurrenceClick?: (occurrence: OccurrenceCardData) => void;
  onActionClick?: (action: string, params?: Record<string, unknown>) => void;
  onConfirmationRequired?: (request: AIConfirmationRequest) => void;
  loadingActionId?: string;
  className?: string;
}

/**
 * Parses the AI message content and extracts component markers
 * Returns an array of segments that can be text or structured components
 */
function parseMessageContent(content: string): ParsedMessageContent {
  const segments: MessageContentSegment[] = [];
  let currentIndex = 0;

  while (currentIndex < content.length) {
    // Find the next component marker
    let nextMarkerIndex = -1;
    let markerType: string | null = null;
    let markerStart = '';

    for (const [type, marker] of Object.entries(COMPONENT_MARKERS)) {
      if (type === 'COMPONENT_END') continue;

      const index = content.indexOf(marker, currentIndex);
      if (index !== -1 && (nextMarkerIndex === -1 || index < nextMarkerIndex)) {
        nextMarkerIndex = index;
        markerType = type;
        markerStart = marker;
      }
    }

    if (nextMarkerIndex === -1) {
      // No more markers, add remaining text
      const remainingText = content.slice(currentIndex).trim();
      if (remainingText) {
        segments.push({ type: 'text', content: remainingText });
      }
      break;
    }

    // Add text before the marker
    const textBefore = content.slice(currentIndex, nextMarkerIndex).trim();
    if (textBefore) {
      segments.push({ type: 'text', content: textBefore });
    }

    // Find the end of the component
    const markerContentStart = nextMarkerIndex + markerStart.length;
    const markerEnd = content.indexOf(COMPONENT_MARKERS.COMPONENT_END, markerContentStart);

    if (markerEnd === -1) {
      // Malformed marker, treat rest as text
      const remainingText = content.slice(nextMarkerIndex).trim();
      if (remainingText) {
        segments.push({ type: 'text', content: remainingText });
      }
      break;
    }

    // Extract and parse the JSON data
    const jsonStr = content.slice(markerContentStart, markerEnd);
    try {
      const data = JSON.parse(jsonStr);

      switch (markerType) {
        case 'OCCURRENCE_CARD_START':
          segments.push({
            type: 'occurrence_card',
            data: data as OccurrenceCardData,
          });
          break;
        case 'OCCURRENCE_TABLE_START':
          segments.push({
            type: 'occurrence_table',
            data: data as OccurrenceTableData,
          });
          break;
        case 'ACTION_BUTTONS_START':
          segments.push({
            type: 'action_buttons',
            data: data as ActionButtonData[],
          });
          break;
        case 'CONFIRMATION_START':
          segments.push({
            type: 'confirmation_dialog',
            data: data as AIConfirmationRequest,
          });
          break;
      }
    } catch {
      // Invalid JSON, treat as text
      segments.push({
        type: 'text',
        content: content.slice(nextMarkerIndex, markerEnd + COMPONENT_MARKERS.COMPONENT_END.length),
      });
    }

    currentIndex = markerEnd + COMPONENT_MARKERS.COMPONENT_END.length;
  }

  return { segments };
}

/**
 * Renders text content with basic markdown-like formatting
 */
function TextContent({ content }: { content: string }) {
  // Split by newlines and render paragraphs
  const paragraphs = content.split('\n\n').filter(Boolean);

  return (
    <>
      {paragraphs.map((paragraph, idx) => (
        <p key={idx} className="whitespace-pre-wrap">
          {paragraph.split('\n').map((line, lineIdx) => (
            <span key={lineIdx}>
              {lineIdx > 0 && <br />}
              {line}
            </span>
          ))}
        </p>
      ))}
    </>
  );
}

export function MessageContent({
  content,
  onOccurrenceClick,
  onActionClick,
  onConfirmationRequired,
  loadingActionId,
  className,
}: MessageContentProps) {
  const parsed = useMemo(() => parseMessageContent(content), [content]);

  const renderSegment = (segment: MessageContentSegment, index: number) => {
    switch (segment.type) {
      case 'text':
        return <TextContent key={index} content={segment.content} />;

      case 'occurrence_card':
        return (
          <OccurrenceCard
            key={index}
            data={segment.data}
            onClick={onOccurrenceClick}
            className="my-2"
          />
        );

      case 'occurrence_table':
        return (
          <OccurrenceTable
            key={index}
            data={segment.data}
            onRowClick={onOccurrenceClick}
            className="my-2"
          />
        );

      case 'action_buttons':
        return (
          <ActionButtonGroup
            key={index}
            buttons={segment.data}
            onClick={onActionClick}
            loadingActionId={loadingActionId}
          />
        );

      case 'confirmation_dialog':
        // Notify parent about confirmation requirement
        // The actual dialog rendering is handled by the parent component
        if (onConfirmationRequired) {
          // Use setTimeout to avoid state update during render
          setTimeout(() => onConfirmationRequired(segment.data), 0);
        }
        return null;

      default:
        return null;
    }
  };

  return (
    <div data-slot="message-content" className={className}>
      {parsed.segments.map(renderSegment)}
    </div>
  );
}

// Export the parser for testing purposes
export { parseMessageContent };

export default MessageContent;
