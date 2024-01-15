import { FC, HTMLAttributes } from "react";

import { RxChevronRight } from "react-icons/rx";
import { ScrollArea } from "../ScrollArea";
import { twMergeClsx } from "~/util/helpers";

export const BreadcrumbRoot: FC<HTMLAttributes<HTMLDivElement>> = ({
  children,
  className,
  ...props
}) => (
  <ScrollArea aria-orientation="horizontal">
    <div
      className={twMergeClsx("cursor-pointer py-4 text-sm", className)}
      {...props}
    >
      <ul className={twMergeClsx("flex flex-row items-center gap-2")}>
        {children}
      </ul>
    </div>
  </ScrollArea>
);

export const Breadcrumb: FC<
  HTMLAttributes<HTMLLIElement> & { noArrow?: boolean }
> = ({ children, className, noArrow, ...props }) => (
  <li
    className={twMergeClsx(
      "flex flex-row items-center gap-2",
      "focus:outline-none focus-visible:outline-offset-2",
      "[&_a]:hover:underline",
      className
    )}
    {...props}
  >
    {noArrow !== true && (
      <RxChevronRight
        aria-hidden
        className="h-4 w-auto fill-current text-gray-8 dark:text-gray-dark-8"
      />
    )}
    <div
      className={twMergeClsx(
        "flex w-max items-center gap-2",
        "[&_a]:flex [&_a]:items-center [&_a]:gap-2",
        "[&_svg]:h-4 [&_svg]:w-auto"
      )}
    >
      {children}
    </div>
  </li>
);
