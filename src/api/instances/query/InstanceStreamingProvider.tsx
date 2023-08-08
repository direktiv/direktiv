import { FC, PropsWithChildren } from "react";

import { useInstanceDetailsStream } from "./details";

type InstanceStreamingProviderType = PropsWithChildren & {
  instanceId: string;
  enabled?: boolean;
};

export const InstanceStreamingProvider: FC<InstanceStreamingProviderType> = ({
  instanceId,
  enabled,
  children,
}) => {
  useInstanceDetailsStream({ instanceId }, { enabled: enabled ?? true });
  return <>{children}</>;
};
