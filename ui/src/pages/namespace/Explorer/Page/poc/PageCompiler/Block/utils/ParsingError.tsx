import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";
import { RefreshCcw, TriangleAlert } from "lucide-react";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { Card } from "~/design/Card";
import { PropsWithChildren } from "react";

type ParsingErrorProps = PropsWithChildren<{
  title: string;
  resetErrorBoundary?: () => void;
}>;

export const ParsingError = ({
  title,
  resetErrorBoundary,
  children,
}: ParsingErrorProps) => (
  <Popover>
    <ButtonBar>
      <PopoverTrigger asChild>
        <Button variant="destructive" size="sm" icon aria-label={title}>
          <TriangleAlert />
        </Button>
      </PopoverTrigger>
      {resetErrorBoundary && (
        <Button
          variant="destructive"
          onClick={resetErrorBoundary}
          size="sm"
          icon
        >
          <RefreshCcw />
        </Button>
      )}
    </ButtonBar>
    <PopoverContent className="flex w-[600px] flex-col gap-5 p-5">
      <Alert variant="error">{title}</Alert>
      {children && <Card className="overflow-scroll p-5">{children}</Card>}
    </PopoverContent>
  </Popover>
);
