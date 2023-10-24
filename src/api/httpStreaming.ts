import { useCallback, useEffect } from "react";

type HttpStreamingOptions = {
  url: string;
  apiKey?: string;
  onMessage?: (message: string) => void;
  onError?: (e: unknown) => void;
  enabled?: boolean;
};

export const useHttpStreaming = ({
  url,
  apiKey,
  onMessage,
  onError,
  enabled = true,
}: HttpStreamingOptions) => {
  const startSteaming = useCallback(
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
        // this only throws if the request is aborted, don't treat it as an error
      }).catch(() => null);

      if (!response || !response.ok || !response.body) {
        return;
      }

      const reader = response.body.getReader();

      return new ReadableStream({
        async start() {
          let finished = false;
          while (!finished) {
            const { done, value } = await reader.read();
            if (done) {
              finished = true;
              break;
            }

            try {
              const chunk = new TextDecoder().decode(value);
              onMessage?.(chunk);
            } catch (error) {
              onError?.(error);
              finished = true;
            }
          }
        },
      });
    },
    [apiKey, onError, onMessage, url]
  );

  useEffect(() => {
    const abortController = new AbortController();
    if (enabled) {
      startSteaming(abortController).catch((e) => {
        onError?.(e);
      });
    }
    return () => {
      abortController.abort();
    };
  }, [enabled, onError, startSteaming]);
};
