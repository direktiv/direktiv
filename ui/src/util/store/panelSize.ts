import { create } from "zustand";
import { persist } from "zustand/middleware";

type PanelState = {
  leftPanelWidth: number;
  actions: {
    setLeftPanelWidth: (width: number) => void;
  };
};

const panelSize = create<PanelState>()(
  persist(
    (set) => ({
      leftPanelWidth: 65,
      actions: {
        setLeftPanelWidth: (width) =>
          set(() => ({
            leftPanelWidth: width,
          })),
      },
    }),
    {
      name: "direktiv-store-panel-size",
    }
  )
);

export const useLeftPanelWidth = () =>
  panelSize((state) => state.leftPanelWidth);
export const useSetLeftPanelWidth = () =>
  panelSize((state) => state.actions.setLeftPanelWidth);
