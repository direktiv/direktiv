import {
  AllBlocks,
  AllBlocksType,
} from "~/pages/namespace/Explorer/Page/poc/schema/blocks";
import { DndContext as DndKitContext, DragEndEvent } from "@dnd-kit/core";
import { FC, PropsWithChildren } from "react";

import { BlockPathType } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/Block";
import { idToPath } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/context/utils";

type DndContextProps = PropsWithChildren & {
  onMove: (
    origin: BlockPathType,
    target: BlockPathType,
    block: AllBlocksType
  ) => void;
};

export const DndContext: FC<DndContextProps> = ({ children, onMove }) => {
  const onDragEnd = (e: DragEndEvent) => {
    if (e.active.id !== null && e.over) {
      const activeId = String(e.active.id);
      const overId = String(e.over.id);

      const origin = idToPath(activeId);
      const target = idToPath(overId);

      if (origin.length === 0) return;
      if (target.length === 0) return;

      const data = e.active.data.current;
      const parsed = AllBlocks.safeParse(data);
      if (!parsed.success) return;

      const block = parsed.data;

      onMove(origin, target, block);
    }
  };

  return <DndKitContext onDragEnd={onDragEnd}>{children}</DndKitContext>;
};
