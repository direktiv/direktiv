import { FC, PropsWithChildren } from "react";

import Badge from "../Badge";
import { BlockPathType } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/Block";
import { GripVertical } from "lucide-react";
import { twMergeClsx } from "~/util/helpers";

type DraggableProps = PropsWithChildren & {
  blockTypeLabel: string;
  blockPath: BlockPathType;
  className?: string;
  isFocused: boolean;
};

export const DragHandle: FC<DraggableProps> = ({
  blockTypeLabel,
  blockPath,
  isFocused,
}) => (
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
);
