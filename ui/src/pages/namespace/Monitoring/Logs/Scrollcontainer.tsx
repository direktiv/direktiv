import Entry from "./Entry";
import LogList from "~/components/Logs";
import { useLogVirtualizer } from "~/components/Logs/useLogVirtualizer";

const ScrollContainer = () => {
  const {
    rowVirtualizer,
    parentRef,
    logLines,
    scrolledToBottom,
    setScrolledToBottom,
  } = useLogVirtualizer();

  const virtualItems = rowVirtualizer.getVirtualItems();

  return (
    <LogList
      ref={parentRef}
      height={rowVirtualizer.getTotalSize()}
      virtualOffset={virtualItems[0]?.start ?? 0}
      scrolledToBottom={scrolledToBottom}
      setScrolledToBottom={setScrolledToBottom}
    >
      {virtualItems.map((virtualItem) => {
        const logEntry = logLines[virtualItem.index];
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
