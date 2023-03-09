import { create } from "zustand";
import { persist } from "zustand/middleware";

interface ApiKeyState {
  apiKey: string | null;
  actions: {
    setApiKey: (apiKey: ApiKeyState["apiKey"]) => void;
  };
}

export const useApiKeyState = create<ApiKeyState>()(
  persist(
    (set) => ({
      apiKey: null,
      actions: {
        setApiKey: (newApiKey) => set(() => ({ apiKey: newApiKey })),
      },
    }),
    {
      name: "directiv-store",
    }
  )
);

interface ThemeState {
  theme: "light" | "dark" | null;
  actions: {
    setTheme: (theme: ThemeState["theme"]) => void;
  };
}

export const useThemeState = create<ThemeState>()(
  persist(
    (set) => ({
      theme: null,
      actions: {
        setTheme: (newTheme) => set(() => ({ theme: newTheme })),
      },
    }),
    {
      name: "directiv-store",
    }
  )
);
