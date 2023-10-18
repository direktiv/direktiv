import { ComponentProps, FC, PropsWithChildren } from "react";

import Button from "~/design/Button";
import { Loader2 } from "lucide-react";

type ButtonWithChildren = Omit<ComponentProps<typeof Button>, "variant">;

export const NotificationButton: FC<ButtonWithChildren> = (props) => (
  <Button {...props} variant="outline" />
);

export const NotificationTitle: FC<PropsWithChildren> = ({ children }) => (
  <div className="px-2 py-1.5 text-sm font-semibold text-gray-9 dark:text-gray-dark-9">
    {children}
  </div>
);

export const NotificationText: FC<PropsWithChildren> = ({ children }) => (
  <div className="px-2 py-1.5 text-sm font-medium text-gray-9 dark:text-gray-dark-9">
    {children}
  </div>
);

export const NotificationLoading: FC<PropsWithChildren> = ({ children }) => (
  <div className="flex items-center">
    <Loader2 className="h-5 animate-spin" />
    <div className="px-2 py-1.5 text-sm font-medium text-gray-9 dark:text-gray-dark-9">
      {children}
    </div>
  </div>
);
