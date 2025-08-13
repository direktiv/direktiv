import { CSSProperties, FC, PropsWithChildren } from "react";
import { GripVertical, LucideIcon } from "lucide-react";

import Badge from "../Badge";
import { Card } from "../Card";
import { DragPayloadSchemaType } from "./schema";
import { pathsEqual } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/context/utils";
import { twMergeClsx } from "~/util/helpers";
import { useBlockTypes } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/context/utils/useBlockTypes";
import { useDraggable } from "@dnd-kit/core";
import { usePageEditorPanel } from "~/pages/namespace/Explorer/Page/poc/BlockEditor/EditorPanelProvider";

type DraggableProps = PropsWithChildren & {
  payload: DragPayloadSchemaType;
  className?: string;
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
  className,
  children,
}) => {
  const { attributes, listeners, setNodeRef, styles } =
    useSharedDraggable(payload);

  const { blockTypes } = useBlockTypes();
  const { panel } = usePageEditorPanel();

  if (payload.type !== "move") return null;

  const isFocused = panel?.action && pathsEqual(panel.path, payload.originPath);

  const findType = blockTypes.find((type) => type.type === payload.block.type);

  const blockTypeLabel = findType ? findType.label : "not found";

  const blockPath = payload.originPath;

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
        <Badge
          className={twMergeClsx(
            "pointer-events-auto absolute -mt-7 text-nowrap rounded-md rounded-b-none px-2 py-1 hover:cursor-move active:cursor-move",
            isFocused && "bg-gray-8 dark:bg-gray-dark-8"
          )}
          variant="secondary"
        >
          <GripVertical
            size={16}
            className={twMergeClsx(
              "mr-2 text-gray-8",
              isFocused && "text-black dark:text-white"
            )}
          />
          <span className="mr-2">
            <b>{blockTypeLabel}</b>
          </span>
          {blockPath.join(".")}
        </Badge>
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
