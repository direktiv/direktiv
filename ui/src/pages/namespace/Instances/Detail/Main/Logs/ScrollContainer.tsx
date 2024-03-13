import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { ArrowDown } from "lucide-react";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Entry from "./Entry";
import { Logs } from "~/design/Logs";
import { twMergeClsx } from "~/util/helpers";
import { useInstanceDetails } from "~/api/instances/query/details";
import { useInstanceId } from "../../store/instanceContext";
import { useLogs } from "~/api/logs/query/logs";
import { useLogsPreferencesWordWrap } from "~/util/store/logs";
import { useTranslation } from "react-i18next";
import { useVirtualizer } from "@tanstack/react-virtual";

const defaultLogHeight = 20;
const ScrollContainer = () => {
  const instanceId = useInstanceId();
  const wordWrap = useLogsPreferencesWordWrap();
  const { data: instanceDetailsData } = useInstanceDetails({ instanceId });
  const lastScrollPos = useRef<{
    scrollOffset: number;
    numberOfLogs: number;
  } | null>(null);

  const { t } = useTranslation();

  const {
    data: logData,
    hasPreviousPage,
    fetchPreviousPage,
    isFetchingPreviousPage,
  } = useLogs({
    instance: instanceId,
  });

  const allLogs = useMemo(
    () => (logData?.pages ?? []).flatMap((x) => x.data ?? []),
    [logData?.pages]
  );

  const allLogIds = new Set(allLogs.map((log) => log.id));

  if (allLogs.length !== allLogIds.size) {
    throw new Error("Duplicate log ids found");
  }

  const numberOfLogs = allLogs.length;

  const [watch, setWatch] = useState(true);

  // The scrollable element for the list
  const parentRef = useRef<HTMLDivElement | null>(null);

  // The virtualizer
  const rowVirtualizer = useVirtualizer({
    count: numberOfLogs,
    getScrollElement: () => parentRef.current,
    estimateSize: () => defaultLogHeight,
    getItemKey: useCallback(
      (index: number) => {
        const uniqueId = allLogs[index]?.id;
        if (!uniqueId)
          throw new Error("Could not find a log id for the virtualizer.");
        return uniqueId;
      },
      [allLogs]
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

  const virtualItems = rowVirtualizer.getVirtualItems();
  const [firstLogEntry] = virtualItems;

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

  const isPending = instanceDetailsData?.instance?.status === "pending";

  if (!logData) return null;

  const items = rowVirtualizer.getVirtualItems();
  const range = rowVirtualizer.range;
  const [lastLineOffset] = rowVirtualizer.getOffsetForIndex(numberOfLogs - 1);
  const progress =
    ((lastScrollPos.current?.scrollOffset ?? 0) / lastLineOffset) * 100;

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
            const logEntry = allLogs[virtualItem.index];
            if (!logEntry) return null;
            return (
              <Entry
                key={virtualItem.key}
                test={virtualItem.key as number}
                data-index={virtualItem.key}
                ref={rowVirtualizer.measureElement}
                logEntry={logEntry}
              />
            );
          })}
        </div>
      </div>
      <Card className="absolute left-0 -top-28 box-border flex w-full items-center justify-center gap-3 bg-white p-4 pr-10">
        <div className="flex flex-col gap-1">
          <Button
            size="sm"
            onClick={() => {
              rowVirtualizer.scrollToIndex(0, { align: "start" });
            }}
          >
            top
          </Button>
          <Button
            size="sm"
            onClick={() => {
              rowVirtualizer.scrollToIndex(300, { align: "start" });
            }}
          >
            bottom
          </Button>
        </div>
        <div className="grid grow grid-cols-2 gap-3">
          <div className="flex flex-col gap-1">
            1st virtual idx {firstLogEntry?.index}
            <br />
            1st visual idx {range?.startIndex}
          </div>
          <div className="flex flex-col gap-1">
            offset {rowVirtualizer.scrollOffset} (
            {lastScrollPos.current?.scrollOffset})<br /> watch{" "}
            {watch ? "✅" : "❌"}
          </div>
        </div>
        <div className="flex h-1 w-40 items-center rounded-sm bg-gray-9 p-1">
          <div
            className="h-1 bg-white"
            style={{
              width: `${progress}%`,
            }}
          ></div>
        </div>
      </Card>
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
