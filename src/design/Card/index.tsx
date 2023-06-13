import React from "react";
import clsx from "clsx";
import { twMerge } from "tailwind-merge";

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  weight?: number;
  noShadow?: boolean;
}
export const Card: React.FC<CardProps> = React.forwardRef<
  HTMLDivElement,
  CardProps
>(({ children, className, weight = 0, noShadow, ...props }, ref) => (
  <div
    ref={ref}
    {...props}
    className={twMerge(
      clsx(
        "rounded-md ring-1",
        "ring-gray-5",
        "dark:ring-gray-dark-5",
        !noShadow && "shadow",
        weight === 0 && "bg-transparent",
        weight === 1 && "bg-white dark:bg-black",
        weight === 2 && "bg-gray-1 dark:bg-gray-dark-1",
        className
      )
    )}
  >
    {children}
  </div>
));

Card.displayName = "Card";
