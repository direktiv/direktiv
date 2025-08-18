import { CSSProperties, FC, PropsWithChildren } from "react";

import { BlockPathType } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/Block";
import { Card } from "../Card";
import { DragHandle } from "./DragHandle";
import { DragPayloadSchemaType } from "./schema";
import { LucideIcon } from "lucide-react";
import { twMergeClsx } from "~/util/helpers";
import { useDraggable } from "@dnd-kit/core";

type DraggableProps = PropsWithChildren & {
  payload: DragPayloadSchemaType;
  blockTypeLabel: string;
  blockPath: BlockPathType;
  className?: string;
  isFocused: boolean;
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
        opacity: 0.7,
      }
    : {};

  return { attributes, listeners, setNodeRef, styles };
};

export const SortableItem: FC<DraggableProps> = ({
  payload,
  blockTypeLabel,
  blockPath,
  isFocused,
  className,
  children,
}) => {
  const { attributes, listeners, setNodeRef, styles } =
    useSharedDraggable(payload);

  return (
    <div style={styles} className="relative">
      <div
        ref={setNodeRef}
        {...listeners}
        {...attributes}
        className={twMergeClsx(
          "pointer-events-none absolute z-40 mt-3 h-[calc(100%-1rem)] w-full",
          className
        )}
      >
        <DragHandle
          isFocused={isFocused}
          blockTypeLabel={blockTypeLabel}
          blockPath={blockPath}
        />
      </div>
      <div className="flex justify-center">
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
      <Card className="z-50 flex items-center justify-center bg-gray-2 p-2 text-sm hover:cursor-move active:cursor-move dark:bg-gray-dark-2">
        <Icon size={16} className="mr-2" /> {children}
      </Card>
    </div>
  );
};
