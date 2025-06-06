import { create } from "zustand";
import { persist } from "zustand/middleware";

type ThemeValueType = "light" | "dark" | null;

interface ThemeState {
  // the theme that the user actively selected, will be store in local storage
  storedTheme: ThemeValueType;
  // the theme that the user's system is using, will be determined very early
  // in the app lifecycle but not stored in local storage
  systemTheme: ThemeValueType;
  actions: {
    setTheme: (theme: ThemeState["storedTheme"]) => void;
    setSystemTheme: (theme: ThemeState["systemTheme"]) => void;
  };
}

const useThemeState = create<ThemeState>()(
  persist(
    (set) => ({
      storedTheme: null,
      systemTheme: null,
      actions: {
        setTheme: (newTheme) => set(() => ({ storedTheme: newTheme })),
        setSystemTheme: (newTheme) => set(() => ({ systemTheme: newTheme })),
      },
    }),
    {
      name: "direktiv-store-theme",
      partialize: (state) => ({
        storedTheme: state.storedTheme, // pick all fields to be persistent and don't persist actions
      }),
    }
  )
);

export const useTheme = () =>
  useThemeState((state) => state.storedTheme ?? state.systemTheme);

export const useThemeActions = () => useThemeState((state) => state.actions);
