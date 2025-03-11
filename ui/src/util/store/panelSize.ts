import { create } from "zustand";

type PanelSizeState = {
  leftPanelWidth: number;
  actions: {
    setLeftPanelWidth: (width: number) => void;
  };
};

const usePanelSizeStore = create<PanelSizeState>((set) => ({
  leftPanelWidth: 65,
  actions: {
    setLeftPanelWidth: (width: number) => set({ leftPanelWidth: width }),
  },
}));

export const useLeftPanelWidth = () =>
  usePanelSizeStore((state) => state.leftPanelWidth);

export const usePanelSizeActions = () =>
  usePanelSizeStore((state) => state.actions);
