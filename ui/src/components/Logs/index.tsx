import { ElementRef, PropsWithChildren, forwardRef } from "react";

import { ArrowDown } from "lucide-react";
import Button from "~/design/Button";
import { Logs } from "~/design/Logs";
import { twMergeClsx } from "~/util/helpers";
import { useTranslation } from "react-i18next";

type LogRef = ElementRef<typeof Logs>;

type LogsContainerProps = {
  height: number;
  virtualOffset: number;
  isStreaming?: boolean;
  scrolledToBottom: boolean;
  setScrolledToBottom: (scrolledToBottom: boolean) => void;
} & PropsWithChildren;

export const LogList = forwardRef<LogRef, LogsContainerProps>(
  (
    {
      height,
      virtualOffset,
      scrolledToBottom,
      isStreaming,
      setScrolledToBottom,
      children,
    },
    ref
  ) => {
    const { t } = useTranslation();

    return (
      <Logs className="h-full overflow-scroll" ref={ref}>
        <div
          className="relative w-full"
          style={{
            height: `${height}px`,
            position: "relative",
          }}
        >
          <div
            style={{
              position: "absolute",
              top: 0,
              left: 0,
              width: "100%",
              transform: `translateY(${virtualOffset}px)`,
            }}
          >
            {children}
          </div>
        </div>
        {isStreaming && (
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
              {t("pages.monitoring.logs.followLogs")} {/* TODO:  */}
              <ArrowDown />
            </Button>
          </div>
        )}
      </Logs>
    );
  }
);

LogList.displayName = "LogsContainer";

export default LogList;
