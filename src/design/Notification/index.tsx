import { FC, PropsWithChildren } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import { Bell } from "lucide-react";
import Button from "~/design/Button";
import { twMergeClsx } from "~/util/helpers";

type NotificationPropsType = PropsWithChildren & {
  className?: string;
  showIndicator?: boolean;
  isLoading?: boolean;
};

const Notification: FC<NotificationPropsType> = ({
  className,
  showIndicator,
  isLoading,
  children,
}) => (
  <div className={twMergeClsx("self-end text-right", className)}>
    <Popover>
      <Button variant="ghost" className="group items-center px-1" role="button">
        <PopoverTrigger>
          <div className="relative h-6 w-6">
            <Bell className="relative" />
            {!isLoading && showIndicator && (
              <div className="absolute top-0 right-0 rounded-full border-2 border-white bg-danger-10 p-1 group-hover:border-gray-3 dark:border-black dark:bg-danger-dark-10 dark:group-hover:border-gray-dark-3"></div>
            )}
          </div>
        </PopoverTrigger>
      </Button>
      <PopoverContent align="end" className="p-4">
        {children}
        {/* <NotificationModal
          isLoading={isLoading}
          showIndicator={showIndicator}
        ></NotificationModal> */}
      </PopoverContent>
    </Popover>
  </div>
);

export default Notification;
