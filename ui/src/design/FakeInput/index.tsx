import { FC, PropsWithChildren } from "react";

import { twMergeClsx } from "~/util/helpers";

type FakeInputProps = PropsWithChildren & {
  wrap?: boolean;
  narrow?: boolean;
  className?: string;
};

export const FakeInput: FC<FakeInputProps> = ({
  children,
  wrap,
  className,
  ...props
}) => (
  <div
    className={twMergeClsx(
      !wrap && "h-9 truncate",
      "rounded-md border bg-transparent px-3 py-2 text-sm",
      "border-gray-4 placeholder:text-gray-8 dark:border-gray-dark-4 dark:placeholder:text-gray-dark-8",
      className
    )}
    {...props}
  >
    {children}
  </div>
);
