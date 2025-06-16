import { DndContext as DndKitContext, DragEndEvent } from "@dnd-kit/core";
import { FC, PropsWithChildren } from "react";

type DndContextProps = PropsWithChildren & {
  onMove: (
    draggableName: string,
    droppableName: string,
    position: "before" | "after" | undefined
  ) => void;
};

export const DndContext: FC<DndContextProps> = ({ children, onMove }) => {
  const onDragEnd = (e: DragEndEvent) => {
    const draggableName = e.active.id?.toString();
    const overId = e.over?.id?.toString();

    if (draggableName && overId) {
      const [droppableName, position] = overId.split("-") as [
        string,
        "before" | "after" | undefined,
      ];

      if (
        draggableName &&
        droppableName &&
        (position === "before" ||
          position === "after" ||
          position === undefined)
      ) {
        onMove(draggableName, droppableName, position);
      }
    }
  };

  return <DndKitContext onDragEnd={onDragEnd}>{children}</DndKitContext>;
};
