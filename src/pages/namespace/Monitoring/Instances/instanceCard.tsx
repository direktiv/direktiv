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
  <Card className="flex flex-col p-5">
    <h3 className="flex items-center gap-x-2 font-medium">
      <Icon className="h-5" />
      {headline}
    </h3>
    <div className="grow">{children}</div>
  </Card>
);
