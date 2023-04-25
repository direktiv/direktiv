import React from "react";
import clsx from "clsx";

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  withBackground?: boolean;
  noShadow?: boolean;
}
export const Card: React.FC<CardProps> = React.forwardRef<
  HTMLDivElement,
  CardProps
>(({ children, className, withBackground, noShadow, ...props }, ref) => (
  <div
    ref={ref}
    {...props}
    className={clsx(
      "rounded-md ring-1",
      "ring-gray-5",
      "dark:ring-gray-dark-5",
      !noShadow && "shadow",
      withBackground
        ? "bg-gray-1 dark:bg-gray-dark-1"
        : "bg-white dark:bg-black",
      className
    )}
  >
    {children}
  </div>
));

Card.displayName = "Card";
