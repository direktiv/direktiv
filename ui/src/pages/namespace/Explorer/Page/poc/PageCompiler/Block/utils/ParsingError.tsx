import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { Card } from "~/design/Card";
import { PropsWithChildren } from "react";
import { RefreshCcw } from "lucide-react";
import { StopPropagation } from "~/components/StopPropagation";
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
      <ButtonBar>
        <StopPropagation>
          <PopoverTrigger asChild>
            <Button variant="destructive" size="sm" aria-label={title}>
              {t("direktivPage.page.error.label")}
            </Button>
          </PopoverTrigger>
        </StopPropagation>
        {resetErrorBoundary && (
          <StopPropagation>
            <Button
              variant="destructive"
              size="sm"
              icon
              onClick={resetErrorBoundary}
            >
              <RefreshCcw />
            </Button>
          </StopPropagation>
        )}
      </ButtonBar>

      <StopPropagation>
        <PopoverContent className="flex w-[600px] flex-col gap-5 p-5">
          <Alert variant="error">{title}</Alert>
          {children && <Card className="overflow-scroll p-5">{children}</Card>}
        </PopoverContent>
      </StopPropagation>
    </Popover>
  );
};
