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
      "flex min-w-40 max-w-64 flex-col items-center justify-center overflow-hidden p-5 text-center text-xs",
      className
    )}
    aria-label="condition"
    {...props}
  >
    {children}
  </Card>
);

export { Condition };
