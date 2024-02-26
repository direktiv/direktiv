import { useEffect, useRef, useState } from "react";
import { useFilters, useInstanceId } from "../../store/instanceContext";

import { ArrowDown } from "lucide-react";
import Button from "~/design/Button";
import Entry from "./Entry";
import { Logs } from "~/design/Logs";
import { twMergeClsx } from "~/util/helpers";
import { useInstanceDetails } from "~/api/instances/query/details";
import { useLogs } from "~/api/logs_DEBRECATED/query/get";
import { useLogsPreferencesWordWrap } from "~/util/store/logs";
import { useTranslation } from "react-i18next";
import { useVirtualizer } from "@tanstack/react-virtual";

const ScrollContainer = () => {
  const instanceId = useInstanceId();
  const wordWrap = useLogsPreferencesWordWrap();
  const { data: instanceDetailsData } = useInstanceDetails({ instanceId });

  const { t } = useTranslation();

  const filters = useFilters();

  const { data: logData } = useLogs({
    instanceId,
    filters,
  });
  const [watch, setWatch] = useState(true);

  // The scrollable element for the list
  const parentRef = useRef<HTMLDivElement | null>(null);

  // The virtualizer
  const rowVirtualizer = useVirtualizer({
    count: logData?.results.length ?? 0,
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
    if (logData?.results.length && watch) {
      rowVirtualizer.scrollToIndex(logData?.results.length - 1),
        { align: "end" };
    }
  }, [logData?.results.length, rowVirtualizer, watch]);

  const isPending = instanceDetailsData?.instance?.status === "pending";

  if (!logData) return null;

  const items = rowVirtualizer.getVirtualItems();

  return (
    <Logs
      wordWrap={wordWrap}
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
        <div
          style={{
            position: "absolute",
            top: 0,
            left: 0,
            width: "100%",
            transform: `translateY(${items[0]?.start}px)`,
          }}
        >
          {items.map((virtualItem) => {
            const logEntry = logData.results[virtualItem.index];
            if (!logEntry) return null;
            return (
              <Entry
                key={virtualItem.key}
                data-index={virtualItem.key}
                ref={rowVirtualizer.measureElement}
                logEntry={logEntry}
              />
            );
          })}
        </div>
      </div>
      {isPending && (
        <div
          className={twMergeClsx(
            "absolute box-border flex w-full pr-10",
            "justify-center transition-all",
            "aria-[hidden=true]:pointer-events-none aria-[hidden=true]:bottom-11 aria-[hidden=true]:opacity-0",
            "aria-[hidden=false]:bottom-16 aria-[hidden=false]:opacity-100"
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
            {t("pages.instances.detail.logs.followLogs")}
            <ArrowDown />
          </Button>
        </div>
      )}
    </Logs>
  );
};

export default ScrollContainer;
