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
  narrow,
  className,
  ...props
}) => (
  <div
    className={twMergeClsx(
      narrow && "min-w-[300px] max-w-[300px]",
      !wrap && "h-9",
      "rounded-md border bg-transparent px-3 py-2 text-sm",
      "border-gray-4 placeholder:text-gray-8",
      className
    )}
    {...props}
  >
    {children}
  </div>
);
