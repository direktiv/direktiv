import { create } from "zustand";
import { persist } from "zustand/middleware";

type PanelState = {
  panelWidth: number;
  setPanelWidth: (width: number) => void;
};

export const panelStore = create<PanelState>()(
  persist(
    (set) => ({
      panelWidth: 65,
      setPanelWidth: (width) =>
        set(() => ({
          panelWidth: width,
        })),
    }),
    {
      name: "panel-store",
    }
  )
);
