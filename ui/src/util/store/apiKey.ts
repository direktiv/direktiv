import { create } from "zustand";
import { getOidcUser } from "~/components/OidcProvider/utils";
import { isEnterprise } from "~/config/env/utils";
import { persist } from "zustand/middleware";

interface ApiKeyState {
  apiKey: string | null;
  actions: {
    setApiKey: (apiKey: ApiKeyState["apiKey"]) => void;
  };
}

const useApiKeyState = create<ApiKeyState>()(
  persist(
    (set) => ({
      apiKey: null,
      actions: {
        setApiKey: (newApiKey) => set(() => ({ apiKey: newApiKey })),
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

export const useApiKey = () =>
  useApiKeyState((state) => {
    if (isEnterprise()) {
      /**
       * in enterprise mode we need to read from a different source
       */
      const user = getOidcUser();
      return user?.access_token ?? null;
    }
    return state.apiKey;
  });

export const useApiActions = () => useApiKeyState((state) => state.actions);
