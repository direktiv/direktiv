import { create } from "zustand";
import { persist } from "zustand/middleware";

type PanelState = {
  leftPanelWidth: number;
  setLeftPanelWidth: (width: number) => void;
};

const panelSize = create<PanelState>()(
  persist(
    (set) => ({
      leftPanelWidth: 65,
      setLeftPanelWidth: (width) =>
        set(() => ({
          leftPanelWidth: width,
        })),
    }),
    {
      name: "direktiv-store-panel-size",
    }
  )
);

export const useLeftPanelWidth = () =>
  panelSize((state) => state.leftPanelWidth);
export const useSetLeftPanelWidth = () =>
  panelSize((state) => state.setLeftPanelWidth);
