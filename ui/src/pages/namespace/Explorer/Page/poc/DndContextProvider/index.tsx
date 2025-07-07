import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";
import { DirektivPagesType } from "../schema";
import { DndContext } from "~/design/DragAndDropEditor/Context.tsx";
import { PropsWithChildren } from "react";
import { usePageEditor } from "../PageCompiler/context/pageCompilerContext";

type DndContextProviderProps = PropsWithChildren & {
  page: DirektivPagesType;
  setPage: (page: DirektivPagesType) => void;
};

export const DndContextProvider = ({ children }: DndContextProviderProps) => {
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
