import { useCallback, useEffect, useRef, useState } from "react";

import { LogEntryType } from "~/api/logs/schema";
import { useVirtualizer } from "@tanstack/react-virtual";

const defaultLogHeight = 20;

type UseLogVirtualizerParams = {
  logs: LogEntryType[];
  fetchPreviousPage: () => void;
  hasPreviousPage: boolean | undefined;
  isFetchingPreviousPage: boolean;
};

export const useLogVirtualizer = ({
  logs,
  fetchPreviousPage,
  hasPreviousPage,
  isFetchingPreviousPage,
}: UseLogVirtualizerParams) => {
  const lastScrollPos = useRef<{
    scrollOffset: number;
    numberOfLogs: number;
  } | null>(null);
  // The scrollable element for the list
  const parentRef = useRef<HTMLDivElement | null>(null);
  const numberOfLogs = logs.length;
  const [watch, setWatch] = useState(true);
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
      if (!instance.range) return;
      const { scrollOffset } = instance;
      lastScrollPos.current = { scrollOffset, numberOfLogs };
    },
  });

  const virtualItems = rowVirtualizer.getVirtualItems();

  useEffect(() => {
    if (numberOfLogs > 0 && watch) {
      rowVirtualizer.scrollToIndex(numberOfLogs, { align: "end" });
    }
  }, [numberOfLogs, rowVirtualizer, watch]);

  useEffect(() => {
    /**
     * the last scroll position is cached in a ref in form of a the startIndex
     * and the number of logs. When the number of logs changes, we need to translate
     * the last scroll position to the new index of the same log entry.
     */
    if (
      !watch &&
      lastScrollPos.current &&
      lastScrollPos.current.numberOfLogs !== numberOfLogs
    ) {
      /**
       * we can utilize the diff to detect if the update was added via a pagination of
       * a new log entry that has been streamed
       */
      const diff = numberOfLogs - lastScrollPos.current.numberOfLogs;
      const newOffset = rowVirtualizer.scrollOffset + diff * defaultLogHeight;
      rowVirtualizer.scrollToOffset(newOffset);
    }
  }, [numberOfLogs, rowVirtualizer, watch]);

  useEffect(() => {
    const [firstLogEntry] = virtualItems;
    if (
      firstLogEntry?.index === 0 &&
      hasPreviousPage &&
      !isFetchingPreviousPage &&
      numberOfLogs === lastScrollPos?.current?.numberOfLogs
    ) {
      fetchPreviousPage();
    }
  }, [
    virtualItems,
    fetchPreviousPage,
    hasPreviousPage,
    isFetchingPreviousPage,
    numberOfLogs,
  ]);

  return { rowVirtualizer, parentRef, setWatch, watch };
};
