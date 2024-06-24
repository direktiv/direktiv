import { create } from "zustand";
import { persist } from "zustand/middleware";
import { z } from "zod";

export const eventsPageSizeValue = ["10", "20", "30", "50"] as const;
export const EventsPageSizeValueSchema = z.enum(eventsPageSizeValue);
export type EventsPageSizeValueType = z.infer<typeof EventsPageSizeValueSchema>;

const defaultPageSize: EventsPageSizeValueType = "10";

interface EventsPageSizeState {
  pageSize: EventsPageSizeValueType;
  actions: {
    setEventsPageSize: (
      EventsPageSize: EventsPageSizeState["pageSize"]
    ) => void;
  };
}

const useEventsPageSizeState = create<EventsPageSizeState>()(
  persist(
    (set) => ({
      pageSize: defaultPageSize,
      actions: {
        setEventsPageSize: (newEventsPageSize) =>
          set(() => ({ pageSize: newEventsPageSize })),
      },
    }),
    {
      name: "direktiv-store-events-page-size",
      partialize: (state) => ({
        pageSize: state.pageSize,
      }),
    }
  )
);

export const useEventsPageSize = () =>
  useEventsPageSizeState((state) => state.pageSize);

export const useEventsPageSizeActions = () =>
  useEventsPageSizeState((state) => state.actions);
