import { ArrowDown } from "lucide-react";
import Button from "~/design/Button";
import Entry from "./Entry";
import { Logs } from "~/design/Logs";
import { twMergeClsx } from "~/util/helpers";
import { useLogVirtualizer } from "~/hooks/useLogVirtualizer";
import { useTranslation } from "react-i18next";

const ScrollContainer = ({ activityId }: { activityId: string }) => {
  const { t } = useTranslation();

  const {
    rowVirtualizer,
    parentRef,
    logs,
    scrolledToBottom,
    setScrolledToBottom,
  } = useLogVirtualizer({
    queryLogsBy: {
      activity: activityId,
    },
  });

  const virtualItems = rowVirtualizer.getVirtualItems();

  return (
    <Logs className="h-full overflow-scroll" ref={parentRef}>
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
            transform: `translateY(${virtualItems[0]?.start}px)`,
          }}
        >
          {virtualItems.map((virtualItem) => {
            const logEntry = logs[virtualItem.index];
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
      <div
        className={twMergeClsx(
          "absolute box-border flex w-full pr-10",
          "justify-center transition-all",
          "aria-[hidden=true]:pointer-events-none aria-[hidden=true]:bottom-11 aria-[hidden=true]:opacity-0",
          "aria-[hidden=false]:bottom-16 aria-[hidden=false]:opacity-100"
        )}
        aria-hidden={scrolledToBottom ? "true" : "false"}
      >
        <Button
          className="bg-white dark:bg-black"
          variant="outline"
          size="sm"
          onClick={() => {
            setScrolledToBottom(true);
          }}
        >
          <ArrowDown />
          {t("pages.monitoring.logs.followLogs")}
          <ArrowDown />
        </Button>
      </div>
    </Logs>
  );
};

export default ScrollContainer;
