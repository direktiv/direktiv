import { useCallback, useEffect } from "react";

import { getAuthHeader } from "./utils";
import { z } from "zod";

type HttpStreamingOptions = {
  url: string;
  apiKey?: string;
  onMessage?: (message: unknown, isFirstMessage: boolean) => void;
  onError?: (e: unknown) => void;
  enabled?: boolean;
};

const decoder = new TextDecoder("iso-8859-2");

const processStream = async ({
  reader,
  onMessage,
  onError,
}: {
  reader: ReadableStreamDefaultReader;
  onMessage: HttpStreamingOptions["onMessage"];
  onError: HttpStreamingOptions["onError"];
}) => {
  let finished = false;
  let isFirstMessage = true;
  while (!finished) {
    const { done, value } = await reader.read();
    if (done) {
      finished = true;
      break;
    }
    try {
      const chunk = decoder.decode(value, {
        stream: true,
      });
      onMessage?.(chunk, isFirstMessage);
      isFirstMessage = false;
    } catch (error) {
      onError?.(error);
      finished = true;
    }
  }
};

/**
 * a react hook that handles a connection to an http endpoint and streams
 * the response. All messages are forwarded to the onMessage callback.
 */
export const useHttpStreaming = ({
  url,
  apiKey,
  onMessage,
  onError,
  enabled = true,
}: HttpStreamingOptions) => {
  const startStreaming = useCallback(
    async (abortController: AbortController) => {
      const response = await fetch(url, {
        signal: abortController.signal,
        ...(apiKey
          ? {
              headers: { ...getAuthHeader(apiKey) },
            }
          : {}),
        /**
         * this throws an error if the request is aborted before the first
         * response is received. We don't want to forward this error to
         * the user
         */
      }).catch(() => null);

      if (!response || !response.ok || !response.body) {
        return;
      }

      processStream({ reader: response.body.getReader(), onMessage, onError });
    },
    [apiKey, onError, onMessage, url]
  );

  useEffect(() => {
    const abortController = new AbortController();
    if (enabled) {
      startStreaming(abortController).catch(() => null);
    }
    return () => {
      abortController.abort();
    };
  }, [enabled, onError, startStreaming]);
};

/**
 * react hook that acts as a proxy for useHttpStreaming
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
  onMessage: (message: T, isFirstMessage: boolean) => void;
}) =>
  useHttpStreaming({
    url,
    apiKey,
    enabled,
    onMessage: (message, isFirstMessage) => {
      const parsedResult = schema.safeParse(message);
      if (parsedResult.success === false) {
        console.error(
          `error parsing streaming result for ${url}`,
          parsedResult.error
        );
        return;
      }
      onMessage(parsedResult.data, isFirstMessage);
    },
  });
