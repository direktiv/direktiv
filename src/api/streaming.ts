import {
  Event,
  EventSourcePolyfill,
  MessageEvent,
} from "event-source-polyfill";
import { useEffect, useRef } from "react";

import { z } from "zod";

/**
 * a react hook that opens a connection to an event source and listens
 * to the onopen, onerror and onmessage callback. It provides a flag to
 * enable/disable the streaming
 */
export const useEventSource = ({
  url,
  apiKey,
  onOpen,
  onMessage,
  onError,
  enabled,
}: {
  url: string;
  apiKey?: string;
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
      // polyfill is required to support sending additional headers
      const listener = new EventSourcePolyfill(url, {
        ...(apiKey
          ? {
              headers: {
                "direktiv-token": apiKey,
              },
            }
          : {}),
      });
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

/**
 * react hook that acts as a proxy for useEventSource
 * and implements schema validation on top of it
 */
export const useStreaming = <T>({
  url,
  apiKey,
  enabled,
  schema,
  onMessage,
}: {
  url: string;
  apiKey?: string;
  enabled: boolean;
  schema: z.ZodSchema<T>;
  onMessage: (msg: T) => void;
}) =>
  useEventSource({
    url,
    apiKey,
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
