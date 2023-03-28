import { FC, HTMLAttributes } from "react";

import clsx from "clsx";

export const BreadcrumbRoot: FC<HTMLAttributes<HTMLDivElement>> = ({
  children,
  className,
  ...props
}) => (
  <div className={clsx("breadcrumbs text-sm", className)} {...props}>
    <ul>{children}</ul>
  </div>
);

export const Breadcrumb: FC<HTMLAttributes<HTMLLIElement>> = ({
  children,
  className,
  ...props
}) => (
  <li
    className={clsx(
      "[&>*>svg]:h-4 [&>*>svg]:w-auto [&>svg]:h-4 [&>svg]:w-auto",
      className
    )}
    {...props}
  >
    {children}
  </li>
);
