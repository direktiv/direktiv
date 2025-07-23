import { AllBlocksType, inlineBlockTypes } from "../schema/blocks";
import { Dialog, DialogContent } from "~/design/Dialog";
import { createContext, useContext, useState } from "react";
import {
  usePageEditor,
  usePageStateContext,
} from "../PageCompiler/context/pageCompilerContext";

import { BlockDeleteForm } from "./components/Delete";
import { BlockPathType } from "../PageCompiler/Block";
import { DndContext } from "~/design/DragAndDrop";
import { DragAndDropPayloadSchemaType } from "~/design/DragAndDrop/schema";
import { EditorPanel } from "./components/EditorPanel";
import { LocalDialogContainer } from "~/design/LocalDialog/container";
import { getBlockTemplate } from "../PageCompiler/context/utils";

type EditorPanelState =
  | null
  | {
      action: null;
      dialog?: BlockPathType | null;
    }
  | {
      action: "create" | "edit" | "delete";
      block: AllBlocksType;
      path: BlockPathType;
      dialog?: BlockPathType | null;
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
  const { addBlock, deleteBlock, moveBlock } = usePageEditor();
  const [panel, setPanel] = useState<EditorPanelState>(null);
  const { mode } = usePageStateContext();

  const createBlock = (type: AllBlocksType["type"], path: BlockPathType) => {
    if (inlineBlockTypes.has(type)) {
      return addBlock(path, getBlockTemplate(type));
    }
    setPanel({
      action: "create",
      block: getBlockTemplate(type),
      path,
    });
  };

  const onDrop = (payload: DragAndDropPayloadSchemaType) => {
    const { drag, drop } = payload;
    if (drag.type === "add") {
      createBlock(drag.blockType, [...drop.targetPath]);
    }
    if (drag.type === "move") {
      moveBlock(drag.originPath, drop.targetPath, drag.block);
    }
  };

  if (mode === "edit") {
    return (
      <DndContext onDrop={onDrop}>
        <EditorPanelContext.Provider value={{ panel, setPanel }}>
          <div className="flex gap-5">
            <div className="w-1/3 max-w-md shrink-0 overflow-visible border-r-2 border-gray-4 pr-2 dark:border-gray-dark-4">
              <EditorPanel />
            </div>
            <LocalDialogContainer className="min-w-0 flex-1">
              {children}
            </LocalDialogContainer>
          </div>
          <Dialog open={panel && panel.action === "delete" ? true : false}>
            <DialogContent>
              {!!panel && panel.action === "delete" && (
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
      </DndContext>
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
