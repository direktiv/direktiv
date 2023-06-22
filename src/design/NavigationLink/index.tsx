import { FC, PropsWithChildren } from "react";

import { twMergeClsx } from "~/util/helpers";

export const createClassNames = (active: boolean) =>
  twMergeClsx(
    active
      ? "bg-primary-50 text-gray-12 dark:bg-primary-700 dark:text-gray-dark-12"
      : "text-gray-11 hover:bg-gray-2 dark:text-gray-dark-11 dark:hover:bg-gray-dark-2",
    "[&>svg]:group group flex items-center rounded-md p-2 text-sm font-medium [&>svg]:mr-3"
  );

export const NavigationLink: FC<
  PropsWithChildren<{ href: string; active?: boolean }>
> = ({ children, href, active }) => (
  <a href={href} className={createClassNames(active ?? false)}>
    {children}
  </a>
);
