import { PropsWithChildren } from "react";
import { twMergeClsx } from "~/util/helpers";

type PanelContainerProps = PropsWithChildren & {
  className: string;
};

export const PanelContainer = ({
  children,
  className,
}: PanelContainerProps) => (
  <div
    data-testid="editor-dragArea"
    className={twMergeClsx(
      "h-[300px] border-b-2 border-gray-4 dark:border-gray-dark-4 sm:h-[calc(100vh-230px)] sm:border-b-0 sm:border-r-2",
      className
    )}
  >
    {children}
  </div>
);
