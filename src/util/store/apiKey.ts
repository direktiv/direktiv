import { create } from "zustand";
import { persist } from "zustand/middleware";

interface ApiKeyState {
  apiKey: string | null;
  requiresAuth: boolean | null;
  actions: {
    setApiKey: (apiKey: ApiKeyState["apiKey"]) => void;
    setRequiresAuth: (requiresAuth: ApiKeyState["requiresAuth"]) => void;
  };
}

const useApiKeyState = create<ApiKeyState>()(
  persist(
    (set) => ({
      apiKey: null,
      requiresAuth: null,
      actions: {
        setApiKey: (newApiKey) => set(() => ({ apiKey: newApiKey })),
        setRequiresAuth: (newRequiresAuth) =>
          set(() => ({ requiresAuth: newRequiresAuth })),
      },
    }),
    {
      name: "direktiv-store-api-key",
      partialize: (state) => ({
        apiKey: state.apiKey, // pick all fields to be persistent and don't persist actions
      }),
    }
  )
);

export const useApiKey = () => useApiKeyState((state) => state.apiKey);
export const useApiActions = () => useApiKeyState((state) => state.actions);
