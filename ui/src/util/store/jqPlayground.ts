import { create } from "zustand";
import { persist } from "zustand/middleware";

interface JqPlaygroundState {
  data: string | null;
  jx: string | null;
  actions: {
    setData: (apiKey: JqPlaygroundState["data"]) => void;
    setJx: (apiKey: JqPlaygroundState["jx"]) => void;
  };
}

const useJqPlaygroundState = create<JqPlaygroundState>()(
  persist(
    (set) => ({
      data: null,
      jx: null,
      actions: {
        setData: (newData) => set(() => ({ data: newData })),
        setJx: (newJx) => set(() => ({ jx: newJx })),
      },
    }),
    {
      name: "direktiv-store-jq-playground",
      partialize: ({ data, jx }) => ({
        // pick all fields to be persistent and don't persist actions
        data,
        jx,
      }),
    }
  )
);

export const useJqPlaygroundData = () =>
  useJqPlaygroundState((state) => state.data);

export const useJqPlaygroundJx = () =>
  useJqPlaygroundState((state) => state.jx);

export const useJqPlaygroundActions = () =>
  useJqPlaygroundState((state) => state.actions);
