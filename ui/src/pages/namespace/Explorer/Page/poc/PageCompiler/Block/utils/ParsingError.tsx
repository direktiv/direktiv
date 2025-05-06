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
      <Button variant="destructive" size="sm" icon>
        <TriangleAlert />
      </Button>
    </PopoverTrigger>
    <PopoverContent className="w-[600px] p-5 flex flex-col gap-5">
      <Alert variant="error">{title}</Alert>
      {children && <Card className="p-5 overflow-scroll">{children}</Card>}
    </PopoverContent>
  </Popover>
);
