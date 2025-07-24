import { CSSProperties, FC, PropsWithChildren } from "react";
import { GripVertical, LucideIcon } from "lucide-react";

import { Card } from "../Card";
import { DragPayloadSchemaType } from "./schema";
import { useDraggable } from "@dnd-kit/core";

type DraggableProps = PropsWithChildren & {
  payload: DragPayloadSchemaType;
};

const useSharedDraggable = (payload: DragPayloadSchemaType) => {
  const { attributes, listeners, setNodeRef, transform } = useDraggable({
    id: JSON.stringify(payload),
    data: payload,
  });

  const styles: CSSProperties = transform
    ? {
        transform: `translate3d(${transform.x}px, ${transform.y}px, 0) scale(1.05)`,
        position: "relative",
        zIndex: "50",
      }
    : {};

  return { attributes, listeners, setNodeRef, styles };
};

export const SortableItem: FC<DraggableProps> = ({ payload, children }) => {
  const { attributes, listeners, setNodeRef, styles } =
    useSharedDraggable(payload);

  return (
    <div style={styles} className="relative m-1">
      <div
        ref={setNodeRef}
        {...listeners}
        {...attributes}
        className="absolute z-40 h-full text-gray-8 dark:text-gray-dark-8"
      >
        <div className="flex h-full w-5 items-center justify-center rounded rounded-e-none border-2 border-r-0 border-gray-4 bg-white p-0 hover:cursor-move hover:border-solid hover:bg-gray-2 active:cursor-move active:border-solid active:bg-gray-2 dark:border-gray-dark-4 dark:bg-black dark:hover:bg-gray-dark-2">
          <GripVertical />
        </div>
      </div>
      <div className="flex justify-center">
        <span className="mr-5"></span>
        <div className="w-full">{children}</div>
      </div>
    </div>
  );
};

export const DraggablePaletteItem: FC<
  DraggableProps & { icon: LucideIcon }
> = ({ payload, icon: Icon, children }) => {
  const { attributes, listeners, setNodeRef, styles } =
    useSharedDraggable(payload);

  return (
    <div ref={setNodeRef} {...listeners} {...attributes} style={styles}>
      <Card className="z-50 flex items-center justify-center bg-gray-2 p-3 text-sm text-black hover:cursor-move active:cursor-move dark:bg-gray-dark-2">
        <Icon size={16} className="mr-2" /> {children}
      </Card>
    </div>
  );
};
