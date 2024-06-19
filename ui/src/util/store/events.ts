import { create } from "zustand";
import { persist } from "zustand/middleware";
import { z } from "zod";

export const eventsPageSizeValue = ["10", "20", "30", "50"] as const;
export const EventsPageSizeValueSchema = z.enum(eventsPageSizeValue);
export type EventsPageSizeValueType = z.infer<typeof EventsPageSizeValueSchema>;

const defaultPageSize: EventsPageSizeValueType = "10";

interface EventsPageSizeState {
  pagesize: EventsPageSizeValueType;
  actions: {
    setEventsPageSize: (
      EventsPageSize: EventsPageSizeState["pagesize"]
    ) => void;
  };
}

const useEventsPageSizeState = create<EventsPageSizeState>()(
  persist(
    (set) => ({
      pagesize: defaultPageSize,
      actions: {
        setEventsPageSize: (newEventsPageSize) =>
          set(() => ({ pagesize: newEventsPageSize })),
      },
    }),
    {
      name: "direktiv-store-pagesize",
      partialize: (state) => ({
        pagesize: state.pagesize,
      }),
    }
  )
);

export const useEventsPageSize = () =>
  useEventsPageSizeState((state) => state.pagesize);

export const useEventsPageSizeActions = () =>
  useEventsPageSizeState((state) => state.actions);
