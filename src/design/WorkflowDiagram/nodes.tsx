import { ComponentProps, FC, PropsWithChildren } from "react";
import { Handle, Position } from "reactflow";

import { Card } from "../Card";
import { Separator } from "../Separator";
import { twMergeClsx } from "~/util/helpers";

type StateProps = {
  data: {
    label: string;
    type: string;
    wasExecuted: boolean;
  };
};

type StartEndProps = {
  data: {
    wasExecuted: boolean;
  };
};

type HandleProps = ComponentProps<typeof Handle> & { highlight?: boolean };

const CustomHandle: FC<HandleProps> = ({ type, position, highlight }) => (
  <Handle
    type={type}
    position={position}
    id="default"
    className={twMergeClsx(
      "h-2 w-2 rounded border",
      "!bg-white dark:!bg-black",
      highlight
        ? "!border-success-9 dark:!border-success-dark-9"
        : "!border-gray-8 dark:!border-gray-dark-8"
    )}
  />
);

export const State: FC<StateProps> = ({ data }) => {
  const { label, type, wasExecuted } = data;
  return (
    <Card
      className={twMergeClsx(
        "flex flex-col",
        wasExecuted
          ? "ring-success-9 dark:ring-success-dark-9"
          : "ring-gray-8 dark:ring-gray-dark-8"
      )}
      background="weight-1"
    >
      <CustomHandle
        type="target"
        position={Position.Left}
        highlight={wasExecuted}
      />
      <div
        className={twMergeClsx(
          "p-1 text-xs font-bold",
          wasExecuted && "text-success-9 dark:text-success-dark-9"
        )}
      >
        {type}
      </div>
      <Separator
        className={twMergeClsx(
          wasExecuted
            ? "bg-success-9 dark:bg-success-dark-9"
            : "bg-gray-8 dark:bg-gray-dark-8"
        )}
      />
      <div
        className={twMergeClsx(
          "p-1 text-xs",
          wasExecuted && "text-success-9 dark:text-success-dark-9"
        )}
      >
        {label}
      </div>
      <CustomHandle
        type="source"
        position={Position.Right}
        highlight={wasExecuted}
      />
    </Card>
  );
};

type StartEndHandleProps = PropsWithChildren & {
  end?: boolean;
  highlight?: boolean;
};

const StartEndHandle: FC<StartEndHandleProps> = ({
  children,
  end,
  highlight,
}) => (
  <Card
    className={twMergeClsx(
      "h-12 w-12 rounded-full p-2",
      highlight
        ? "ring-success-9 dark:ring-success-dark-9"
        : "ring-gray-8 dark:ring-gray-dark-8"
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

export const Start: FC<StartEndProps> = ({ data }) => (
  <StartEndHandle highlight={data.wasExecuted}>
    <CustomHandle
      type="source"
      position={Position.Right}
      highlight={data.wasExecuted}
    />
  </StartEndHandle>
);

export const End: FC<StartEndProps> = ({ data }) => (
  <StartEndHandle highlight={data.wasExecuted} end>
    <CustomHandle
      type="target"
      position={Position.Left}
      highlight={data.wasExecuted}
    />
  </StartEndHandle>
);
