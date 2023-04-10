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
  <div className={clsx("mt-8 flow-root")} ref={ref} {...props}>
    <div className="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
      <div className="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
        <div
          className={clsx(
            "overflow-hidden ring-1 sm:rounded-lg",
            " bg-white ring-black/5 ",
            " dark:bg-black dark:ring-white/5",
            noShadow ? "shadow-none" : "shadow",
            withBackground && "bg-gray-2 dark:bg-gray-dark-2",
            className
          )}
        >
          {children}
        </div>
      </div>
    </div>
  </div>
));

Card.displayName = "Card";
