import { DndContext as DndKitContext, DragEndEvent } from "@dnd-kit/core";
import { FC, PropsWithChildren } from "react";

import { AllBlocksType } from "~/pages/namespace/Explorer/Page/poc/schema/blocks";

type DndContextProps = PropsWithChildren & {
  onMove: (
    draggableName: string,
    droppableName: string,
    element: AllBlocksType
  ) => void;
};

export const DndContext: FC<DndContextProps> = ({ children, onMove }) => {
  const onDragEnd = (e: DragEndEvent) => {
    const draggableName = e.active.id?.toString();
    const overId = e.over?.id?.toString();

    if (draggableName && overId) {
      const [droppableName] = overId.split("-") as [string];

      const element = e.active.data.current;

      if (draggableName && droppableName && element) {
        onMove(draggableName, droppableName, element);
      }
    }
  };

  return <DndKitContext onDragEnd={onDragEnd}>{children}</DndKitContext>;
};
