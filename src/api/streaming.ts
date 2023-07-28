import { useEffect, useRef } from "react";

import { z } from "zod";

export const useEventSource = ({
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
      // when streaming is enabled and there is no event source yet, create one
      const listener = new EventSource(url);
      eventSource.current = listener;
      // connect all the callbacks
      if (onOpen) listener.onopen = onOpen;
      if (onError) listener.onerror = onError;
      if (onMessage) listener.onmessage = onMessage;
    }
  };

  useEffect(() => {
    startSteaming();
    return () => {
      // close connection on unmount to prevent memory leaks
      stopStreaming();
    };
  });
};

export const useStreaming = <T>({
  url,
  enabled,
  schema,
  onMessage,
}: {
  url: string;
  enabled: boolean;
  schema: z.ZodSchema<T>;
  onMessage: (msg: T) => void;
}) =>
  useEventSource({
    url,
    enabled,
    onMessage: (msg) => {
      if (!msg.data) return null;
      let msgJson = null;
      try {
        // try to parse the response as json
        msgJson = JSON.parse(msg.data);
      } catch (e) {
        console.error(
          `error parsing streaming result (${msg.data}) from ${url}} as JSON`
        );
        return;
      }

      const parsedResult = schema.safeParse(msgJson);
      if (parsedResult.success === false) {
        console.error(`error parsing streaming result for ${url}`);
        return;
      }

      onMessage(parsedResult.data);
    },
  });
