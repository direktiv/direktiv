import { FC, PropsWithChildren } from "react";

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

type ATagProps = JSX.IntrinsicElements["a"];
type NavigationLinkProps = PropsWithChildren<
  ATagProps & {
    active?: boolean;
  }
>;

export const NavigationLink: FC<NavigationLinkProps> = ({
  children,
  active,
  className,
  ...props
}) => (
  <a className={createClassNames(active ?? false, className)} {...props}>
    {children}
  </a>
);
