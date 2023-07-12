import { ComponentProps, FC } from "react";
import { Handle, Position } from "reactflow";

import { Card } from "../Card";
import { Separator } from "../Separator";
import { twMergeClsx } from "~/util/helpers";

type StateProps = {
  data: {
    label: string;
    type: string;
  };
};

type HandleProps = ComponentProps<typeof Handle>;

const CustomHandle: FC<HandleProps> = ({ type, position }) => (
  <Handle
    type={type}
    position={position}
    id="default"
    className={twMergeClsx(
      "h-2 w-2 rounded border",
      "border-gray-8 !bg-white",
      "dark:border-gray-dark-8 dark:!bg-black"
    )}
  />
);

export function State(props: StateProps) {
  const { data } = props;
  const { label, type } = data;
  return (
    <Card
      className="flex flex-col ring-gray-8 dark:ring-gray-dark-8"
      background="weight-1"
    >
      <CustomHandle type="target" position={Position.Left} />
      <div className="p-1 text-xs font-bold">{type}</div>
      <Separator className="bg-gray-8 dark:bg-gray-dark-8" />
      <div className="p-1 text-xs">{label}</div>
      <CustomHandle type="source" position={Position.Right} />
    </Card>
  );
}

export function Start() {
  return (
    <div className="normal">
      <Handle type="source" position={Position.Right} />
      <div className="start" />
    </div>
  );
}

export function End() {
  return (
    <div className="normal">
      <div className="end" />
      <Handle type="target" position={Position.Left} />
    </div>
  );
}
