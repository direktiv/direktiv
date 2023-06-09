import React, { HTMLAttributes } from "react";

import clsx from "clsx";

export const ButtonBar: React.FC<HTMLAttributes<HTMLDivElement>> = ({
  children,
  ...props
}) => (
  <div
    {...props}
    className={clsx(
      "[&_button]:rounded-none",
      "[&_button]:active:scale-100 [&_button]:active:outline-none",
      "[&_button]:border [&_button]:focus:ring-0 [&_button]:focus:ring-offset-0",
      "[&_button]:active:ring-0 [&_button]:active:ring-offset-0",
      "[&>*:first-child]:rounded-l-md",
      "[&>*:last-child]:rounded-r-md"
    )}
  >
    {children}
  </div>
);
ButtonBar.displayName = "ButtonBar";
