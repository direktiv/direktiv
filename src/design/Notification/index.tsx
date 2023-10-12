import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import { Bell } from "lucide-react";
import Button from "~/design/Button";
import { twMergeClsx } from "~/util/helpers";

const Notification = ({
  className,
  showIndicator,
  isLoading,
}: {
  className?: string;
  showIndicator?: boolean;
  isLoading?: boolean;
}): JSX.Element => (
  <div className={twMergeClsx("self-end text-right", className)}>
    <Popover>
      <Button variant="ghost" className="items-center px-1" role="button">
        <PopoverTrigger>
          <div className="relative h-6 w-6">
            <Bell className="relative" />
            {showIndicator && (
              <div className="absolute top-0 right-0 rounded-full border-2 border-white bg-danger-10 p-1 dark:border-black dark:bg-danger-dark-10"></div>
            )}
          </div>
        </PopoverTrigger>
      </Button>
      <PopoverContent align="end" className="p-4">
        {isLoading && "loading..."}
        {showIndicator && !isLoading && (
          <div className="absolute top-0 right-0 rounded-full border-2 border-white bg-danger-10 p-1 dark:border-black dark:bg-danger-dark-10">
            <h5>You have unset secrets!</h5>
          </div>
        )}
        {!showIndicator && !isLoading && "Everything is fine."}
      </PopoverContent>
    </Popover>
  </div>
);

export default Notification;
