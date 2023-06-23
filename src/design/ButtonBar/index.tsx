import React, { HTMLAttributes } from "react";

import { twMergeClsx } from "~/util/helpers";

export const ButtonBar: React.FC<HTMLAttributes<HTMLDivElement>> = ({
  children,
  ...props
}) => (
  <div
    {...props}
    className={twMergeClsx(
      "[&_button]:rounded-none",
      "[&_button]:mr-[-1px]",
      "[&_button]:active:outline-none",
      "[&_button]:border [&_button]:focus:ring-0 [&_button]:focus:ring-offset-0",
      "[&_button]:active:ring-0 [&_button]:active:ring-offset-0",
      "[&>*:first-child]:rounded-l-md",
      "[&>*:last-child]:rounded-r-md",
      "flex items-end"
    )}
  >
    {children}
  </div>
);
ButtonBar.displayName = "ButtonBar";
