import { FC, PropsWithChildren } from "react";

import { LucideIcon } from "lucide-react";

type NoResultProps = PropsWithChildren<{
  button?: JSX.Element;
  icon: LucideIcon;
}>;

const NoResult: FC<NoResultProps> = ({ children, icon: Icon, button }) => (
  <div className="flex flex-col items-center gap-y-5 p-10">
    <div className="flex flex-col items-center justify-center gap-1">
      <Icon />
      <span className="text-center text-sm">{children}</span>
    </div>
    {button && <div className="flex flex-col gap-5 sm:flex-row">{button}</div>}
  </div>
);

export default NoResult;
