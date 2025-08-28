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
      // Todo:
      // - consolidate this with the input styling?
      // - focus ring does not work as expected
      narrow && "min-w-[300px] max-w-[300px]",
      !wrap && "h-9",
      "rounded-md border bg-transparent px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
      "border-gray-4 placeholder:text-gray-8 focus:ring-gray-4 focus:ring-offset-gray-1",
      "dark:border-gray-dark-4 dark:placeholder:text-gray-dark-8 dark:focus:ring-gray-dark-4",
      "dark:focus:ring-offset-gray-dark-1",
      className
    )}
    {...props}
  >
    {children}
  </div>
);
