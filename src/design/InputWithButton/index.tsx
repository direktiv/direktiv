import React, { HTMLAttributes } from "react";

import { twMergeClsx } from "~/util/helpers";

export const InputWithButton: React.FC<HTMLAttributes<HTMLDivElement>> = ({
  children,
  className,
  ...props
}) => {
  const [firstChild, secondChild] = Array.isArray(children)
    ? children
    : [children];
  return (
    <div
      {...props}
      className={twMergeClsx(
        "&_input]:pr-10 relative flex w-full items-end",
        className
      )}
    >
      {firstChild}
      <div
        className={twMergeClsx(
          "absolute right-1 flex h-9 items-center justify-center [&_button]:h-8 [&_button]:w-8"
        )}
      >
        {secondChild}
      </div>
    </div>
  );
};
InputWithButton.displayName = "InputWithButton";

export const InputWithButtonWithState: React.FC<
  HTMLAttributes<HTMLDivElement> & {
    icon?: boolean;
  }
> = ({ children, className, ...props }) => {
  const [firstChild, secondChild] = Array.isArray(children)
    ? children
    : [children];
  return (
    <div
      {...props}
      className={twMergeClsx(
        "relative flex items-end p-1",
        "[&_input]:border-0 [&_input]:outline-none [&_input]:focus:border-0 [&_input]:focus:outline-none",
        className
      )}
    >
      {firstChild}
      <div
        className={twMergeClsx(
          "absolute right-1 flex h-9 items-center justify-center [&_button]:h-8"
        )}
      >
        {secondChild}
      </div>
    </div>
  );
};
InputWithButtonWithState.displayName = "InputWithButtonWithState";
