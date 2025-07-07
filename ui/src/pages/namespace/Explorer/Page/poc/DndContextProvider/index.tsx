import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";
import { DndContext } from "~/design/DragAndDropEditor/Context.tsx";
import { PropsWithChildren } from "react";
import { usePageEditor } from "../PageCompiler/context/pageCompilerContext";

export const DndContextProvider = ({ children }: PropsWithChildren) => {
  const { addBlock, moveBlock } = usePageEditor();

  const onMove = (
    origin: BlockPathType | null,
    target: BlockPathType,
    block: AllBlocksType
  ) => {
    origin === null
      ? addBlock(target, block)
      : moveBlock(origin, target, block);
  };

  return <DndContext onMove={onMove}>{children} </DndContext>;
};
