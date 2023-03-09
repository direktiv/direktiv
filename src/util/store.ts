import { create } from "zustand";
import { persist } from "zustand/middleware";

interface ApiKeyState {
  apiKey: string | null;
  setApiKey: (apiKey: ApiKeyState["apiKey"]) => void;
}

export const useApiKeyState = create<ApiKeyState>()(
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
