import {
  addBlockToPage,
  moveBlockWithinPage,
} from "../PageCompiler/context/utils";

import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";
import { DirektivPagesType } from "../schema";
import { DndContext } from "~/design/DragAndDropEditor/Context.tsx";
import { PropsWithChildren } from "react";

type DndContextProviderProps = PropsWithChildren & {
  page: DirektivPagesType;
  setPage: (page: DirektivPagesType) => void;
};

export const DndContextProvider = ({
  page,
  setPage,
  children,
}: DndContextProviderProps) => {
  const onMove = (
    origin: BlockPathType | null,
    target: BlockPathType,
    block: AllBlocksType
  ) => {
    const updatedPage =
      origin === null
        ? addBlockToPage(page, target, block)
        : moveBlockWithinPage(page, origin, target, block);

    setPage(updatedPage);
  };

  return <DndContext onMove={onMove}>{children} </DndContext>;
};
