import { create } from "zustand";
import { persist } from "zustand/middleware";
import { z } from "zod";

export const pageSizeValue = ["10", "20", "30", "50"] as const;
export const PageSizeValueSchema = z.enum(pageSizeValue);
export type PageSizeValueType = z.infer<typeof PageSizeValueSchema>;

const defaultPageSize: PageSizeValueType = "10";

interface PageSizeState {
  events: PageSizeValueType;
  eventListeners: PageSizeValueType;
  actions: {
    setEventsPageSize: (EventsPageSize: PageSizeState["events"]) => void;
    setEventListenersPageSize: (
      EventsPageSize: PageSizeState["events"]
    ) => void;
  };
}

const usePageSizeState = create<PageSizeState>()(
  persist(
    (set) => ({
      events: defaultPageSize,
      eventListeners: defaultPageSize,
      actions: {
        setEventsPageSize: (newEventsPageSize) =>
          set(() => ({ events: newEventsPageSize })),
        setEventListenersPageSize(EventsPageSize) {
          set(() => ({ eventListeners: EventsPageSize }));
        },
      },
    }),
    {
      name: "direktiv-store-page-size",
      // pick all fields to be persistent and don't persist actions
      partialize: ({ actions: _, ...rest }) => ({
        ...rest,
      }),
    }
  )
);

export const useEventsPageSize = () =>
  usePageSizeState((state) => state.events);

export const useEventListenersPageSize = () =>
  usePageSizeState((state) => state.eventListeners);

export const usePageSizeActions = () =>
  usePageSizeState((state) => state.actions);
