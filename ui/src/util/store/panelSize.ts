import { create } from "zustand";
import { persist } from "zustand/middleware";

type PanelState = {
  panelWidth: number;
  setPanelWidth: (width: number) => void;
};

const panelSize = create<PanelState>()(
  persist(
    (set) => ({
      panelWidth: 65,
      setPanelWidth: (width) =>
        set(() => ({
          panelWidth: width,
        })),
    }),
    {
      name: "direktiv-store-panel-size",
    }
  )
);

export const usePanelSize = () => panelSize((state) => state.panelWidth);
export const useSetPanelSize = () => panelSize((state) => state.setPanelWidth);
