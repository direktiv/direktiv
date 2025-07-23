import { CSSProperties, FC, PropsWithChildren } from "react";
import { GripVertical, LucideIcon } from "lucide-react";

import { Card } from "../Card";
import { DragPayloadSchemaType } from "./schema";
import { twMergeClsx } from "~/util/helpers";
import { useDraggable } from "@dnd-kit/core";

type DraggableProps = PropsWithChildren & {
  payload: DragPayloadSchemaType;
  className?: string;
};

const useSharedDragable = (payload: DragPayloadSchemaType) => {
  const { attributes, listeners, setNodeRef, transform } = useDraggable({
    id: JSON.stringify(payload),
    data: payload,
  });

  const styles: CSSProperties = transform
    ? {
        transform: `translate3d(${transform.x}px, ${transform.y}px, 0) scale(1.05)`,
        zIndex: "51",
      }
    : {};

  return { attributes, listeners, setNodeRef, styles };
};

export const SortableItem: FC<DraggableProps> = ({
  payload,
  className,
  children,
}) => {
  const { attributes, listeners, setNodeRef, styles } =
    useSharedDragable(payload);

  return (
    <div style={styles} className="relative">
      <div
        ref={setNodeRef}
        {...listeners}
        {...attributes}
        className={twMergeClsx(
          "absolute right-0 z-10 mt-2 h-[calc(100%-1rem)] text-gray-8 opacity-70 dark:text-gray-dark-8",
          className
        )}
      >
        <div className="flex h-full w-5 items-center justify-center rounded border-2 border-gray-4 bg-white p-0 hover:cursor-move hover:border-solid hover:bg-gray-2 active:cursor-move active:border-solid active:bg-gray-2 dark:border-gray-dark-4 dark:bg-black dark:hover:bg-gray-dark-2">
          <GripVertical />
        </div>
      </div>
      <div className="flex justify-center">
        <div className="w-full">{children}</div>
      </div>
    </div>
  );
};

export const DragablePaletteItem: FC<DraggableProps & { icon: LucideIcon }> = ({
  payload,
  icon: Icon,
  children,
}) => {
  const { attributes, listeners, setNodeRef, styles } =
    useSharedDragable(payload);

  return (
    <div ref={setNodeRef} {...listeners} {...attributes} style={styles}>
      <Card className="z-50 m-4 flex items-center justify-center bg-gray-2 p-4 text-sm text-black hover:cursor-move active:cursor-move dark:bg-gray-dark-2">
        <Icon size={16} className="mr-4" /> {children}
      </Card>
    </div>
  );
};
