import { FC, HTMLAttributes } from "react";

import { RxChevronRight } from "react-icons/rx";
import clsx from "clsx";

export const BreadcrumbRoot: FC<HTMLAttributes<HTMLDivElement>> = ({
  children,
  className,
  ...props
}) => (
  <div className={clsx("cursor-pointer py-2 text-sm", className)} {...props}>
    <ul className={clsx("flex flex-row flex-wrap items-center")}>{children}</ul>
  </div>
);

export const Breadcrumb: FC<
  HTMLAttributes<HTMLLIElement> & { noArrow?: boolean }
> = ({ children, className, noArrow, ...props }) => (
  <li
    className={clsx(
      "flex flex-row items-center gap-2",
      "focus:outline-none focus-visible:outline-offset-2",
      "hover:underline",
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
      className={clsx(
        "flex items-center gap-2",
        "[&_a]:flex [&_a]:items-center [&_a]:gap-2",
        "[&_svg]:h-4 [&_svg]:w-auto"
      )}
    >
      {children}
    </div>
  </li>
  // </li>
);
