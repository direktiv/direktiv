import { create } from "zustand";
import { persist } from "zustand/middleware";
import { z } from "zod";

export const eventListenersPageSizeValue = ["10", "20", "30", "50"] as const;
export const EventListenersPageSizeValueSchema = z.enum(
  eventListenersPageSizeValue
);
export type EventListenersPageSizeValueType = z.infer<
  typeof EventListenersPageSizeValueSchema
>;

const defaultPageSize: EventListenersPageSizeValueType = "10";

interface EventListenersPageSizeState {
  pageSize: EventListenersPageSizeValueType;
  actions: {
    setEventListenersPageSize: (
      EventListenersPageSize: EventListenersPageSizeState["pageSize"]
    ) => void;
  };
}

const useEventListenersPageSizeState = create<EventListenersPageSizeState>()(
  persist(
    (set) => ({
      pageSize: defaultPageSize,
      actions: {
        setEventListenersPageSize: (newEventListenersPageSize) =>
          set(() => ({ pageSize: newEventListenersPageSize })),
      },
    }),
    {
      name: "direktiv-store-eventlisteners-page-size",
      partialize: (state) => ({
        pageSize: state.pageSize,
      }),
    }
  )
);

export const useEventListenersPageSize = () =>
  useEventListenersPageSizeState((state) => state.pageSize);

export const useEventListenersPageSizeActions = () =>
  useEventListenersPageSizeState((state) => state.actions);
