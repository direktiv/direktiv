import { FC, HTMLAttributeAnchorTarget, PropsWithChildren } from "react";

import clsx from "clsx";
import { twMerge } from "tailwind-merge";

export const createClassNames = (active: boolean, className?: string) =>
  twMerge(
    clsx(
      active
        ? "bg-primary-50 text-gray-12 dark:bg-primary-700 dark:text-gray-dark-12"
        : "text-gray-11 hover:bg-gray-2 dark:text-gray-dark-11 dark:hover:bg-gray-dark-2",
      "[&>svg]:group group flex items-center rounded-md p-2 text-sm font-medium [&>svg]:mr-3",
      className
    )
  );

export const NavigationLink: FC<
  PropsWithChildren<{
    href: string;
    active?: boolean;
    className?: string;
    target?: HTMLAttributeAnchorTarget;
  }>
> = ({ children, href, active, className, target }) => (
  <a
    href={href}
    className={createClassNames(active ?? false, className)}
    target={target}
    rel={target === "_blank" ? "noopener noreferrer" : ""}
  >
    {children}
  </a>
);
