import { ArrowDown } from "lucide-react";
import Button from "~/design/Button";
import Entry from "./Entry";
import { Logs } from "~/design/Logs";
import { twMergeClsx } from "~/util/helpers";
import { useInstanceDetails } from "~/api/instances/query/details";
import { useInstanceId } from "../../store/instanceContext";
import { useLogVirtualizer } from "./useLogVirtualizer";
import { useLogs } from "~/api/logs/query/logs";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";

const ScrollContainer = () => {
  const instanceId = useInstanceId();
  const { data: instanceDetailsData } = useInstanceDetails({ instanceId });

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

  const { rowVirtualizer, parentRef, setWatch, watch } = useLogVirtualizer({
    logs: allLogs,
    fetchPreviousPage,
    hasPreviousPage,
    isFetchingPreviousPage,
  });

  const isPending = instanceDetailsData?.instance?.status === "pending";

  if (!logData) return null;

  const items = rowVirtualizer.getVirtualItems();

  return (
    <Logs
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
