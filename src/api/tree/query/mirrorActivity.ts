import {
  MirrorActivityLogSchema,
  MirrorActivityLogSchemaType,
} from "../schema/mirror";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import { memo } from "react";
import moment from "moment";
import { treeKeys } from "..";
import { useApiKey } from "~/util/store/apiKey";
import { useNamespace } from "~/util/store/namespace";
import { useStreaming } from "~/api/streaming";

const updateCache = (
  oldData: MirrorActivityLogSchemaType | undefined,
  msg: MirrorActivityLogSchemaType
) => {
  if (!oldData) {
    return msg;
  }
  /**
   * Dedup logs. The onMessage callback gets called in two different cases:
   *
   * case 1:
   * when the SSE connection is established, the whole set of logs is received
   *
   * case 2:
   * after the connection is established and only some new log entries are received
   *
   * it's also important to note that multiple components can subscribe to the same
   * cache, so we can have case 1 and 2 at the same time, or case 1 after case 2
   */
  const lastCachedLog = oldData.results[oldData.results.length - 1];
  let newResults: typeof oldData.results = [];

  // there was a previous cache, but with no entries yet
  if (!lastCachedLog) {
    newResults = msg.results;
    // there was a previous cache with entries
  } else {
    const newestLogTimeFromCache = moment(lastCachedLog.t);
    // new results are all logs that are newer than the last cached log
    newResults = msg.results.filter((entry) =>
      newestLogTimeFromCache.isBefore(entry.t)
    );
  }

  return {
    ...oldData,
    results: [...oldData.results, ...newResults],
  };
};

export const useMirrorActivityLogStream = (
  { activityId }: { activityId: string },
  { enabled = true }: { enabled?: boolean } = {}
) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const queryClient = useQueryClient();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useStreaming({
    url: `/api/namespaces/${namespace}/activities/${activityId}/logs`,
    apiKey: apiKey ?? undefined,
    enabled,
    schema: MirrorActivityLogSchema,
    onMessage: (msg) => {
      queryClient.setQueryData<MirrorActivityLogSchemaType>(
        treeKeys.activityLog(namespace, {
          apiKey: apiKey ?? undefined,
          activityId,
        }),
        (oldData) => updateCache(oldData, msg)
      );
    },
  });
};

type MirrorActivityLogSubscriberType = {
  activityId: string;
  enabled?: boolean;
};

export const MirrorActivityLogSubscriber = memo(
  ({ activityId, enabled }: MirrorActivityLogSubscriberType) => {
    useMirrorActivityLogStream({ activityId }, { enabled: enabled ?? true });
    return null;
  }
);

MirrorActivityLogSubscriber.displayName = "MirrorActivityLogSubscriber";

export const useMirrorActivityLog = ({
  activityId,
}: {
  activityId: string;
}) => {
  const apiKey = useApiKey();
  const namespace = useNamespace();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useQuery<MirrorActivityLogSchemaType>({
    queryKey: treeKeys.activityLog(namespace, {
      apiKey: apiKey ?? undefined,
      activityId,
    }),
    /**
     * This hook is only used to subscribe to the correct cache key. Data for this key
     * will be added by a streaming subscriber. We don't have a non-streaming endpoint
     * for initial data. So the queryFn is missing on purpose and the enabled flag is set
     * to false.
     */
    enabled: false,
  });
};
