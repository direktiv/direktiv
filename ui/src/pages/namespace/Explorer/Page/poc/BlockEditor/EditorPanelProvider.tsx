import { AllBlocksType, inlineBlockTypes } from "../schema/blocks";
import { Dialog, DialogContent } from "~/design/Dialog";
import { PropsWithChildren, createContext, useContext, useState } from "react";
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

const PagePreviewContainer = ({ children }: PropsWithChildren) => (
  <div className="grow p-3 lg:h-[calc(100vh-230px)] lg:overflow-y-scroll">
    {children}
  </div>
);

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
          <div className="grow gap-5 lg:flex">
            <div className="h-[300px] overflow-y-visible border-b-2 border-gray-4 p-3 dark:border-gray-dark-4 lg:h-[calc(100vh-230px)] lg:w-1/3 lg:border-b-0 lg:border-r-2">
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
