import { UseEventsStreamParams, useEventsStream } from "./stream";

import { memo } from "react";
import { useEvents } from "./get";

const MemoizedEventStream = memo((params: UseEventsStreamParams) => {
  useEventsStream(params);
  return null;
});

MemoizedEventStream.displayName = "MemoizedEventStream";

export const EventStreamingSubscriber = (params: UseEventsStreamParams) => {
  const { isFetching } = useEvents(params);

  /**
   * when events are fetched (via non-streaming api), the subscription
   * must be paused/reinitialized to prevent losing some logs. Because
   * when the request finishes, the cache will be overwritten and all
   * logs that streamed in since the request was kicked off will be
   * lost.
   *
   * Reinitializing the subscription will fix that issue, because a new
   * subscription will retrieve all realtime logs plus a couple of
   * seconds of old logs to prevent race conditions.
   */
  let enableStreaming = !isFetching;

  // if explicitely disabled by params
  if (params.enabled === false) {
    enableStreaming = false;
  }

  return <MemoizedEventStream {...params} enabled={enableStreaming} />;
};

EventStreamingSubscriber.displayName = "EventStreamingSubscriber";
