import Entry from "./Entry";
import LogList from "~/components/Logs";
import { useInstanceDetails } from "~/api/instances/query/details";
import { useInstanceId } from "../../store/instanceContext";
import { useLogVirtualizer } from "~/components/Logs/useLogVirtualizer";

const ScrollContainer = () => {
  const instanceId = useInstanceId();
  const { data: instanceDetailsData } = useInstanceDetails({ instanceId });
  const isStreaming = instanceDetailsData?.instance?.status === "pending";
  const {
    rowVirtualizer,
    parentRef,
    logs,
    scrolledToBottom,
    setScrolledToBottom,
  } = useLogVirtualizer({
    queryLogsBy: {
      instance: instanceId,
    },
  });

  const virtualItems = rowVirtualizer.getVirtualItems();

  return (
    <LogList
      ref={parentRef}
      height={rowVirtualizer.getTotalSize()}
      virtualOffset={virtualItems[0]?.start ?? 0}
      isStreaming={isStreaming}
      scrolledToBottom={scrolledToBottom}
      setScrolledToBottom={setScrolledToBottom}
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
    </LogList>
  );
};

export default ScrollContainer;
