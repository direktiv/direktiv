import { FC, PropsWithChildren } from "react";

import { useNamespaceLogsStream } from "./logs";

type NamespaceLogsStreamingProviderType = PropsWithChildren & {
  enabled?: boolean;
};

export const NamespaceLogsStreamingProvider: FC<
  NamespaceLogsStreamingProviderType
> = ({ enabled, children }) => {
  useNamespaceLogsStream({ enabled: enabled ?? true });
  return <>{children}</>;
};
