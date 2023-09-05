import { create } from "zustand";
import { persist } from "zustand/middleware";

interface JqPlaygroundState {
  input: string | null;
  query: string | null;
  actions: {
    setInput: (apiKey: JqPlaygroundState["input"]) => void;
    setQuery: (apiKey: JqPlaygroundState["query"]) => void;
  };
}

const useJqPlaygroundState = create<JqPlaygroundState>()(
  persist(
    (set) => ({
      input: null,
      query: null,
      actions: {
        setInput: (newInput) => set(() => ({ input: newInput })),
        setQuery: (newQuery) => set(() => ({ query: newQuery })),
      },
    }),
    {
      name: "direktiv-store-jq-playground",
      partialize: (state) => ({
        input: state.input, // pick all fields to be persistent and don't persist actions
        query: state.query,
      }),
    }
  )
);

export const useJqPlaygroundInput = () =>
  useJqPlaygroundState((state) => state.input);

export const useJqPlaygroundQuery = () =>
  useJqPlaygroundState((state) => state.query);

export const useJqPlaygroundActions = () =>
  useJqPlaygroundState((state) => state.actions);
