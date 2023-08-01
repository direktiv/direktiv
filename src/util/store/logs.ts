import { create } from "zustand";
import { persist } from "zustand/middleware";

interface LogsPreferencesState {
  maximizedPanel: "none" | "logs" | "diagram" | "input" | "output";
  wordWrap: boolean;
  verboseLogs: boolean;
  actions: {
    setMaximizedPanel: (layout: LogsPreferencesState["maximizedPanel"]) => void;
    setWordWrap: (layout: LogsPreferencesState["wordWrap"]) => void;
    setVerboseLogs: (layout: LogsPreferencesState["verboseLogs"]) => void;
  };
}

const useLogsPreferencesState = create<LogsPreferencesState>()(
  persist(
    (set) => ({
      maximizedPanel: "none",
      wordWrap: false,
      verboseLogs: false,
      actions: {
        setMaximizedPanel: (newVal) => set(() => ({ maximizedPanel: newVal })),
        setWordWrap: (newVal) => set(() => ({ wordWrap: newVal })),
        setVerboseLogs: (newVal) => set(() => ({ verboseLogs: newVal })),
      },
    }),
    {
      name: "direktiv-store-logspreferences",
      partialize: (state) => ({
        // pick all fields to be persistent and don't persist actions
        wordWrap: state.wordWrap,
        verboseLogs: state.verboseLogs,
      }),
    }
  )
);

export const useLogsPreferencesMaximizedPanel = () =>
  useLogsPreferencesState((state) => state.maximizedPanel);

export const useLogsPreferencesWordWrap = () =>
  useLogsPreferencesState((state) => state.wordWrap);

export const useLogsPreferencesVerboseLogs = () =>
  useLogsPreferencesState((state) => state.verboseLogs);

export const useLogsPreferencesActions = () =>
  useLogsPreferencesState((state) => state.actions);
