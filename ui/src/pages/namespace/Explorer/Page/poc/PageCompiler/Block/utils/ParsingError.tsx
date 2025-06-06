import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { PropsWithChildren } from "react";
import { TriangleAlert } from "lucide-react";

type ParsingErrorProps = PropsWithChildren<{
  title: string;
}>;

export const ParsingError = ({ title, children }: ParsingErrorProps) => (
  <Popover>
    <PopoverTrigger asChild>
      <Button variant="destructive" size="sm" icon aria-label={title}>
        <TriangleAlert />
      </Button>
    </PopoverTrigger>
    <PopoverContent className="flex w-[600px] flex-col gap-5 p-5">
      <Alert variant="error">{title}</Alert>
      {children && <Card className="overflow-scroll p-5">{children}</Card>}
    </PopoverContent>
  </Popover>
);
