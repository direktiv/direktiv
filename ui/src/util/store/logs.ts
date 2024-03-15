import { create } from "zustand";
import { persist } from "zustand/middleware";

export type InstanceLayout = "none" | "logs" | "diagram" | "input-output";

interface LogsPreferencesState {
  maximizedPanel: InstanceLayout;
  verboseLogs: boolean;
  actions: {
    setMaximizedPanel: (layout: LogsPreferencesState["maximizedPanel"]) => void;
    setVerboseLogs: (layout: LogsPreferencesState["verboseLogs"]) => void;
  };
}

const useLogsPreferencesState = create<LogsPreferencesState>()(
  persist(
    (set) => ({
      maximizedPanel: "none",
      verboseLogs: false,
      actions: {
        setMaximizedPanel: (newVal) => set(() => ({ maximizedPanel: newVal })),
        setVerboseLogs: (newVal) => set(() => ({ verboseLogs: newVal })),
      },
    }),
    {
      name: "direktiv-store-logspreferences",
      partialize: (state) => ({
        // pick all fields to be persistent and don't persist actions
        maximizedPanel: state.maximizedPanel,
        verboseLogs: state.verboseLogs,
      }),
    }
  )
);

export const useLogsPreferencesMaximizedPanel = () =>
  useLogsPreferencesState((state) => state.maximizedPanel);

export const useLogsPreferencesVerboseLogs = () =>
  useLogsPreferencesState((state) => state.verboseLogs);

export const useLogsPreferencesActions = () =>
  useLogsPreferencesState((state) => state.actions);
