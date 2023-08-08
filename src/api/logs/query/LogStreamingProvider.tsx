import { FC, PropsWithChildren } from "react";
import { FiltersObj, useLogsStream } from "./get";

type LogStreamingProviderType = PropsWithChildren & {
  instanceId: string;
  filters?: FiltersObj;
  enabled?: boolean;
};

export const LogStreamingProvider: FC<LogStreamingProviderType> = ({
  instanceId,
  filters,
  enabled,
  children,
}) => {
  useLogsStream({ instanceId, filters }, { enabled: enabled ?? true });
  return <>{children}</>;
};
