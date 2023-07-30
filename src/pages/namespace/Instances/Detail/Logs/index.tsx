import { FC, useEffect, useRef, useState } from "react";

import Button from "~/design/Button";
import Entry from "./Entry";
import { Logs } from "~/design/Logs";
import { twMergeClsx } from "~/util/helpers";
import { useLogs } from "~/api/logs/query/get";
import { useVirtualizer } from "@tanstack/react-virtual";

const LogsPanel: FC<{ instanceId: string }> = ({ instanceId }) => {
  const { data } = useLogs({ instanceId });
  const [watch, setWatch] = useState(true);

  // The scrollable element for the list
  const parentRef = useRef<HTMLDivElement | null>(null);

  // The virtualizer
  const rowVirtualizer = useVirtualizer({
    count: data?.results.length ?? 0,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 20,
    getItemKey: (index) => data?.results[index]?.t ?? index,
    // observeElementRect: (a, cb) => {
    //   cb({
    //     width: 10,
    //     height: 100,
    //   });
    //   console.log("---", a.scrollElement?.clientHeight);
    // },
  });

  useEffect(() => {
    if (data?.results.length && watch) {
      rowVirtualizer.scrollToIndex(data?.results.length);
    }
  }, [data?.results.length, rowVirtualizer, watch]);

  // useEffect(() => {
  //   rowVirtualizer.measure();
  // }, [rowVirtualizer]);

  if (!data) return null;

  // const height = true; // heightContainerRef.current?.clientHeight;

  return (
    <>
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
      </Logs>
      <Button
        className={twMergeClsx(
          "absolute bottom-0 mb-10 place-self-center bg-white dark:bg-black",
          // "transition-all duration-100"
          // "aria-[hidden=false]:fade-in",
          "aria-[hidden=true]:hidden"
          // "aria-[hidden=true]:animate-out"
          // watch && "animate-out"
        )}
        aria-hidden={watch ? "true" : "false"}
        variant="outline"
        size="sm"
        onClick={() => {
          setWatch((old) => !old);
        }}
      >
        {watch ? "stop following logs" : "follow logs"}
      </Button>
    </>
  );
};

export default LogsPanel;
