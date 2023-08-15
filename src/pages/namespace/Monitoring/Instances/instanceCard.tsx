import { FC, PropsWithChildren, ReactNode } from "react";

import { Card } from "~/design/Card";
import { LucideIcon } from "lucide-react";

type InstanceCardProps = PropsWithChildren & {
  headline: string;
  refetchButton: ReactNode;
  icon: LucideIcon;
};

export const InstanceCard: FC<InstanceCardProps> = ({
  children,
  headline,
  refetchButton,
  icon: Icon,
}) => (
  <Card className="flex flex-col">
    <div className="flex items-center gap-x-2 border-b border-gray-5 p-5 font-medium dark:border-gray-dark-5">
      <Icon className="h-5" />
      <h3 className="grow">{headline}</h3>
      {refetchButton}
    </div>
    {children}
  </Card>
);
