import { FC, PropsWithChildren } from "react";

import clsx from "clsx";

export const NavigationLink: FC<
  PropsWithChildren<{ href: string; active?: boolean }>
> = ({ children, href, active }) => (
  <a
    href={href}
    className={clsx(
      active
        ? "bg-primary50 text-gray-gray12 dark:bg-primary700 dark:text-grayDark-gray12"
        : "text-gray-gray11 hover:bg-gray-gray2 dark:text-grayDark-gray11 dark:hover:bg-grayDark-gray2",
      "[&>svg]:group group flex items-center rounded-md p-2 text-sm font-medium [&>svg]:mr-3"
    )}
  >
    {children}
  </a>
);
