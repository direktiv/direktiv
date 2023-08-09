import { FC, PropsWithChildren, SVGProps } from "react";

import { Card } from "~/design/Card";

type InstanceCardProps = PropsWithChildren & {
  headline: string;
  icon: FC<SVGProps<SVGSVGElement>>;
};

export const InstanceCard: FC<InstanceCardProps> = ({
  children,
  headline,
  icon: Icon,
}) => (
  <Card className="flex flex-col">
    <h3 className="flex items-center gap-x-2 border-b border-gray-5 p-5 font-medium dark:border-gray-dark-5">
      <Icon className="h-5" />
      {headline}
    </h3>
    {children}
  </Card>
);
