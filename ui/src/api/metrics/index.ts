import { forceLeadingSlash } from "../files/utils";

export const metricsKeys = {
  metrics: (
    namespace: string,
    { apiKey, path }: { apiKey?: string; path?: string }
  ) =>
    [
      {
        scope: "metrics",
        apiKey,
        namespace,
        path: forceLeadingSlash(path),
      },
    ] as const,
};
