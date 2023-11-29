import { FC, PropsWithChildren } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { twMergeClsx } from "~/util/helpers";

type FilepickerPropsType = PropsWithChildren & {
  className?: string;
};

const Filepicker: FC<FilepickerPropsType> = ({ className, children }) => (
  <div className={twMergeClsx("", className)}>
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="primary" data-testid="filepicker-button">
          <div className="relative">Browse Files</div>
        </Button>
      </PopoverTrigger>
      <PopoverContent className="bg-gray-1 dark:bg-gray-dark-1" align="start">
        {children}
      </PopoverContent>
    </Popover>
  </div>
);

export { Filepicker };
