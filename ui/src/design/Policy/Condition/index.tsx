import type { PropsWithChildren } from "react";

import { Card } from "~/design/Card";

type ConditionComponentProps = PropsWithChildren;

const Condition = ({ children }: ConditionComponentProps) => (
  <Card
    background="weight-2"
    className="flex min-w-40 flex-col items-center justify-center py-2 text-center text-xs"
    aria-label="condition"
  >
    {children}
  </Card>
);

export { Condition };
