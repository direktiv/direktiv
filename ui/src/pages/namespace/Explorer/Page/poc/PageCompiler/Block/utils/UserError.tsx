import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { PropsWithChildren } from "react";

type UserErrorProps = PropsWithChildren<{
  title: string;
}>;

export const UserError = ({ title, children }: UserErrorProps) => (
  <Alert variant="error">
    {title}
    {children && (
      <Popover>
        <PopoverTrigger asChild>
          <Button size="sm" className="block mt-2">
            show details
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[400px] p-5">
          <div className="overflow-scroll">{children}</div>
        </PopoverContent>
      </Popover>
    )}
  </Alert>
);
