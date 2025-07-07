import {
  addBlockToPage,
  moveBlockWithinPage,
} from "../PageCompiler/context/utils";
import {
  usePage,
  usePageEditor,
} from "../PageCompiler/context/pageCompilerContext";

import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";
import { DndContext } from "~/design/DragAndDropEditor/Context.tsx";
import { PropsWithChildren } from "react";

export const DndContextProvider = ({ children }: PropsWithChildren) => {
  const onMove = (
    origin: BlockPathType | null,
    target: BlockPathType,
    block: AllBlocksType
  ) => {
    const page = usePage();
    const { setPage } = usePageEditor();

    const updatedPage =
      origin === null
        ? addBlockToPage(page, target, block)
        : moveBlockWithinPage(page, origin, target, block);

    setPage(updatedPage);
  };

  return <DndContext onMove={onMove}>{children} </DndContext>;
};
