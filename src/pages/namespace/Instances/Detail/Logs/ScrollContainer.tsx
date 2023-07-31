import { FC, useEffect, useRef, useState } from "react";
import { FiltersObj, useLogs } from "~/api/logs/query/get";

import { ArrowDown } from "lucide-react";
import Button from "~/design/Button";
import Entry from "./Entry";
import { Logs } from "~/design/Logs";
import { twMergeClsx } from "~/util/helpers";
import { useVirtualizer } from "@tanstack/react-virtual";

const ScrollContainer: FC<{
  instanceId: string;
  query: FiltersObj;
}> = ({ instanceId, query }) => {
  const { data } = useLogs({
    instanceId,
    filters: query,
  });
  const [watch, setWatch] = useState(true);

  // The scrollable element for the list
  const parentRef = useRef<HTMLDivElement | null>(null);

  // The virtualizer
  const rowVirtualizer = useVirtualizer({
    count: data?.results.length ?? 0,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 20,
    getItemKey: (index) => data?.results[index]?.t ?? index,
  });

  useEffect(() => {
    if (data?.results.length && watch) {
      rowVirtualizer.scrollToIndex(data?.results.length);
    }
  }, [data?.results.length, rowVirtualizer, watch]);

  if (!data) return null;

  return (
    <Logs
      linewrap={true}
      className="h-full overflow-scroll"
      ref={parentRef}
      onScroll={(e) => {
        const element = e.target as HTMLDivElement;
        if (element) {
          const { scrollHeight, scrollTop, clientHeight } = element;
          const scrollDistanceToBottom =
            scrollHeight - scrollTop - clientHeight;
          setWatch(scrollDistanceToBottom < 100);
        }
      }}
    >
      <div
        className="relative w-full"
        style={{
          height: `${rowVirtualizer.getTotalSize()}px`,
        }}
      >
        {rowVirtualizer.getVirtualItems().map((virtualItem) => {
          const logEntry = data.results[virtualItem.index];
          if (!logEntry) return null;
          return (
            <Entry
              key={virtualItem.key}
              logEntry={logEntry}
              className="absolute top-0 left-0 w-full"
              style={{
                height: `${virtualItem.size}px`,
                transform: `translateY(${virtualItem.start}px)`,
              }}
            />
          );
        })}
      </div>
      <div
        className={twMergeClsx(
          "absolute box-border flex w-full pr-10",
          "justify-center transition-all",
          "aria-[hidden=true]:pointer-events-none aria-[hidden=true]:bottom-5 aria-[hidden=true]:opacity-0",
          "aria-[hidden=false]:bottom-10 aria-[hidden=false]:opacity-100"
        )}
        aria-hidden={watch ? "true" : "false"}
      >
        <Button
          className="bg-white dark:bg-black"
          variant="outline"
          size="sm"
          onClick={() => {
            setWatch(true);
          }}
        >
          <ArrowDown />
          follow logs
          <ArrowDown />
        </Button>
      </div>
    </Logs>
  );
};

export default ScrollContainer;
