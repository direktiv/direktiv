import { Bell, Loader2 } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { twMergeClsx } from "~/util/helpers";

const Notification = ({
  className,
  showIndicator,
  isLoading,
  text,
}: {
  className?: string;
  showIndicator?: boolean;
  isLoading?: boolean;
  text?: string;
}): JSX.Element => (
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
        {isLoading && (
          <div className=" flex">
            <Loader2 className="h-5 animate-spin" />
            {text}
          </div>
        )}

        {showIndicator && !isLoading && (
          <div className="">
            <h5>{text}</h5>
          </div>
        )}
        {!showIndicator && !isLoading && "Everything is fine."}
      </PopoverContent>
    </Popover>
  </div>
);

export default Notification;
