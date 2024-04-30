import Row from "../Row";
import { useEvents } from "~/api/eventsv2/query/get";
import { useRef } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";

const EventsScroller = () => {
  const parentRef = useRef<HTMLDivElement | null>(null);

  const { data } = useEvents({
    enabled: true,
  });

  const virtualizer = useVirtualizer({
    count: data?.length || 0,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 35,
  });

  if (!data || data.length === 0) return null;

  return (
    <>
      {/* The scrollable element for your list */}
      <div
        ref={parentRef}
        style={{
          height: `400px`,
          overflow: "auto", // Make it scroll!
        }}
      >
        {/* The large inner element to hold all of the items */}
        <div
          style={{
            height: `${virtualizer.getTotalSize()}px`,
            width: "100%",
            position: "relative",
          }}
        >
          {/* Only the visible items in the virtualizer, manually positioned to be in view */}
          {virtualizer.getVirtualItems().map((virtualItem) => {
            const item = data[virtualItem.index];

            if (!item) return null;

            return (
              <Row
                key={item.event.id}
                event={item.event}
                receivedAt={item.receivedAt}
                namespace={item.namespace}
                onClick={() => "To be done"}
              />
            );
          })}
        </div>
      </div>
    </>
  );
};

export default EventsScroller;
