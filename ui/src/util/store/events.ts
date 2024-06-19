import { create } from "zustand";
import { persist } from "zustand/middleware";

type EventsPageSizeValueType = 10 | 20 | 30 | 50 | null;

interface EventsPageSizeState {
  pagesize: EventsPageSizeValueType;
  actions: {
    setEventsPageSize: (
      EventsPageSize: EventsPageSizeState["pagesize"]
    ) => void;
  };
}

export const useEventsPageSizeState = create<EventsPageSizeState>()(
  persist(
    (set) => ({
      pagesize: null,
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
