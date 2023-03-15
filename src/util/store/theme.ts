import { create } from "zustand";
import { persist } from "zustand/middleware";

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
      name: "directiv-store-theme",
      partialize: (state) => ({
        theme: state.theme, // pick all fields to persistend, and don't persist actions
      }),
    }
  )
);

export const useTheme = () => useThemeState((state) => state.theme);

export const useThemeActions = () => useThemeState((state) => state.actions);
