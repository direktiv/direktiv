import * as DropdownMenuPrimitive from "@radix-ui/react-dropdown-menu";
import * as React from "react";

import { Bell, Loader2, LucideIcon } from "lucide-react";
import { FC, PropsWithChildren } from "react";
import {
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
} from "~/design/Popover";

import Button from "~/design/Button";
import { twMergeClsx } from "~/util/helpers";

const NotificationClose = PopoverClose;

const NotificationMenuSeparator = React.forwardRef<
  React.ElementRef<typeof DropdownMenuPrimitive.Separator>,
  React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Separator>
>(({ className, ...props }, ref) => (
  <DropdownMenuPrimitive.Separator
    ref={ref}
    className={twMergeClsx(
      "my-1 h-px bg-gray-3 dark:bg-gray-dark-3",
      className
    )}
    {...props}
  />
));

NotificationMenuSeparator.displayName =
  DropdownMenuPrimitive.Separator.displayName;

const NotificationTitle: FC<PropsWithChildren> = ({ children }) => (
  <div className="px-2 py-1.5 text-sm font-semibold text-gray-9 dark:text-gray-dark-9">
    {children}
  </div>
);

const NotificationText: FC<PropsWithChildren> = ({ children }) => (
  <div className="px-2 py-1.5 text-sm font-medium text-gray-11 dark:text-gray-dark-11">
    {children}
  </div>
);

const NotificationLoading: FC<PropsWithChildren> = ({ children }) => (
  <div className="flex flex-col focus:bg-gray-3 dark:focus:bg-gray-dark-3">
    <div className="flex items-center py-1.5 px-2">
      <div className="w-max">
        <Loader2 className="animate-spin text-gray-11 dark:text-gray-dark-11" />
      </div>
      <NotificationText>{children}</NotificationText>
    </div>
  </div>
);

function NotificationMessage({
  text,
  icon: Icon,
}: {
  text: string;
  icon: LucideIcon;
}) {
  return (
    <div className="flex flex-col focus:bg-gray-3 dark:focus:bg-gray-dark-3">
      <div className="flex items-center py-1.5 px-2">
        <div className="w-max">
          <Icon
            className="text-gray-11 dark:text-gray-dark-11"
            aria-hidden="true"
          />
        </div>
        <NotificationText>{text}</NotificationText>
      </div>
    </div>
  );
}

type NotificationPropsType = PropsWithChildren & {
  className?: string;
  showIndicator?: boolean;
};

const Notification: FC<NotificationPropsType> = ({
  className,
  showIndicator,
  children,
}) => (
  <div className={twMergeClsx("", className)}>
    <Popover>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          className="group items-center px-1"
          role="button"
        >
          <div className="relative h-6 w-6">
            <Bell className="relative" />
            {showIndicator && (
              <div className="absolute top-0 right-0 rounded-full border-2 border-white bg-danger-10 p-1 transition-colors group-hover:border-gray-3 dark:border-black dark:bg-danger-dark-10 dark:group-hover:border-gray-dark-3"></div>
            )}
          </div>
        </Button>
      </PopoverTrigger>
      <PopoverContent className="bg-gray-1 dark:bg-gray-dark-1" align="end">
        {children}
      </PopoverContent>
    </Popover>
  </div>
);

export {
  Notification,
  NotificationClose,
  NotificationLoading,
  NotificationMessage,
  NotificationMenuSeparator,
  NotificationTitle,
  NotificationText,
};
