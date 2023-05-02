import * as React from "react";

import clsx from "clsx";

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "default" | "secondary" | "destructive" | "outline";
}

function Badge({ className, variant = "default", ...props }: BadgeProps) {
  return (
    <div
      className={clsx(
        "focus:ring-ring inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2",
        variant === "default" &&
          "border-transparent bg-gray-12 text-white hover:bg-gray-12/80",
        variant === "default" &&
          "dark:bg-gray-dark-12 dark:text-black dark:hover:bg-gray-dark-12/80",
        variant === "secondary" &&
          "border-transparent bg-gray-8 text-black hover:bg-gray-8/80",
        variant === "secondary" &&
          "dark:bg-gray-dark-8 dark:text-white dark:hover:bg-gray-dark-8/80",
        variant === "destructive" &&
          "border-transparent bg-danger-11 text-white hover:bg-danger-11/80",
        variant === "destructive" &&
          "dark:bg-danger-dark-11 dark:text-black dark:hover:bg-danger-dark-11/80",
        variant === "outline" && "text-black dark:text-white",
        className
      )}
      {...props}
    />
  );
}
Badge.displayName = "Badge";

export default Badge;
