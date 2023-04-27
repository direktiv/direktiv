import { FC, HTMLAttributes } from "react";

import clsx from "clsx";

export const BreadcrumbRoot: FC<HTMLAttributes<HTMLDivElement>> = ({
  children,
  className,
  ...props
}) => (
  <div
    className={clsx(
      "breadcrumbs cursor-pointer py-2 text-sm",

      className
    )}
    {...props}
  >
    <ul className={clsx("flex flex-row flex-wrap items-center")}>{children}</ul>
  </div>
);

export const Breadcrumb: FC<HTMLAttributes<HTMLLIElement>> = ({
  children,
  className,
  ...props
}) => (
  <li
    className={clsx(
      "inline [&>*>svg]:h-4 [&>*>svg]:w-auto [&>*>a>svg]:h-4 [&>*>a>svg]:w-auto",
      "[&>*>a]:flex [&>*>a]:items-center",
      "[&>*]:before:h-2 [&>*]:before:w-2 [&>*]:before:rotate-45",
      "[&>*]:before:border-t [&>*]:before:border-r",
      "[&>*]:before:ml-2 [&>*]:before:mr-3",
      "[&>*]:before:border-gray-11 ",
      "dark:[&>*]:before:border-gray-dark-11 ",
      "[&>*]:first:before:hidden",
      className
    )}
    {...props}
  >
    <div
      className={clsx(
        "flex flex-row items-center gap-2",
        "focus:outline-none focus-visible:outline-offset-2",
        "hover:underline hover:underline-offset-1",
        ""
      )}
    >
      {children}
    </div>
  </li>
);
