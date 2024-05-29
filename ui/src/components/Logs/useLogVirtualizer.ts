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

  const prevNumberOfLogLines = useRef<number | null>(null);
  const prevOldestLogLineId = useRef<number | null>(null);

  const {
    data: logLines = [],
    hasNextPage,
    fetchNextPage,
    isFetchingNextPage,
  } = useLogs(queryLogsBy);

  const numberOfLogLines = logLines.length;

  const rowVirtualizer = useVirtualizer({
    count: numberOfLogLines,
    getScrollElement: () => parentRef.current,
    estimateSize: () => defaultLogHeight,
    getItemKey: useCallback(
      (index: number) => {
        const uniqueId = logLines[index]?.id;
        if (!uniqueId)
          throw new Error("Could not find a log line id for the virtualizer.");
        return uniqueId;
      },
      [logLines]
    ),
    /**
     * overscan is the number of items to render above and below
     * the visible window. More items = less flickering when
     * scrolling, but more memory usage and initial load time.
     * I tested it with around 2000 items and 40 was a good fit
     * to have no flickering with pretty high scrolling speed.
     */
    overscan: 40,
    onChange(instance) {
      prevNumberOfLogLines.current = numberOfLogLines;
      prevOldestLogLineId.current = logLines?.[0]?.id ?? null;
      /**
       * when the last x log lines are visible in the list (x being
       * logLinesThreshold), the user is considered to be at the
       * bottom of the list.
       */
      if (!instance.range) return;
      const { endIndex: lastVisibleIndex } = instance.range;
      const logLinesThreshold = 5;
      setScrolledToBottom(
        lastVisibleIndex >= numberOfLogLines - 1 - logLinesThreshold
      );
    },
  });

  const virtualItems = rowVirtualizer.getVirtualItems();

  /**
   * fetch the previous log page when a users scrolls to the top
   */
  useEffect(() => {
    const [firstLogLine] = virtualItems;
    if (
      firstLogLine?.index === 0 &&
      hasNextPage &&
      !isFetchingNextPage &&
      prevNumberOfLogLines?.current === numberOfLogLines
    ) {
      fetchNextPage();
    }
  }, [
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    numberOfLogLines,
    virtualItems,
  ]);

  /**
   * when the user reached the bottom of the list we need to keep the scoll
   * position to the very bottom to make sure new log entries are in the
   * viewport.
   */
  useEffect(() => {
    if (numberOfLogLines > 0 && scrolledToBottom) {
      rowVirtualizer.scrollToIndex(numberOfLogLines - 1, { align: "end" });
    }
  }, [numberOfLogLines, rowVirtualizer, scrolledToBottom]);

  /**
   * maintain the scroll position when a set of new log lines is added to the top
   * of the list
   */
  useEffect(() => {
    if (
      !scrolledToBottom &&
      prevNumberOfLogLines.current &&
      prevNumberOfLogLines.current !== numberOfLogLines &&
      // this will make sure the new received log lines were added at the top
      prevOldestLogLineId.current !== (logLines?.[0]?.id ?? null)
    ) {
      /**
       * To maintain the old scroll position we need to know how many log
       * lines were added to the top of the list.
       *
       * Example:
       * The user was at a scrollOffset of 100 and then 200 new lines have
       * been added to the top. We now need to add 200 times the height of
       * a log entry (200 * defaultLogHeight) to that offset to stay at the
       * same scroll position.
       */
      const { scrollOffset: currentOffset } = rowVirtualizer;
      const numberOfNewLogLines =
        numberOfLogLines - prevNumberOfLogLines.current;
      const newOffset = currentOffset + numberOfNewLogLines * defaultLogHeight;
      prevNumberOfLogLines.current = numberOfLogLines;
      rowVirtualizer.scrollToOffset(newOffset);
    }
  }, [logLines, numberOfLogLines, rowVirtualizer, scrolledToBottom]);

  return {
    rowVirtualizer,
    parentRef,
    logLines,
    scrolledToBottom,
    setScrolledToBottom,
  };
};
