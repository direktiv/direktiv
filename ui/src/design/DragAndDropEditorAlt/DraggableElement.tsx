import {
  CSSProperties,
  FC,
  HTMLAttributes,
  PropsWithChildren,
  forwardRef,
} from "react";
import { GripVertical, LucideIcon } from "lucide-react";

import { Card } from "../Card";
import { PayloadSchemaType } from "./schema";
import { useDraggable } from "@dnd-kit/core";

type DraggableProps = PropsWithChildren & {
  payload: PayloadSchemaType;
};

type Transform = ReturnType<typeof useDraggable>["transform"];

type DraggableElementUnstyledProps = PropsWithChildren &
  HTMLAttributes<HTMLDivElement> & {
    payload: PayloadSchemaType;
    style: (transform: Transform) => CSSProperties;
  };

const DraggableElementUnstyled = forwardRef<
  HTMLDivElement,
  DraggableElementUnstyledProps
>(({ payload, style, children, ...props }, ref) => {
  const { attributes, listeners, setNodeRef, transform } = useDraggable({
    // TODO: use a better id?
    id: JSON.stringify(payload),
    data: payload,
  });

  return (
    <div
      ref={(node) => {
        setNodeRef(node);
        if (typeof ref === "function") {
          ref(node);
        } else if (ref) {
          ref.current = node;
        }
      }}
      {...listeners}
      {...attributes}
      {...props}
      style={style(transform)}
    >
      {children}
    </div>
  );
});

DraggableElementUnstyled.displayName = "DraggableElementUnstyled";

export const DraggableElementSort: FC<DraggableProps> = ({
  payload,
  children,
}) => (
  <DraggableElementUnstyled
    payload={payload}
    className="relative m-1"
    style={(transform) =>
      transform
        ? {
            transform: `translate3d(${transform.x}px, ${transform.y}px, 0) scale(1.05)`,
            zIndex: 20,
          }
        : {}
    }
  >
    <div className="absolute z-20 h-full text-gray-8 dark:text-gray-dark-8">
      <div className="flex h-full w-5 items-center justify-center rounded rounded-e-none border-2 border-r-0 border-gray-4 bg-white p-0 hover:cursor-move hover:border-solid hover:bg-gray-2 active:cursor-move active:border-solid active:bg-gray-2 dark:border-gray-dark-4 dark:bg-black dark:hover:bg-gray-dark-2">
        <GripVertical />
      </div>
    </div>
    <div className="flex justify-center">
      <span className="mr-5"></span>
      <div className="w-full">{children}</div>
    </div>
  </DraggableElementUnstyled>
);

export const DraggableElementAdd: FC<DraggableProps & { icon: LucideIcon }> = ({
  payload,
  icon: Icon,
  children,
}) => (
  <DraggableElementUnstyled
    payload={payload}
    className="relative m-1"
    style={(transform) =>
      transform
        ? {
            transform: `translate3d(${transform.x}px, ${transform.y}px, 0) scale(1.05)`,
            zIndex: 50,
          }
        : {}
    }
  >
    <Card className="z-50 m-4 flex items-center justify-center bg-gray-2 p-4 text-sm text-black dark:bg-gray-dark-2">
      <Icon size={16} className="mr-4" /> {children}
    </Card>
  </DraggableElementUnstyled>
);
