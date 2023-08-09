import { memo } from "react";
import { useNamespaceLogsStream } from "./logs";

type LogStreamingProviderTypeProps = { enabled?: boolean };

export const NamespaceLogsStreamingProvider = memo(
  ({ enabled }: LogStreamingProviderTypeProps) => {
    useNamespaceLogsStream({ enabled: enabled ?? true });
    return null;
  }
);

NamespaceLogsStreamingProvider.displayName = "NamespaceLogsStreamingProvider";
