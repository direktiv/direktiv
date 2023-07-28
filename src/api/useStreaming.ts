import { useEffect, useRef } from "react";

export const useStreaming = ({
  url,
  onOpen,
  onMessage,
  onError,
  enabled,
}: {
  url: string;
  onOpen?: (e: Event) => void;
  onMessage?: (e: MessageEvent) => void;
  onError?: (e: Event) => void;
  enabled?: boolean;
}) => {
  const eventSource = useRef<EventSource | null>(null);

  const stopStreaming = () => {
    eventSource.current?.close();
    eventSource.current = null;
  };

  const startSteaming = () => {
    if (enabled && eventSource.current === null) {
      const listener = new EventSource(url);
      eventSource.current = listener;
      if (onOpen) listener.onopen = onOpen;
      if (onError) listener.onerror = onError;
      if (onMessage) listener.onmessage = onMessage;
    }
  };

  useEffect(() => {
    startSteaming();
    return () => {
      stopStreaming();
    };
  });
};

export default useStreaming;
