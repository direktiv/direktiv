import { FC, PropsWithChildren } from "react";
import { HoverContainer, HoverElement } from "~/design/HoverContainer";

import Badge from "~/design/Badge";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { Settings } from "lucide-react";
import { twMergeClsx } from "~/util/helpers";
import { useDroppable } from "@dnd-kit/core";

type DroppableProps = PropsWithChildren & {
  position: string;
  onClick: () => void;
};

export const DroppableElement: FC<DroppableProps> = ({
  position,
  children,
  onClick,
}) => {
  const id = position;
  const { setNodeRef, isOver } = useDroppable({
    id,
  });

  return (
    <div ref={setNodeRef} aria-label={id} className="relative">
      {children}
      <Droppable isOver={isOver} id={id} onClick={onClick} />
    </div>
  );
};

export const Droppable = ({
  id,
  isOver,
  onClick,
}: {
  id: string;
  isOver: boolean;
  onClick: () => void;
}) => (
  <HoverContainer>
    <div
      className={twMergeClsx(
        isOver && "border-primary border-4 border-dashed",
        !isOver && "border",
        "border-primary flex h-24 w-full items-center justify-center bg-slate-50"
      )}
    >
      <HoverElement className="bg-white opacity-100" variant="alwaysVisible">
        <Button disabled={!id} icon variant="outline" onClick={onClick}>
          <Settings size={16} />
        </Button>
      </HoverElement>
      <div className="flex flex-col">
        <Badge variant="outline">{id}</Badge>
      </div>
    </div>
  </HoverContainer>
);

type PlaceholderProps = PropsWithChildren & {
  name: string;
  onClick: () => void;
};

export const Placeholder: FC<PlaceholderProps> = ({ name, onClick }) => (
  <div aria-label={name} className="relative">
    <HoverContainer>
      <Card
        className={twMergeClsx(
          "flex h-24 w-full items-center justify-center bg-white"
        )}
      >
        <HoverElement className="bg-white opacity-100" variant="alwaysVisible">
          <Button icon variant="outline" onClick={onClick}>
            <Settings size={16} />
          </Button>
        </HoverElement>
        <div className="flex flex-col">
          <Badge variant="outline" className="h-6">
            {name}
          </Badge>
        </div>
      </Card>
    </HoverContainer>
  </div>
);
