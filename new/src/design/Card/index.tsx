import React from "react";
import { twMergeClsx } from "~/util/helpers";

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  background?: "none" | "weight-1" | "weight-2";
  noShadow?: boolean;
}
export const Card: React.FC<CardProps> = React.forwardRef<
  HTMLDivElement,
  CardProps
>(({ children, className, background = "none", noShadow, ...props }, ref) => (
  <div
    ref={ref}
    {...props}
    className={twMergeClsx(
      "rounded-md ring-1",
      "ring-gray-5",
      "dark:ring-gray-dark-5",
      !noShadow && "shadow",
      background === "weight-1" && "bg-white dark:bg-black",
      background === "weight-2" && "bg-gray-1 dark:bg-gray-dark-1",
      className
    )}
  >
    {children}
  </div>
));

Card.displayName = "Card";
