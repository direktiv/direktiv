import { FC, PropsWithChildren } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import { Bell } from "lucide-react";
import Button from "~/design/Button";
import { twMergeClsx } from "~/util/helpers";

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
      <PopoverContent
        align="end"
        className=" bg-gray-1 p-4 px-1.5 text-[10px] font-medium opacity-100 dark:border-gray-dark-4 dark:bg-gray-dark-1"
      >
        {children}
      </PopoverContent>
    </Popover>
  </div>
);

export default Notification;
