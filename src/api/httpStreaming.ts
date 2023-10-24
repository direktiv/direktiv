import { useCallback, useEffect } from "react";

type HttpStreamingOptions = {
  url: string;
  apiKey?: string;
  onMessage?: (message: string) => void;
  onError?: (e: unknown) => void;
  enabled?: boolean;
};

const decoder = new TextDecoder("iso-8859-2");

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
              headers: {
                "direktiv-token": apiKey,
              },
            }
          : {}),
        // this only throws if the request is aborted before the first response is received
      }).catch(() => null);

      if (!response || !response.ok || !response.body) {
        return;
      }

      const reader = response.body.getReader();

      let finished = false;

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
          onMessage?.(chunk);
        } catch (error) {
          onError?.(error);
          finished = true;
        }
      }
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
