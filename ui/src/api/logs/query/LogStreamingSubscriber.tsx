import {
  UseLogsParams,
  UseLogsStreamParams,
  useLogs,
  useLogsStream,
} from "./logs";

import { memo } from "react";

const MemoizedLogsStream = memo((params: UseLogsStreamParams) => {
  useLogsStream(params);
  return null;
});

MemoizedLogsStream.displayName = "MemoizedLogsStream";

export const LogStreamingSubscriber = (params: UseLogsParams) => {
  const { isFetching } = useLogs(params);

  /**
   * when the logs are fetched (via non-streaming api), the subscription
   * must be paused/reinitialized to prevent losing some logs. Because
   * when the request finishes, the cache will be overwritten and all
   * logs that streamed in since the request was kicked off will be
   * lost.
   *
   * Reinitializing the subscription will fix that issue, because a new
   * subscription will retrieve all realtime logs plus a couple of
   * seconds of old logs to prevent race conditions.
   */
  const enableStreaming = !!params.enabled || !isFetching;
  return <MemoizedLogsStream {...params} enabled={enableStreaming} />;
};

LogStreamingSubscriber.displayName = "LogStreamingSubscriber";
