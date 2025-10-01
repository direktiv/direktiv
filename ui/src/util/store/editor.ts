import { create } from "zustand";
import { persist } from "zustand/middleware";

const availableLayouts = ["code"] as const;

export type LayoutsType = (typeof availableLayouts)[number];

interface EditorState {
  layout: LayoutsType;
  actions: {
    setLayout: (layout: EditorState["layout"]) => void;
  };
}

const useEditorState = create<EditorState>()(
  persist(
    (set) => ({
      layout: availableLayouts[0],
      actions: {
        setLayout: (newLayout) => set(() => ({ layout: newLayout })),
      },
    }),
    {
      name: "direktiv-store-editor",
      partialize: (state) => ({
        layout: state.layout, // pick all fields to be persistent and don't persist actions
      }),
    }
  )
);

export const useEditorLayout = () => useEditorState((state) => state.layout);
