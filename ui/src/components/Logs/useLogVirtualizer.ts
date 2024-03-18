import { useCallback, useEffect, useRef, useState } from "react";

import { useLogs } from "~/api/logs/query/logs";
import { useVirtualizer } from "@tanstack/react-virtual";

const defaultLogHeight = 20;

type useLogVirtualizerParams = {
  queryLogsBy?: Parameters<typeof useLogs>[0];
};

/**
 * this hook is used to render a virtualized list of log entries. It utilizes
 * tanstack/react-virtual and adds data fetching, pagination and some custom
 * scrolling logic on top of that.
 */
export const useLogVirtualizer = ({
  queryLogsBy,
}: useLogVirtualizerParams = {}) => {
  const parentRef = useRef<HTMLDivElement | null>(null);

  const [scrolledToBottom, setScrolledToBottom] = useState(true);

  const prevNumberOfLogs = useRef<number | null>(null);
  const prevOldestLogId = useRef<number | null>(null);

  const {
    data: logs = [],
    hasPreviousPage,
    fetchPreviousPage,
    isFetchingPreviousPage,
  } = useLogs(queryLogsBy);

  const numberOfLogs = logs.length;

  const rowVirtualizer = useVirtualizer({
    count: numberOfLogs,
    getScrollElement: () => parentRef.current,
    estimateSize: () => defaultLogHeight,
    getItemKey: useCallback(
      (index: number) => {
        const uniqueId = logs[index]?.id;
        if (!uniqueId)
          throw new Error("Could not find a log id for the virtualizer.");
        return uniqueId;
      },
      [logs]
    ),
    /**
     * Start at the bottom, this is especially important to avoid
     * triggering fetchPreviousPage right away when the page loads.
     * It also avoids flickering, because the useEffect will initiate
     * a bottom scroll anyway.
     */
    initialOffset: 999999,
    /**
     * overscan is the number of items to render above and below
     * the visible window. More items = less flickering when
     * scrolling, but more memory usage and initial load time.
     * I tested it with around 2000 items and 40 was a good fit
     * to have no flickering with pretty high scrolling speed.
     */
    overscan: 40,
    onChange(instance) {
      prevNumberOfLogs.current = numberOfLogs;
      prevOldestLogId.current = logs?.[0]?.id ?? null;
      /**
       * when the last x loglines are visible in the list (x being
       * loglinesThreashold), the user is considered to be at the
       * bottom of the list.
       */
      if (!instance.range) return;
      const { endIndex: lastVisibleIndex } = instance.range;
      const loglinesThreashold = 5;
      setScrolledToBottom(
        lastVisibleIndex >= numberOfLogs - 1 - loglinesThreashold
      );
    },
  });

  const virtualItems = rowVirtualizer.getVirtualItems();

  /**
   * fetch the previous log page when a users scrolls to the top
   */
  useEffect(() => {
    const [firstLogEntry] = virtualItems;
    if (
      firstLogEntry?.index === 0 &&
      hasPreviousPage &&
      !isFetchingPreviousPage &&
      numberOfLogs === prevNumberOfLogs?.current
    ) {
      fetchPreviousPage();
    }
  }, [
    fetchPreviousPage,
    hasPreviousPage,
    isFetchingPreviousPage,
    numberOfLogs,
    virtualItems,
  ]);

  /**
   * when the user reached the bottom of the list we need to keep the scoll
   * position to the very bottom to make sure new log entries are in the
   * viewport.
   */
  useEffect(() => {
    if (numberOfLogs > 0 && scrolledToBottom) {
      rowVirtualizer.scrollToIndex(numberOfLogs - 1, { align: "end" });
    }
  }, [numberOfLogs, rowVirtualizer, scrolledToBottom]);

  /**
   * maintain the scroll position when a set of new logs is added to the top
   * of the list
   */
  useEffect(() => {
    if (
      !scrolledToBottom &&
      prevNumberOfLogs.current &&
      prevNumberOfLogs.current !== numberOfLogs &&
      // this will make sure the new received logs were added at the top
      prevOldestLogId.current !== (logs?.[0]?.id ?? null)
    ) {
      /**
       * To maintin the old scroll position we need to know how many logs
       * were added to the top of the list.
       *
       * Example:
       * The user was at a scrollOffset of 100 and then 200 new logs have
       * been added to the top. We now need to add 200 times the height of
       * a log entry (200 * defaultLogHeight) to that offset to stay at the
       * same scroll position.
       */
      const { scrollOffset: currentOffset } = rowVirtualizer;
      const numberOfNewLogs = numberOfLogs - prevNumberOfLogs.current;
      const newOffset = currentOffset + numberOfNewLogs * defaultLogHeight;
      rowVirtualizer.scrollToOffset(newOffset);
    }
  }, [logs, numberOfLogs, rowVirtualizer, scrolledToBottom]);

  return {
    rowVirtualizer,
    parentRef,
    logs,
    scrolledToBottom,
    setScrolledToBottom,
  };
};
