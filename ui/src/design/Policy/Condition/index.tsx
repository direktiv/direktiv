import type { HTMLAttributes, PropsWithChildren } from "react";

import { Card } from "~/design/Card";
import { twMergeClsx } from "~/util/helpers";

type ConditionComponentProps = PropsWithChildren &
  HTMLAttributes<HTMLDivElement>;

const Condition = ({
  children,
  className,
  ...props
}: ConditionComponentProps) => (
  <Card
    background="weight-2"
    className={twMergeClsx(
      "flex h-[64px] w-40 flex-col items-center justify-center overflow-hidden p-2 text-center text-xs",
      className
    )}
    aria-label="condition"
    {...props}
  >
    {children}
  </Card>
);

export { Condition };
