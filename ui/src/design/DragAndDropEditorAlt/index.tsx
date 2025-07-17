import { DndContext as DndKitContext, DragEndEvent } from "@dnd-kit/core";
import { FC, PropsWithChildren } from "react";
import { PayloadSchema, PayloadSchemaType } from "./schema";

type DndContextProps = PropsWithChildren & {
  onDrop: (payload: PayloadSchemaType) => void;
};

export const DndContext: FC<DndContextProps> = ({ children, onDrop }) => {
  const onDragEnd = (e: DragEndEvent) => {
    //  validate payload and only call onDrop if it is valid
    const parsedPayload = PayloadSchema.safeParse(e.active.data.current);
    if (!parsedPayload.success) return;
    const payload = parsedPayload.data;

    onDrop(payload);
  };

  return <DndKitContext onDragEnd={onDragEnd}>{children}</DndKitContext>;
};
