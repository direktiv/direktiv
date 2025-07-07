import { FC, PropsWithChildren } from "react";

import { twMergeClsx } from "~/util/helpers";

export const activeClassNames = "bg-primary-50 dark:bg-primary-700";
export const inactiveClassNames = "hover:bg-gray-2 dark:hover:bg-gray-dark-2";
export const baseClassNames =
  "[&>svg]:group group flex items-center rounded-md p-2 text-sm font-medium [&>svg]:mr-3";

const createClassNames = (active: boolean, className?: string) =>
  twMergeClsx(
    active ? activeClassNames : inactiveClassNames,
    baseClassNames,
    className
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
