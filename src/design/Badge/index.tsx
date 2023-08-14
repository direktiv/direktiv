import { Check, Loader2, X } from "lucide-react";
import React, { FC, PropsWithChildren } from "react";

import { twMergeClsx } from "~/util/helpers";

export type BadgeProps = React.HTMLAttributes<HTMLDivElement> &
  PropsWithChildren & {
    variant?: "secondary" | "destructive" | "outline" | "success";
    icon?: "pending" | "complete" | "failed" | "crashed";
  };

const Badge: FC<BadgeProps> = ({
  className,
  variant,
  icon,
  children,
  ...props
}) => (
  <div
    className={twMergeClsx(
      "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs transition-colors focus:outline-none ",
      !variant && "border-transparent bg-gray-12 text-gray-1",
      !variant && "dark:bg-gray-dark-12 dark:text-gray-dark-1",
      variant === "secondary" && "border-transparent bg-gray-5 text-gray-12",
      variant === "secondary" && "dark:bg-gray-dark-5 dark:text-gray-dark-12",
      variant === "destructive" &&
        "border-transparent bg-danger-10 text-gray-1",
      variant === "destructive" &&
        "dark:bg-danger-dark-10 dark:text-gray-dark-1",
      variant === "success" && "border-transparent bg-success-10 text-gray-1",
      variant === "success" && "dark:bg-success-dark-10 dark:text-gray-dark-1",
      variant === "outline" && "border text-gray-12 dark:text-gray-dark-12",
      className
    )}
    {...props}
  >
    {children}
    {icon === "pending" && <Loader2 className="h-3 animate-spin" />}
    {icon === "complete" && <Check className="h-3" />}
    {(icon === "failed" || icon === "crashed") && <X className="h-3" />}
  </div>
);

Badge.displayName = "Badge";

export default Badge;
