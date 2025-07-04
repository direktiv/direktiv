import { DndContext as DndKitContext, DragEndEvent } from "@dnd-kit/core";
import { FC, PropsWithChildren } from "react";

import { AllBlocksType } from "~/pages/namespace/Explorer/Page/poc/schema/blocks";
import { BlockPathType } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/Block";
import { idToPath } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/context/utils";

type DndContextProps = PropsWithChildren & {
  onMove: (
    draggableName: BlockPathType,
    droppableName: BlockPathType,
    block: AllBlocksType
  ) => void;
};

export const DndContext: FC<DndContextProps> = ({ children, onMove }) => {
  const onDragEnd = (e: DragEndEvent) => {
    const draggableName = idToPath(String(e.active.id));
    const overId = String(e.over?.id);

    if (draggableName && overId) {
      const droppableName = idToPath(overId);

      const block = e.active.data.current as AllBlocksType;

      if (draggableName && droppableName && block) {
        onMove(draggableName, droppableName, block);
      }
    }
  };

  return <DndKitContext onDragEnd={onDragEnd}>{children}</DndKitContext>;
};
