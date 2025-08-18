import { DndContext as DndKitContext, DragEndEvent } from "@dnd-kit/core";
import {
  DragAndDropPayloadSchemaType,
  DragPayloadSchema,
  DropPayloadSchema,
} from "./schema";
import { FC, PropsWithChildren } from "react";

type DndContextProps = PropsWithChildren & {
  onDrop: (payload: DragAndDropPayloadSchemaType) => void;
  onDrag?: () => void;
};

export const DndContext: FC<DndContextProps> = ({
  children,
  onDrag,
  onDrop,
}) => {
  const onDragEnd = (e: DragEndEvent) => {
    const parsedDragPayload = DragPayloadSchema.safeParse(
      e.active.data.current
    );
    if (!parsedDragPayload.success) return;

    const parsedDropPayload = DropPayloadSchema.safeParse(e.over?.data.current);
    if (!parsedDropPayload.success) return;

    onDrop({
      drag: parsedDragPayload.data,
      drop: parsedDropPayload.data,
    });
  };

  return (
    <DndKitContext onDragEnd={onDragEnd} onDragStart={onDrag}>
      {children}
    </DndKitContext>
  );
};
