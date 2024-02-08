import { FC, PropsWithChildren } from "react";

import { Card } from "~/design/Card";

type PluginSelectorProps = PropsWithChildren & {
  title: string;
};

export const PluginSelector: FC<PluginSelectorProps> = ({
  title,
  children,
}) => (
  <fieldset className="flex items-center gap-5">
    <label className="text-sm">{title}</label>
    {children}
  </fieldset>
);

export const PluginWrapper: FC<PropsWithChildren> = ({ children }) => (
  <Card className="flex flex-col gap-5 p-5" noShadow>
    {children}
  </Card>
);
