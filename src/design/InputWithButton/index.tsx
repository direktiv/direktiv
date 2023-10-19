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
        "relative flex w-full items-end [&_input]:pr-10",
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
