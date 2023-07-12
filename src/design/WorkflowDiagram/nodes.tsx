import { ComponentProps, FC, PropsWithChildren } from "react";
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

type StartEndHandleProps = PropsWithChildren & { end?: boolean };

const StartEndHandle: FC<StartEndHandleProps> = ({ children, end }) => (
  <Card
    className={twMergeClsx(
      "h-12 w-12 rounded-full p-2 ring-gray-8 dark:ring-gray-dark-8"
    )}
    background="weight-1"
  >
    <div
      className={twMergeClsx(
        "h-full w-full rounded-full",
        end
          ? "bg-success-9 dark:bg-success-dark-9"
          : "bg-gray-9 dark:bg-gray-dark-9"
      )}
    >
      {children}
    </div>
  </Card>
);

export function Start() {
  return (
    <StartEndHandle>
      <CustomHandle type="source" position={Position.Right} />
    </StartEndHandle>
  );
}

export function End() {
  return (
    <StartEndHandle end>
      <CustomHandle type="target" position={Position.Left} />
    </StartEndHandle>
  );
}
