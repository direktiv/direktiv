import { DndContext as DndKitContext, DragEndEvent } from "@dnd-kit/core";
import { FC, PropsWithChildren } from "react";

type DndContextProps = PropsWithChildren & {
  onMove: (draggableName: string, droppableName: string) => void;
};

export const DndContext: FC<DndContextProps> = ({ children, onMove }) => {
  const onDragEnd = (e: DragEndEvent) => {
    const draggableName = e.active.id ? e.active.id.toString() : undefined;
    const droppableName =
      e.over?.id !== null ? e.over?.id.toString() : undefined;

    if (draggableName && droppableName) onMove?.(draggableName, droppableName);
  };
  return <DndKitContext onDragEnd={onDragEnd}>{children}</DndKitContext>;
};
