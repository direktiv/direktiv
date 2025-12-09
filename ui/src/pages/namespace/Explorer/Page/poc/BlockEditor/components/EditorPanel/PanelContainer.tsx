import { PropsWithChildren } from "react";

type PanelContainerProps = PropsWithChildren & {
  className: string;
};

export const PanelContainer = ({
  children,
  className,
}: PanelContainerProps) => (
  <div data-testid="editor-dragArea" className={className}>
    {children}
  </div>
);
