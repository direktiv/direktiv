const availableLayouts = ["code"] as const;

export type LayoutsType = (typeof availableLayouts)[number];

interface EditorState {
  layout: LayoutsType;
  actions: {
    setLayout: (layout: EditorState["layout"]) => void;
  };
}

// setLayout is not in use currently
