import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { PropsWithChildren } from "react";
import { StopPropagation } from "~/components/StopPropagation";
import { TriangleAlert } from "lucide-react";
import { useTranslation } from "react-i18next";

type ParsingErrorProps = PropsWithChildren<{
  title: string;
  resetErrorBoundary?: () => void;
}>;

export const ParsingError = ({
  title,
  resetErrorBoundary,
  children,
}: ParsingErrorProps) => {
  const { t } = useTranslation();
  return (
    <Popover>
      <StopPropagation>
        <PopoverTrigger asChild>
          <Button variant="destructive" size="sm" icon aria-label={title}>
            <TriangleAlert />
          </Button>
        </PopoverTrigger>
      </StopPropagation>

      <StopPropagation>
        <PopoverContent className="flex w-[600px] flex-col gap-5 p-5">
          <Alert variant="error">
            <div className="flex">
              <span className="grow">{title}</span>

              {resetErrorBoundary && (
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={resetErrorBoundary}
                >
                  {t("direktivPage.page.error.retry")}
                </Button>
              )}
            </div>
          </Alert>
          {children && <Card className="overflow-scroll p-5">{children}</Card>}
        </PopoverContent>
      </StopPropagation>
    </Popover>
  );
};
