import * as React from "react";

import clsx from "clsx";

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "secondary" | "destructive" | "outline";
}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div
      className={clsx(
        "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs transition-colors focus:outline-none ",
        !variant && "border-transparent bg-gray-12 text-gray-1",
        !variant && "dark:bg-gray-dark-12 dark:text-gray-dark-1",
        variant === "secondary" && "border-transparent bg-gray-5 text-gray-12",
        variant === "secondary" && "dark:bg-gray-dark-5 dark:text-gray-dark-12",
        variant === "destructive" &&
          "border-transparent bg-danger-10 text-gray-1",
        variant === "destructive" &&
          "dark:bg-danger-dark-10 dark:text-gray-dark-1",
        variant === "outline" && "border text-gray-12 dark:text-gray-dark-12",
        className
      )}
      {...props}
    />
  );
}
Badge.displayName = "Badge";

export default Badge;
