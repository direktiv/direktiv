import { useEffect, useRef, useState } from "react";

import { LogEntry } from "~/design/Logs";
import LogList from "~/components/Logs";
import { useVirtualizer } from "@tanstack/react-virtual";

const ScrollContainer = ({ logs }: { logs: string[] }) => {
  const [scrolledToBottom, setScrolledToBottom] = useState(true);

  // The scrollable element for the list
  const parentRef = useRef<HTMLDivElement | null>(null);

  // The virtualizer
  const rowVirtualizer = useVirtualizer({
    count: logs.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 20,
    /**
     * overscan is the number of items to render above and below
     * the visible window. More items = less flickering when
     * scrolling, but more memory usage and initial load time.
     * I tested it with around 2000 items and 40 was a good fit
     * to have no flickering with pretty high scrolling speed.
     */
    overscan: 40,
  });

  useEffect(() => {
    if (logs.length && scrolledToBottom) {
      rowVirtualizer.scrollToIndex(logs.length - 1), { align: "end" };
    }
  }, [logs.length, rowVirtualizer, scrolledToBottom]);

  const virtualItems = rowVirtualizer.getVirtualItems();

  return (
    <LogList
      ref={parentRef}
      height={rowVirtualizer.getTotalSize()}
      virtualOffset={virtualItems[0]?.start ?? 0}
      scrolledToBottom={scrolledToBottom}
      setScrolledToBottom={setScrolledToBottom}
      scrollButtonClassName="aria-[hidden=true]:bottom-6, aria-[hidden=false]:bottom-11"
      onScroll={(e) => {
        const element = e.target as HTMLDivElement;
        if (element) {
          const { scrollHeight, scrollTop, clientHeight } = element;
          const scrollDistanceToBottom =
            scrollHeight - scrollTop - clientHeight;
          setScrolledToBottom(scrollDistanceToBottom < 100);
        }
      }}
    >
      {virtualItems.map((virtualItem) => {
        const logEntry = logs[virtualItem.index];
        if (!logEntry) return null;
        return (
          <LogEntry
            key={virtualItem.key}
            data-index={virtualItem.key}
            ref={rowVirtualizer.measureElement}
          >
            {logEntry}
          </LogEntry>
        );
      })}
    </LogList>
  );
};

export default ScrollContainer;
