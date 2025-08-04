import { Dialog, DialogContent } from "~/design/Dialog";
import { PropsWithChildren, createContext, useContext, useState } from "react";
import {
  usePageEditor,
  usePageStateContext,
} from "../PageCompiler/context/pageCompilerContext";

import { BlockDeleteForm } from "./components/Delete";
import { BlockPathType } from "../PageCompiler/Block";
import { BlockType } from "../schema/blocks";
import { DndContext } from "~/design/DragAndDrop";
import { DragAndDropPayloadSchemaType } from "~/design/DragAndDrop/schema";
import { EditorPanel } from "./components/EditorPanel";
import { LocalDialogContainer } from "~/design/LocalDialog/container";
import { useBlockTypes } from "../PageCompiler/context/utils/useBlockTypes";

type EditorPanelState =
  | null
  | {
      action: null;
      dialog?: BlockPathType | null;
    }
  | {
      action: "create" | "edit" | "delete";
      block: BlockType;
      path: BlockPathType;
      dialog?: BlockPathType | null;
    };

type EditorPanelContextType = {
  panel: EditorPanelState;
  setPanel: React.Dispatch<React.SetStateAction<EditorPanelState>>;
};

const EditorPanelContext = createContext<EditorPanelContextType | null>(null);

const PagePreviewContainer = ({ children }: PropsWithChildren) => (
  <div className="grow px-3 py-5 sm:h-[calc(100vh-230px)] sm:overflow-y-scroll">
    {children}
  </div>
);

export const EditorPanelLayoutProvider = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const { getBlockConfig } = useBlockTypes();

  const { addBlock, deleteBlock, moveBlock } = usePageEditor();
  const [panel, setPanel] = useState<EditorPanelState>(null);
  const { mode } = usePageStateContext();

  const createBlock = (type: BlockType["type"], path: BlockPathType) => {
    const blockConfig = getBlockConfig(type);

    if (!blockConfig) throw new Error(`No blockConfig found for ${type}`);

    if (!blockConfig.formComponent) {
      return addBlock(path, blockConfig.defaultValues);
    }
    setPanel({
      action: "create",
      block: blockConfig.defaultValues,
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
          <div className="grow sm:grid sm:grid-cols-[350px_1fr]">
            <div className="h-[300px] overflow-y-visible border-b-2 border-gray-4 p-3 dark:border-gray-dark-4 sm:h-[calc(100vh-230px)] sm:border-b-0 sm:border-r-2">
              <EditorPanel />
            </div>
            <PagePreviewContainer>
              <LocalDialogContainer className="min-w-0 flex-1">
                {children}
              </LocalDialogContainer>
            </PagePreviewContainer>
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

  return (
    <LocalDialogContainer>
      <PagePreviewContainer>{children}</PagePreviewContainer>
    </LocalDialogContainer>
  );
};

export const usePageEditorPanel = () => {
  const context = useContext(EditorPanelContext);
  if (!context)
    throw new Error(
      "usePageEditorPanel must be used within EditorPanelLayoutProvider"
    );
  return context;
};
