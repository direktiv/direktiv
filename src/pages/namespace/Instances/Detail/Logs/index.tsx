import { FC, useEffect, useRef } from "react";

import Entry from "./Entry";
import { Logs } from "~/design/Logs";
import { useLogs } from "~/api/logs/query/get";
import { useVirtualizer } from "@tanstack/react-virtual";

const LogsPanel: FC<{ instanceId: string; stream: boolean }> = ({
  instanceId,
  stream,
}) => {
  const { data } = useLogs({ instanceId }, { stream });

  // The scrollable element for the list
  const parentRef = useRef<HTMLDivElement | null>(null);

  // The virtualizer
  const rowVirtualizer = useVirtualizer({
    count: data?.results.length ?? 0,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 35,
    getItemKey: (index) => data?.results[index]?.t ?? index,
  });

  useEffect(() => {
    if (data?.results.length) {
      rowVirtualizer.scrollToIndex(data?.results.length);
    }
  }, [data?.results.length, rowVirtualizer]);

  if (!data) return null;

  return (
    <Logs
      linewrap={true}
      className="grow"
      ref={parentRef}
      style={{
        height: `800px`,
        overflow: "auto", // make it scroll
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
  );
};

export default LogsPanel;
