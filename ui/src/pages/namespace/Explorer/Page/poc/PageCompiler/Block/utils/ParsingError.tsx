import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";
import { PropsWithChildren, memo, useEffect } from "react";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { Card } from "~/design/Card";
import { DirektivPagesType } from "../../../schema";
import { StopPropagation } from "~/components/StopPropagation";
import { TriangleAlert } from "lucide-react";
import { usePage } from "../../context/pageCompilerContext";

type ParsingErrorProps = PropsWithChildren<{
  title: string;
  page?: DirektivPagesType;
  resetErrorBoundary?: () => void;
}>;

/**
 * memo() with a custom compare function that always returns true will ensure that this component
 * never rerenders when the page prop changes meaning that the page prop always contains the page
 * that caused the error. The usePage hook is used to compare the current page with the page that
 * caused the error to determine if the error boundary should be reset.
 *
 * Resetting in this case means that the component tree beneath the error boundary will be rendered
 * again which could either solve the error or mount this error boundary again.
 */
export const ParsingError = memo(
  ({
    title,
    page: pageWithError,
    resetErrorBoundary,
    children,
  }: ParsingErrorProps) => {
    const pageWithErrorStringified = JSON.stringify(pageWithError);
    const currentPage = usePage();
    const currentPageStringified = JSON.stringify(currentPage);

    useEffect(() => {
      if (pageWithErrorStringified !== currentPageStringified) {
        resetErrorBoundary?.();
      }
    }, [currentPageStringified, pageWithErrorStringified, resetErrorBoundary]);

    return (
      <Popover>
        <ButtonBar>
          <StopPropagation>
            <PopoverTrigger asChild>
              <Button variant="destructive" size="sm" icon aria-label={title}>
                <TriangleAlert />
              </Button>
            </PopoverTrigger>
          </StopPropagation>
        </ButtonBar>
        <StopPropagation>
          <PopoverContent className="flex w-[600px] flex-col gap-5 p-5">
            <Alert variant="error">{title}</Alert>
            {children && (
              <Card className="overflow-scroll p-5">{children}</Card>
            )}
          </PopoverContent>
        </StopPropagation>
      </Popover>
    );
  },
  () => true
);

ParsingError.displayName = "ParsingError";
