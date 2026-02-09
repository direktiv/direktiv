import { ComponentProps, FC, PropsWithChildren } from "react";
import { Handle, Position } from "reactflow";

import { Card } from "../Card";
import { Orientation } from "./types";
import { Separator } from "../Separator";
import { twMergeClsx } from "~/util/helpers";

export type DiagramElementStatus = "pending" | "complete" | "failed";

type StateProps = {
  data: {
    label: string;
    type: string;
    status: DiagramElementStatus;
    orientation: Orientation;
  };
};

type StartEndProps = {
  data: {
    status: DiagramElementStatus;
    orientation: Orientation;
  };
};

type HandleProps = ComponentProps<typeof Handle> & {
  status?: DiagramElementStatus;
};

const CustomHandle: FC<HandleProps> = ({
  type,
  position,
  status = "pending",
}) => (
  <Handle
    type={type}
    position={position}
    id="default"
    className={twMergeClsx(
      "size-2 rounded border",
      "!bg-white dark:!bg-black",
      status === "complete" && "!border-success-9 dark:!border-success-dark-9",
      status === "failed" && "!border-danger-9 dark:!border-danger-dark-9",
      status === "pending" && "border-gray-8 dark:!border-gray-dark-8"
    )}
  />
);

export const State: FC<StateProps> = ({ data }) => {
  const { label, type, status, orientation } = data;
  return (
    <Card
      className={twMergeClsx(
        "flex flex-col",
        status === "complete" && "ring-success-9 dark:ring-success-dark-9",
        status === "failed" && "ring-danger-9 dark:ring-danger-dark-9",
        status === "pending" && "ring-gray-8 dark:ring-gray-dark-8"
      )}
      background="weight-1"
    >
      <CustomHandle
        type="target"
        position={orientation === "horizontal" ? Position.Left : Position.Top}
        status={status}
      />
      <div
        className={twMergeClsx(
          "p-1 text-xs font-bold",
          status === "complete" && "text-success-9 dark:text-success-dark-9",
          status === "failed" && "text-danger-9 dark:text-danger-dark-9"
        )}
      >
        {type}
      </div>
      <Separator
        className={twMergeClsx(
          status === "complete" && "bg-success-9 dark:bg-success-dark-9",
          status === "failed" && "bg-danger-9 dark:bg-danger-dark-9",
          status === "pending" && "bg-gray-8 dark:bg-gray-dark-8"
        )}
      />
      <div
        className={twMergeClsx(
          "p-1 text-xs",
          status === "complete" && "text-success-9 dark:text-success-dark-9",
          status === "failed" && "text-danger-9 dark:text-danger-dark-9"
        )}
      >
        {label}
      </div>
      <CustomHandle
        type="source"
        position={
          orientation === "horizontal" ? Position.Right : Position.Bottom
        }
        status={status}
      />
    </Card>
  );
};

type StartEndHandleProps = PropsWithChildren & {
  end?: boolean;
  status: DiagramElementStatus;
};

const StartEndHandle: FC<StartEndHandleProps> = ({
  children,
  end = false,
  status = "pending",
}) => (
  <Card
    className={twMergeClsx(
      "size-12 rounded-full p-2",
      status === "complete" && "ring-success-9 dark:ring-success-dark-9",
      status === "failed" && "ring-danger-9 dark:ring-danger-dark-9",
      status === "pending" && "ring-gray-8 dark:ring-gray-dark-8"
    )}
    background="weight-1"
  >
    <div
      className={twMergeClsx(
        "size-full rounded-full",
        end && "bg-gray-8 dark:bg-gray-dark-8",
        !end && [
          "ring-1",
          status === "complete" && "ring-success-9 dark:ring-success-dark-9",
          status === "failed" && "ring-danger-9 dark:ring-danger-dark-9",
          status === "pending" && "ring-gray-8 dark:ring-gray-dark-8",
        ]
      )}
    >
      {children}
    </div>
  </Card>
);

export const Start: FC<StartEndProps> = ({ data }) => (
  <StartEndHandle status={data.status}>
    <CustomHandle
      type="source"
      position={
        data.orientation === "horizontal" ? Position.Right : Position.Bottom
      }
      status={data.status}
    />
  </StartEndHandle>
);

export const End: FC<StartEndProps> = ({ data }) => (
  <StartEndHandle status={data.status} end>
    <CustomHandle
      type="target"
      position={
        data.orientation === "horizontal" ? Position.Left : Position.Top
      }
      status={data.status}
    />
  </StartEndHandle>
);
