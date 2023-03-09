import { create } from "zustand";
import { persist } from "zustand/middleware";

interface GlobalState {
  apiKey: string | null;
  setApiKey: (apiKey: GlobalState["apiKey"]) => void;
}

export const useGlobalStore = create<GlobalState>()(
  persist(
    (set) => ({
      apiKey: null,
      setApiKey: (newApiKey) => set(() => ({ apiKey: newApiKey })),
    }),
    {
      name: "directiv-store",
    }
  )
);
