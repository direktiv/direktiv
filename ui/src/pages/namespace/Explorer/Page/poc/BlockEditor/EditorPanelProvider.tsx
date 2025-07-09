import { Dialog, DialogContent } from "~/design/Dialog";
import { createContext, useContext, useState } from "react";
import {
  usePageEditor,
  usePageStateContext,
} from "../PageCompiler/context/pageCompilerContext";

import { AllBlocksType } from "../schema/blocks";
import { BlockDeleteForm } from "./components/Delete";
import { BlockPathType } from "../PageCompiler/Block";
import { EditorPanel } from "./components/EditorPanelContent";
import { LocalDialogContainer } from "~/components/LocalDialog";

type EditorPanelState = null | {
  action: "create" | "edit" | "delete";
  block: AllBlocksType;
  path: BlockPathType;
};

type EditorPanelContextType = {
  panel: EditorPanelState;
  setPanel: React.Dispatch<React.SetStateAction<EditorPanelState>>;
};

const EditorPanelContext = createContext<EditorPanelContextType | null>(null);

export const EditorPanelLayoutProvider = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const { deleteBlock } = usePageEditor();
  const [panel, setPanel] = useState<EditorPanelState>(null);
  const { mode } = usePageStateContext();

  if (mode === "edit") {
    return (
      <EditorPanelContext.Provider value={{ panel, setPanel }}>
        <div className="flex gap-5">
          <div className="w-1/3 max-w-md shrink-0 overflow-x-hidden">
            <EditorPanel />
          </div>
          <LocalDialogContainer className="min-w-0 flex-1">
            {children}
          </LocalDialogContainer>
        </div>
        <Dialog open={panel && panel.action === "delete" ? true : false}>
          <DialogContent>
            {!!panel && (
              <BlockDeleteForm
                path={panel.path}
                onSubmit={() => {
                  deleteBlock(panel.path);
                  setPanel(null);
                }}
                onCancel={() => {
                  setPanel({ ...panel, action: "edit" });
                }}
              />
            )}
          </DialogContent>
        </Dialog>
      </EditorPanelContext.Provider>
    );
  }

  return <LocalDialogContainer>{children}</LocalDialogContainer>;
};

export const usePageEditorPanel = () => {
  const context = useContext(EditorPanelContext);
  if (!context)
    throw new Error(
      "usePageEditorPanel must be used within EditorPanelLayoutProvider"
    );
  return context;
};
